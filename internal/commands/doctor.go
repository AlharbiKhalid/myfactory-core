// Command doctor: readiness report for a MyFactory-enabled repo.
// It never fails harshly; it reports.
package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AlharbiKhalid/myfactory-core/internal/fsops"
	"github.com/AlharbiKhalid/myfactory-core/internal/gitutil"
	"github.com/AlharbiKhalid/myfactory-core/internal/project"
	"github.com/AlharbiKhalid/myfactory-core/internal/yamlmini"
)

const (
	statusOK      = "OK     "
	statusMissing = "MISSING"
	statusWarn    = "WARN   "
	statusInfo    = "INFO   "
)

// Doctor implements `myfactory doctor`.
func Doctor(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fl.SetOutput(stderr)
	target := fl.String("target", "", "Target directory (default: current directory).")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	targetDir, err := project.ResolveTarget(*target)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "MyFactory doctor report for: %s\n\n", targetDir)

	var lines []string
	okCount, total := 0, 0
	add := func(present bool, label string) bool {
		status := statusMissing
		if present {
			status = statusOK
			okCount++
		}
		total++
		lines = append(lines, fmt.Sprintf("[%s] %s", status, label))
		return present
	}

	add(fsops.DirExists(project.MetadataDir(targetDir)), ".ApplicationFactory/ metadata directory")
	add(fsops.FileExists(project.ProductManifest(targetDir)), ".ApplicationFactory/product.yaml")
	hasConfig := add(fsops.FileExists(project.Config(targetDir)), ".ApplicationFactory/config.yaml")
	add(fsops.DirExists(project.TaskPackagesDir(targetDir)), ".ApplicationFactory/task-packages/")
	add(fsops.DirExists(project.OrchestratorDir(targetDir)), ".ApplicationFactory/orchestrator/")
	add(fsops.FileExists(filepath.Join(project.OrchestratorDir(targetDir), "HERMES-CONTROLLER-PROMPT.md")),
		"Hermes controller prompt")

	for _, sub := range []string{"00-product", "01-business", "02-architecture", "03-delivery", "04-qa", "05-operations"} {
		add(fsops.DirExists(filepath.Join(targetDir, "docs", sub)), "docs/"+sub+"/")
	}

	requiredFiles := []string{
		"docs/00-product/prd.md",
		"docs/00-product/acceptance-criteria.md",
		"docs/01-business/business-rules.yaml",
		"docs/02-architecture/system-overview.md",
		"docs/03-delivery/work-breakdown.yaml",
		"docs/03-delivery/missions.yaml",
		"docs/03-delivery/sprints.yaml",
		"docs/04-qa/test-strategy.md",
	}
	for _, rel := range requiredFiles {
		add(fsops.FileExists(filepath.Join(targetDir, filepath.FromSlash(rel))), rel)
	}

	add(fsops.FileExists(filepath.Join(targetDir, "AGENTS.md")), "AGENTS.md (Codex overlay)")
	claudeCmds := filepath.Join(targetDir, ".claude", "commands")
	matches, _ := filepath.Glob(filepath.Join(claudeCmds, "myfactory-*.md"))
	add(fsops.DirExists(claudeCmds) && len(matches) > 0, ".claude/commands/myfactory-*.md (Claude overlay)")

	isRepo := add(gitutil.IsGitRepo(targetDir), "Git repository")
	if isRepo {
		lines = append(lines, fmt.Sprintf("[%s] Git provider detected: %s", statusInfo, gitutil.DetectProvider(targetDir)))
	}

	// Plane configuration.
	if hasConfig {
		config, err := yamlmini.LoadFile(project.Config(targetDir))
		if err == nil {
			if yamlmini.GetBool(config, "plane.enabled", false) {
				keyEnv := yamlmini.GetString(config, "plane.api_key_env", "PLANE_API_KEY")
				baseURL := yamlmini.GetString(config, "plane.base_url", "CHANGE_ME")
				if baseURL == "" || baseURL == "CHANGE_ME" {
					lines = append(lines, fmt.Sprintf("[%s] Plane enabled but base_url is not configured", statusWarn))
				}
				if os.Getenv(keyEnv) != "" {
					lines = append(lines, fmt.Sprintf("[%s] Plane API key present in $%s", statusOK, keyEnv))
				} else {
					lines = append(lines, fmt.Sprintf("[%s] Plane enabled but $%s is not set", statusWarn, keyEnv))
				}
			} else {
				lines = append(lines, fmt.Sprintf("[%s] Plane disabled in config (dry-run only)", statusInfo))
			}
			hermes := "disabled"
			if yamlmini.GetBool(config, "orchestration.hermes.enabled", false) {
				hermes = "enabled"
			}
			lines = append(lines, fmt.Sprintf("[%s] Hermes orchestrator: %s", statusInfo, hermes))
		}
	}

	for _, line := range lines {
		fmt.Fprintln(stdout, line)
	}
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Readiness: %d/%d checks passed.\n", okCount, total)
	if okCount < total {
		fmt.Fprintln(stdout, "Run `myfactory init` to add missing structure (existing files are preserved).")
	} else {
		fmt.Fprintln(stdout, "Structure looks ready. Next: myfactory discover --print-prompt")
	}
	return 0
}
