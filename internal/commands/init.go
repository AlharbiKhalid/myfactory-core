// Command init: non-interactive project setup.
//
// Adds MyFactory structure to an existing repository from the embedded
// templates. It asks no questions, calls no AI tools, contacts no external
// services, never deletes files, and never overwrites without --force.
package commands

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlharbiKhalid/myfactory-core/internal/assets"
	"github.com/AlharbiKhalid/myfactory-core/internal/fsops"
	"github.com/AlharbiKhalid/myfactory-core/internal/gitutil"
	"github.com/AlharbiKhalid/myfactory-core/internal/project"
)

var keyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,15}$`)

// DeriveKey derives a project key like CLINIC_BOOKING from a folder name.
func DeriveKey(folderName string) string {
	cleaned := regexp.MustCompile(`[^A-Za-z0-9]+`).ReplaceAllString(folderName, "_")
	cleaned = strings.ToUpper(strings.Trim(cleaned, "_"))
	if cleaned == "" {
		return "APP"
	}
	if cleaned[0] < 'A' || cleaned[0] > 'Z' {
		cleaned = "P_" + cleaned
	}
	if len(cleaned) > 16 {
		cleaned = cleaned[:16]
	}
	if !keyPattern.MatchString(cleaned) {
		return "APP"
	}
	return cleaned
}

// DeriveName derives a human-readable product name from a folder name.
func DeriveName(folderName string) string {
	words := strings.TrimSpace(regexp.MustCompile(`[-_]+`).ReplaceAllString(folderName, " "))
	if words == "" {
		return "New Product"
	}
	return titleCase(words)
}

func titleCase(s string) string {
	fields := strings.Fields(s)
	for i, w := range fields {
		fields[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(fields, " ")
}

// Init implements `myfactory init`.
func Init(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("init", flag.ContinueOnError)
	fl.SetOutput(stderr)
	target := fl.String("target", "", "Target directory (default: current directory).")
	key := fl.String("key", "", "Stable project key (default: derived from folder name).")
	name := fl.String("name", "", "Product name (default: derived from folder name).")
	description := fl.String("description", "", "Short product description.")
	gitProvider := fl.String("git-provider", "", "Git provider: github, gitlab, or none (default: auto-detect, falling back to github).")
	withCodex := fl.Bool("with-codex", true, "Include Codex overlay (AGENTS.md). Default.")
	noCodex := fl.Bool("no-codex", false, "Skip Codex overlay.")
	withClaude := fl.Bool("with-claude", true, "Include Claude overlay (.claude/commands). Default.")
	noClaude := fl.Bool("no-claude", false, "Skip Claude overlay.")
	withGitHub := fl.Bool("with-github", false, "Force-include GitHub helper files.")
	withGitLab := fl.Bool("with-gitlab", false, "Force-include GitLab helper files.")
	force := fl.Bool("force", false, "Overwrite existing files (default: never).")
	dryRun := fl.Bool("dry-run", false, "Print planned actions without changes.")
	fl.Usage = func() {
		fmt.Fprintln(stderr, "usage: myfactory init [flags]")
		fmt.Fprintln(stderr, "Non-interactive setup. Copies missing factory files into the target repo.")
		fmt.Fprintln(stderr, "Existing files are always preserved unless --force. Nothing is ever deleted.")
		fmt.Fprintln(stderr, "Discovery is done later by AI agents.")
		fl.PrintDefaults()
	}
	if err := fl.Parse(args); err != nil {
		return 2
	}
	if *gitProvider != "" && *gitProvider != "github" && *gitProvider != "gitlab" && *gitProvider != "none" {
		fmt.Fprintf(stderr, "ERROR: --git-provider must be github, gitlab, or none (got %q)\n", *gitProvider)
		return 2
	}
	codex := *withCodex && !*noCodex
	claude := *withClaude && !*noClaude

	targetDir, err := project.ResolveTarget(*target)
	if err != nil {
		fmt.Fprintf(stdout, "ERROR: %v\n", err)
		return 1
	}
	if !fsops.DirExists(targetDir) {
		fmt.Fprintf(stdout, "ERROR: target directory does not exist: %s\n", targetDir)
		return 1
	}

	projectKey := strings.ToUpper(strings.TrimSpace(*key))
	if projectKey == "" {
		projectKey = DeriveKey(filepath.Base(targetDir))
	}
	if !keyPattern.MatchString(projectKey) {
		fmt.Fprintf(stdout, "ERROR: project key must be 2-16 chars, uppercase, A-Z/0-9/_ and start with a letter. Got: %s\n", projectKey)
		return 1
	}
	productName := strings.TrimSpace(*name)
	if productName == "" {
		productName = DeriveName(filepath.Base(targetDir))
	}
	desc := strings.TrimSpace(*description)

	detected := gitutil.DetectProvider(targetDir)
	provider := *gitProvider
	if provider == "" {
		if detected != "unknown" {
			provider = detected
		} else {
			provider = "github"
		}
	}
	includeGitHub := *withGitHub || provider == "github"
	includeGitLab := *withGitLab || provider == "gitlab"

	actions := &fsops.Actions{}
	if *dryRun {
		fmt.Fprintln(stdout, "DRY RUN: no files will be changed.")
		fmt.Fprintln(stdout)
	}

	productRepo, err := assets.ProductRepo()
	if err != nil {
		fmt.Fprintf(stdout, "ERROR: %v\n", err)
		return 1
	}

	// 1. Product template (docs, .ApplicationFactory, .github when selected).
	// Excluding the ".github" directory prunes the whole subtree in CopyTree.
	// (The legacy Python CLI excluded only the files and left empty .github
	// directories behind; skipping the subtree is the intended behavior.)
	exclude := map[string]bool{}
	if !includeGitHub {
		exclude[".github"] = true
	}
	if err := fsops.CopyTree(productRepo, targetDir, actions, *force, *dryRun, exclude); err != nil {
		fmt.Fprintf(stdout, "ERROR: %v\n", err)
		return 1
	}

	// 2. Ensure metadata directories exist even if templates change.
	if err := fsops.EnsureDir(project.TaskPackagesDir(targetDir), *dryRun); err != nil {
		fmt.Fprintf(stdout, "ERROR: %v\n", err)
		return 1
	}
	if err := fsops.EnsureDir(project.OrchestratorDir(targetDir), *dryRun); err != nil {
		fmt.Fprintf(stdout, "ERROR: %v\n", err)
		return 1
	}

	// 3. GitLab helper files (template ships GitHub ones; GitLab is generated).
	if includeGitLab {
		content, err := fs.ReadFile(productRepo, ".github/pull_request_template.md")
		if err != nil {
			content = []byte("# Merge Request\n")
		}
		rel := filepath.Join(".gitlab", "merge_request_templates", "Default.md")
		if err := fsops.WriteFile(filepath.Join(targetDir, rel), rel, string(content), actions, *force, *dryRun); err != nil {
			fmt.Fprintf(stdout, "ERROR: %v\n", err)
			return 1
		}
	}

	// 4. Codex overlay: AGENTS.md only if absent (or --force).
	if codex {
		if overlay, err := assets.CodexOverlay(); err == nil {
			if err := fsops.CopyTree(overlay, targetDir, actions, *force, *dryRun, nil); err != nil {
				fmt.Fprintf(stdout, "ERROR: %v\n", err)
				return 1
			}
		}
	}

	// 5. Claude overlay: .claude/commands/*.
	if claude {
		if overlay, err := assets.ClaudeOverlay(); err == nil {
			if err := fsops.CopyTree(overlay, targetDir, actions, *force, *dryRun, nil); err != nil {
				fmt.Fprintf(stdout, "ERROR: %v\n", err)
				return 1
			}
		}
	}

	// 6. Fill placeholders where the placeholder text still exists.
	descValue := desc
	if descValue == "" {
		descValue = "CHANGE_ME"
	}
	replacements := [][2]string{
		{"key: CHANGE_ME", "key: " + fsops.YAMLString(projectKey)},
		{"name: CHANGE_ME", "name: " + fsops.YAMLString(productName)},
		{"description: CHANGE_ME", "description: " + fsops.YAMLString(descValue)},
	}
	placeholderFiles := []string{
		project.ProductManifest(targetDir),
		project.Config(targetDir),
		filepath.Join(targetDir, "docs", "03-delivery", "work-breakdown.yaml"),
	}
	// Unlike the legacy Python CLI, placeholders are only filled in files this
	// run created or force-overwrote. This makes repeat init truly idempotent:
	// a second run never consumes placeholders deeper in an existing file
	// (e.g. plane.workspace.name) that the user has chosen to keep.
	written := map[string]bool{}
	for _, rel := range actions.Created {
		written[rel] = true
	}
	for _, rel := range actions.Overwritten {
		written[rel] = true
	}
	var updated []string
	if !*dryRun {
		for _, f := range placeholderFiles {
			rel, relErr := filepath.Rel(targetDir, f)
			if relErr != nil || !written[rel] {
				continue
			}
			changed, err := fsops.ReplacePlaceholders(f, replacements, *dryRun)
			if err != nil {
				fmt.Fprintf(stdout, "ERROR: %v\n", err)
				return 1
			}
			if changed {
				updated = append(updated, rel)
			}
		}
		configRel, relErr := filepath.Rel(targetDir, project.Config(targetDir))
		if provider != "github" && relErr == nil && written[configRel] {
			_, _ = fsops.ReplacePlaceholders(
				project.Config(targetDir),
				[][2]string{{
					"provider: github        # github | gitlab | none",
					"provider: " + provider + "        # github | gitlab | none",
				}},
				*dryRun,
			)
		}
	}

	// 7. Summary.
	mode := "complete"
	if *dryRun {
		mode = "planned"
	}
	fmt.Fprintf(stdout, "MyFactory init %s for: %s\n", mode, targetDir)
	fmt.Fprintf(stdout, "Project key:  %s\n", projectKey)
	fmt.Fprintf(stdout, "Product name: %s\n", productName)
	detectedNote := ""
	if *gitProvider == "" && detected != "unknown" {
		detectedNote = " (detected)"
	}
	fmt.Fprintf(stdout, "Git provider: %s%s\n", provider, detectedNote)
	gitRepo := "no"
	if gitutil.IsGitRepo(targetDir) {
		gitRepo = "yes"
	}
	fmt.Fprintf(stdout, "Git repo:     %s\n", gitRepo)
	fmt.Fprintf(stdout, "Overlays:     codex=%s, claude=%s\n", onOff(codex), onOff(claude))
	fmt.Fprintln(stdout)
	for _, line := range actions.SummaryLines() {
		fmt.Fprintln(stdout, line)
	}
	if len(updated) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Placeholders filled in:")
		for _, f := range updated {
			fmt.Fprintf(stdout, "  %s\n", f)
		}
	}
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Next steps:")
	fmt.Fprintln(stdout, "  myfactory doctor")
	fmt.Fprintln(stdout, "  myfactory discover --print-prompt")
	fmt.Fprintln(stdout, "  myfactory plan --dry-run")
	return 0
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
