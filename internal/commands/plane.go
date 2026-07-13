// Command plane sync: dry-run synchronization plan for Plane.
//
// Mapping:
//   - MyFactory Mission -> Plane Module (or label; future mapping configurable)
//   - MyFactory Sprint  -> Plane Cycle
//   - MyFactory Task    -> Plane Issue / Work Item
//
// Default is always dry-run. Live sync requires --apply AND plane.enabled:
// true AND the configured API key environment variable. Live API calls are
// not implemented in this milestone; --apply explains what is missing instead.
package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AlharbiKhalid/myfactory-core/internal/project"
	"github.com/AlharbiKhalid/myfactory-core/internal/yamlmini"
)

// Plane implements `myfactory plane`.
func Plane(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(stdout, "usage: myfactory plane sync [flags]")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Plane integration. Plane is the execution tracker; Git remains the source of truth.")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "subcommands:")
		fmt.Fprintln(stdout, "  sync    Show what would be created/updated in Plane (dry-run by default).")
		return 0
	}
	if args[0] != "sync" {
		fmt.Fprintf(stderr, "ERROR: unknown plane subcommand %q (expected: sync)\n", args[0])
		return 2
	}
	return planeSync(args[1:], stdout, stderr)
}

func planeSync(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("plane sync", flag.ContinueOnError)
	fl.SetOutput(stderr)
	target := fl.String("target", "", "Target directory (default: current directory).")
	_ = fl.Bool("dry-run", false, "Print the sync plan without calling Plane (default).")
	apply := fl.Bool("apply", false, "Attempt live sync (requires Plane config + API key).")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	targetDir, err := project.ResolveTarget(*target)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 1
	}

	if err := project.EnsureInitialized(targetDir); err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 1
	}

	config, err := yamlmini.LoadFile(project.Config(targetDir))
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 1
	}
	// Fail closed: a delivery file that is missing, unparseable, empty, or
	// missing its expected top-level list must surface as an error, never as
	// a zero-item sync plan. Zero items are only valid when the file exists
	// and explicitly defines the list (e.g. `work_items: []`).
	delivery := filepath.Join(targetDir, "docs", "03-delivery")
	loadDelivery := func(name, key string) ([]map[string]any, bool) {
		path := filepath.Join(delivery, name)
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			fmt.Fprintf(stderr, "ERROR: required delivery file is missing: %s\n", path)
			fmt.Fprintln(stderr, "Run `myfactory init` to restore templates, or `myfactory plan --print-prompt` to regenerate the plan.")
			return nil, false
		}
		data, err := yamlmini.LoadFile(path)
		if err != nil {
			fmt.Fprintf(stderr, "ERROR: %v\n", err)
			return nil, false
		}
		value, present := data[key]
		if !present {
			fmt.Fprintf(stderr, "ERROR: %s is empty or has the wrong structure: missing top-level %q key\n", path, key)
			return nil, false
		}
		switch value.(type) {
		case nil, []any:
		default:
			fmt.Fprintf(stderr, "ERROR: %s has the wrong structure: top-level %q must be a list\n", path, key)
			return nil, false
		}
		return yamlmini.Items(data, key), true
	}
	missions, ok := loadDelivery("missions.yaml", "missions")
	if !ok {
		return 1
	}
	sprints, ok := loadDelivery("sprints.yaml", "sprints")
	if !ok {
		return 1
	}
	work, ok := loadDelivery("work-breakdown.yaml", "work_items")
	if !ok {
		return 1
	}

	enabled := yamlmini.GetBool(config, "plane.enabled", false)
	keyEnv := yamlmini.GetString(config, "plane.api_key_env", "PLANE_API_KEY")
	hasKey := os.Getenv(keyEnv) != ""
	baseURL := yamlmini.GetString(config, "plane.base_url", "CHANGE_ME")

	mode := "DRY RUN"
	if *apply {
		mode = "APPLY"
	}
	fmt.Fprintf(stdout, "Plane sync plan (%s) for: %s\n", mode, targetDir)
	fmt.Fprintf(stdout, "Plane enabled in config: %s\n", pyBool(enabled))
	fmt.Fprintf(stdout, "Plane base_url: %s\n", baseURL)
	fmt.Fprintf(stdout, "API key ($%s) present: %s\n", keyEnv, pyBool(hasKey))
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Mapping: Mission -> Plane Module/Label | Sprint -> Plane Cycle | Task -> Plane Issue")
	fmt.Fprintln(stdout)

	describe := func(kind string, items []map[string]any) {
		placeholders := 0
		for _, item := range items {
			title, _ := item["title"].(string)
			if title == "" || title == "CHANGE_ME" {
				placeholders++
			}
		}
		suffix := ""
		if placeholders > 0 {
			suffix = fmt.Sprintf(" (%d still placeholder)", placeholders)
		}
		fmt.Fprintf(stdout, "%s: %d defined%s\n", kind, len(items), suffix)
		for _, item := range items {
			id := stringOr(item["id"], "?")
			title := stringOr(item["title"], "?")
			status := stringOr(item["status"], "")
			if status == "" {
				status = stringOr(item["state"], "-")
			}
			fmt.Fprintf(stdout, "  would create/update: [%s] %s (status: %s)\n", id, title, status)
		}
		if len(items) == 0 {
			fmt.Fprintln(stdout, "  nothing to sync.")
		}
		fmt.Fprintln(stdout)
	}

	describe("Missions (-> Plane Modules/Labels)", missions)
	describe("Sprints (-> Plane Cycles)", sprints)
	describe("Tasks (-> Plane Issues)", work)

	fmt.Fprintln(stdout, "Sync rules: create_missing=true, update_existing=true, delete_from_plane=false (never deletes).")

	if *apply {
		var missing []string
		if !enabled {
			missing = append(missing, "plane.enabled is false in .ApplicationFactory/config.yaml")
		}
		if baseURL == "" || baseURL == "CHANGE_ME" {
			missing = append(missing, "plane.base_url is not configured")
		}
		if !hasKey {
			missing = append(missing, fmt.Sprintf("$%s environment variable is not set", keyEnv))
		}
		if len(missing) > 0 {
			fmt.Fprintln(stdout, "\nCannot apply - missing requirements:")
			for _, m := range missing {
				fmt.Fprintf(stdout, "  - %s\n", m)
			}
			return 1
		}
		fmt.Fprintln(stdout, "\nLive Plane sync is not implemented in this milestone.")
		fmt.Fprintln(stdout, "The dry-run plan above is what a live sync would perform.")
		return 1
	}

	fmt.Fprintln(stdout, "\nThis was a dry run. No Plane API calls were made.")
	return 0
}

// pyBool matches the legacy Python CLI's True/False rendering.
func pyBool(v bool) string {
	if v {
		return "True"
	}
	return "False"
}

func stringOr(v any, def string) string {
	switch t := v.(type) {
	case string:
		if t == "" {
			return def
		}
		return t
	case nil:
		return def
	default:
		return fmt.Sprintf("%v", t)
	}
}
