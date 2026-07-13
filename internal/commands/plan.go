// Command plan: planning readiness report and planning prompt.
// The CLI does not generate plans with AI; it reports readiness and prints
// the prompt an AI agent uses to populate the delivery files.
package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlharbiKhalid/myfactory-core/internal/project"
	"github.com/AlharbiKhalid/myfactory-core/internal/yamlmini"
)

const placeholder = "CHANGE_ME"

type readinessSection struct {
	name  string
	files []string
}

var readinessSections = []readinessSection{
	{"product", []string{
		"docs/00-product/idea-brief.md",
		"docs/00-product/prd.md",
		"docs/00-product/user-journeys.md",
		"docs/00-product/acceptance-criteria.md",
	}},
	{"business", []string{
		"docs/01-business/business-rules.yaml",
		"docs/01-business/decision-tables.md",
	}},
	{"architecture", []string{
		"docs/02-architecture/system-overview.md",
		"docs/02-architecture/api-contracts.md",
	}},
	{"qa", []string{
		"docs/04-qa/test-strategy.md",
	}},
}

const planningPrompt = `# MyFactory Planning Prompt

You are an AI planning agent inside a MyFactory-enabled repository at: {target}

Git is the source of truth. Plane is only the execution tracker.
Mapping: MyFactory Mission = larger goal. MyFactory Sprint = Plane Cycle.
MyFactory Task = Plane Issue.

## Read first

1. ` + "`.ApplicationFactory/product.yaml` and `.ApplicationFactory/config.yaml`" + `
2. ` + "`docs/00-product/prd.md` and `docs/00-product/acceptance-criteria.md`" + `
3. ` + "`docs/01-business/business-rules.yaml`" + `
4. ` + "`docs/02-architecture/system-overview.md` and `api-contracts.md`" + `
5. ` + "`docs/04-qa/test-strategy.md`" + `
6. Existing ` + "`docs/03-delivery/`" + ` files - extend them; never renumber IDs.

If product/business/architecture docs are still placeholders, stop and tell the
user to run discovery first.

## Produce

1. ` + "`docs/03-delivery/work-breakdown.yaml`" + ` - work items following the file's
   ` + "`work_item_schema`" + ` (id, type, title, module, priority, state, source_docs,
   acceptance_criteria, definition_of_done, dependencies). Use the task ID
   convention with the project key. Reference BR-* business rule IDs where
   business logic is touched. Dependencies must form a DAG.
2. ` + "`docs/03-delivery/missions.yaml`" + ` - MISSION-### entries with goal, status,
   source_docs, success_criteria, and the sprints that deliver them.
3. ` + "`docs/03-delivery/sprints.yaml`" + ` - SPRINT-### entries with mission_id, scope
   (work item IDs), entry/exit criteria, and validation_required (functional QA
   always; business QA when BR-* rules are in scope).

## Rules

- Every work item must trace to at least one source doc.
- Size items so one agent finishes one item in one session.
- Respect dependencies across sprints.
- Do not implement anything. Do not modify product/business/architecture docs.
- Record assumptions and open questions in the delivery files as comments.

## Finish

Summarize the plan and offer to commit with:
` + "`docs(delivery): populate work breakdown, missions, and sprints`" + `
Then the user runs: myfactory plane sync --dry-run
`

// Plan implements `myfactory plan`.
func Plan(args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("plan", flag.ContinueOnError)
	fl.SetOutput(stderr)
	target := fl.String("target", "", "Target directory (default: current directory).")
	_ = fl.Bool("dry-run", false, "Report readiness without changing anything (default).")
	printPrompt := fl.Bool("print-prompt", false, "Print the planning prompt for Claude/Codex.")
	if err := fl.Parse(args); err != nil {
		return 2
	}
	targetDir, err := project.ResolveTarget(*target)
	if err != nil {
		fmt.Fprintf(stderr, "ERROR: %v\n", err)
		return 1
	}

	if *printPrompt {
		fmt.Fprintln(stdout, strings.ReplaceAll(planningPrompt, "{target}", targetDir))
		return 0
	}

	fmt.Fprintf(stdout, "MyFactory planning readiness for: %s\n\n", targetDir)
	allReady := true
	stateLabels := map[string]string{
		"filled":      "ready      ",
		"placeholder": "placeholder",
		"missing":     "missing    ",
		"unreadable":  "unreadable ",
	}
	for _, section := range readinessSections {
		fmt.Fprintf(stdout, "%s:\n", section.name)
		for _, rel := range section.files {
			state := fileState(filepath.Join(targetDir, filepath.FromSlash(rel)))
			fmt.Fprintf(stdout, "  [%s] %s\n", stateLabels[state], rel)
			if state != "filled" {
				allReady = false
			}
		}
		fmt.Fprintln(stdout)
	}

	delivery := []struct {
		rel     string
		listKey string
	}{
		{"docs/03-delivery/work-breakdown.yaml", "work_items"},
		{"docs/03-delivery/missions.yaml", "missions"},
		{"docs/03-delivery/sprints.yaml", "sprints"},
	}
	fmt.Fprintln(stdout, "delivery (planning output):")
	for _, d := range delivery {
		path := filepath.Join(targetDir, filepath.FromSlash(d.rel))
		label := "missing    "
		if _, err := os.Stat(path); err == nil {
			items, err := realItems(path, d.listKey)
			switch {
			case err != nil:
				label = "unreadable "
				fmt.Fprintf(stderr, "WARNING: %v\n", err)
			case len(items) > 0:
				label = "populated  "
			default:
				label = "template   "
			}
		}
		fmt.Fprintf(stdout, "  [%s] %s\n", label, d.rel)
	}
	fmt.Fprintln(stdout)

	if allReady {
		fmt.Fprintln(stdout, "Source docs are ready for planning.")
		fmt.Fprintln(stdout, "Next: myfactory plan --print-prompt  (paste the prompt into Claude/Codex)")
	} else {
		fmt.Fprintln(stdout, "Source docs are not fully ready. Missing/placeholder files above need")
		fmt.Fprintln(stdout, "discovery first: myfactory discover --print-prompt")
	}
	return 0
}

// realItems returns non-placeholder entries of a delivery file's main list.
func realItems(path, listKey string) ([]map[string]any, error) {
	data, err := yamlmini.LoadFile(path)
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for _, item := range yamlmini.Items(data, listKey) {
		title, _ := item["title"].(string)
		if title != "" && title != placeholder {
			out = append(out, item)
		}
	}
	return out, nil
}

func fileState(path string) string {
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "missing"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "unreadable"
	}
	if strings.Contains(string(data), placeholder) {
		return "placeholder"
	}
	return "filled"
}
