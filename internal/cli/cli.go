// Package cli dispatches myfactory subcommands.
//
// It intentionally uses only the standard library: the command surface is
// small and stable, and zero dependencies keeps CGO_ENABLED=0 cross-builds
// and supply-chain review trivial.
package cli

import (
	"fmt"
	"io"

	"github.com/AlharbiKhalid/myfactory-core/internal/commands"
	"github.com/AlharbiKhalid/myfactory-core/internal/version"
)

const usage = `usage: myfactory [-h] [--version] COMMAND ...

MyFactory: reusable AI software factory CLI. init sets up structure;
discovery/planning are done by AI agents using generated prompts.

commands:
  init          Set up MyFactory structure in a repo (non-interactive).
  doctor        Report MyFactory readiness of a repo.
  discover      Print the AI discovery prompt (the CLI never runs discovery itself).
  plan          Report planning readiness; print the planning prompt for AI agents.
  plane         Plane execution-tracker integration (dry-run by default).
  orchestrator  Hermes orchestrator/controller helpers.
  version       Print version, git commit, and build date.

options:
  -h, --help    show this help message and exit
  --version     show program's version number and exit

Run 'myfactory COMMAND --help' for command-specific flags.`

// Run executes the CLI and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stdout, usage)
		return 0
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Fprintln(stdout, usage)
		return 0
	case "--version":
		fmt.Fprintf(stdout, "myfactory %s\n", version.Version)
		return 0
	case "init":
		return commands.Init(args[1:], stdout, stderr)
	case "doctor":
		return commands.Doctor(args[1:], stdout, stderr)
	case "discover":
		return commands.Discover(args[1:], stdout, stderr)
	case "plan":
		return commands.Plan(args[1:], stdout, stderr)
	case "plane":
		return commands.Plane(args[1:], stdout, stderr)
	case "orchestrator":
		return commands.Orchestrator(args[1:], stdout, stderr)
	case "version":
		return commands.Version(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "myfactory: unknown command %q\n\n", args[0])
		fmt.Fprintln(stderr, usage)
		return 2
	}
}
