// Command orchestrator prompt: print the Hermes controller prompt.
// Prefers the project's own copy under .ApplicationFactory/orchestrator/;
// falls back to the embedded template so the command works before init.
package commands

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/AlharbiKhalid/myfactory-core/internal/assets"
	"github.com/AlharbiKhalid/myfactory-core/internal/fsops"
	"github.com/AlharbiKhalid/myfactory-core/internal/project"
)

const promptFilename = "HERMES-CONTROLLER-PROMPT.md"

// Orchestrator implements `myfactory orchestrator`.
func Orchestrator(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Fprintln(stdout, "usage: myfactory orchestrator prompt [flags]")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Hermes controls sprint execution: it delegates to agents and enforces QA gates.")
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "subcommands:")
		fmt.Fprintln(stdout, "  prompt    Print the Hermes controller prompt.")
		return 0
	}
	if args[0] != "prompt" {
		fmt.Fprintf(stderr, "ERROR: unknown orchestrator subcommand %q (expected: prompt)\n", args[0])
		return 2
	}
	return orchestratorPrompt(args[1:], stdout, stderr)
}

func orchestratorPrompt(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("orchestrator prompt", flag.ContinueOnError)
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
	// The embedded fallback exists so the command works before init, but an
	// explicitly requested target that does not exist is a user error and
	// must not silently fall back.
	if *target != "" {
		if err := project.EnsureDir(targetDir); err != nil {
			fmt.Fprintf(stderr, "ERROR: %v\n", err)
			return 1
		}
	}

	projectPrompt := filepath.Join(project.OrchestratorDir(targetDir), promptFilename)
	if fsops.FileExists(projectPrompt) {
		content, err := fsops.ReadFileString(projectPrompt)
		if err != nil {
			fmt.Fprintf(stderr, "ERROR: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "# Source: project: %s\n\n", projectPrompt)
		fmt.Fprintln(stdout, content)
		return 0
	}

	content, err := assets.ReadFile("product-repo/.ApplicationFactory/orchestrator/" + promptFilename)
	if err != nil {
		fmt.Fprintln(stdout, "ERROR: Hermes controller prompt not found in project or embedded templates.")
		return 1
	}
	fmt.Fprintf(stdout, "# Source: embedded template (project not initialized - run `myfactory init`)\n\n")
	fmt.Fprintln(stdout, string(content))
	return 0
}
