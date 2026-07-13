package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func capture(f func(args []string, stdout, stderr io.Writer) int, args ...string) (string, string, int) {
	var out, errOut bytes.Buffer
	code := f(args, &out, &errOut)
	return out.String(), errOut.String(), code
}

func initializedDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test", "--description", "d"); code != 0 {
		t.Fatalf("setup init failed: %s", out)
	}
	return dir
}

func TestDoctorInitializedRepo(t *testing.T) {
	dir := initializedDir(t)
	out, _, code := capture(Doctor, "--target", dir)
	if code != 0 {
		t.Fatalf("doctor exit %d", code)
	}
	for _, want := range []string{
		"[OK     ] .ApplicationFactory/product.yaml",
		"[OK     ] .ApplicationFactory/config.yaml",
		"[OK     ] docs/03-delivery/missions.yaml",
		"[OK     ] AGENTS.md (Codex overlay)",
		"[OK     ] .claude/commands/myfactory-*.md (Claude overlay)",
		"[MISSING] Git repository",
		"[INFO   ] Plane disabled in config (dry-run only)",
		"[INFO   ] Hermes orchestrator: disabled",
		"Readiness: 22/23 checks passed.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("doctor output missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorUninitializedRepo(t *testing.T) {
	out, _, code := capture(Doctor, "--target", t.TempDir())
	if code != 0 {
		t.Fatalf("doctor must not fail harshly, exit %d", code)
	}
	if !strings.Contains(out, "[MISSING] .ApplicationFactory/product.yaml") {
		t.Errorf("doctor should report missing manifest:\n%s", out)
	}
	if !strings.Contains(out, "Run `myfactory init`") {
		t.Errorf("doctor should suggest init:\n%s", out)
	}
}

func TestDiscoverPrompt(t *testing.T) {
	dir := t.TempDir()
	out, _, code := capture(Discover, "--target", dir, "--print-prompt")
	if code != 0 {
		t.Fatalf("discover exit %d", code)
	}
	for _, want := range []string{
		"# MyFactory Product Discovery Prompt",
		dir,
		"docs/00-product/idea-brief.md",
		"Git is the source of truth. Chat is not.",
		"Do NOT create implementation tasks",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("discover prompt missing %q", want)
		}
	}

	claudeOut, _, _ := capture(Discover, "--target", dir, "--agent", "claude", "--print-prompt")
	if !strings.Contains(claudeOut, "Note for Claude") {
		t.Error("claude agent note missing")
	}
	codexOut, _, _ := capture(Discover, "--target", dir, "--agent", "codex", "--print-prompt")
	if !strings.Contains(codexOut, "Note for Codex") {
		t.Error("codex agent note missing")
	}
	if _, _, code := capture(Discover, "--agent", "gemini"); code == 0 {
		t.Error("invalid agent accepted")
	}
}

func TestPlanPrompt(t *testing.T) {
	dir := t.TempDir()
	out, _, code := capture(Plan, "--target", dir, "--print-prompt")
	if code != 0 {
		t.Fatalf("plan --print-prompt exit %d", code)
	}
	for _, want := range []string{
		"# MyFactory Planning Prompt",
		"docs/03-delivery/work-breakdown.yaml",
		"MISSION-###",
		"SPRINT-###",
		"Dependencies must form a DAG",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plan prompt missing %q", want)
		}
	}
}

func TestPlanDryRunReadiness(t *testing.T) {
	dir := initializedDir(t)
	out, _, code := capture(Plan, "--target", dir, "--dry-run")
	if code != 0 {
		t.Fatalf("plan exit %d", code)
	}
	for _, want := range []string{
		"MyFactory planning readiness for:",
		"[placeholder] docs/00-product/prd.md",
		"[template   ] docs/03-delivery/missions.yaml",
		"Source docs are not fully ready.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plan readiness missing %q:\n%s", want, out)
		}
	}

	// Uninitialized: everything missing.
	emptyOut, _, _ := capture(Plan, "--target", t.TempDir(), "--dry-run")
	if !strings.Contains(emptyOut, "[missing    ] docs/00-product/prd.md") {
		t.Errorf("plan should report missing files:\n%s", emptyOut)
	}
}

func TestPlaneSyncDryRun(t *testing.T) {
	dir := initializedDir(t)
	out, _, code := capture(Plane, "sync", "--target", dir, "--dry-run")
	if code != 0 {
		t.Fatalf("plane sync exit %d:\n%s", code, out)
	}
	for _, want := range []string{
		"Plane sync plan (DRY RUN) for:",
		"Plane enabled in config: False",
		"Mapping: Mission -> Plane Module/Label | Sprint -> Plane Cycle | Task -> Plane Issue",
		"Missions (-> Plane Modules/Labels): 1 defined (1 still placeholder)",
		"would create/update: [MISSION-001] CHANGE_ME (status: draft)",
		"Sprints (-> Plane Cycles): 1 defined (1 still placeholder)",
		"would create/update: [SPRINT-001] CHANGE_ME (status: draft)",
		"Tasks (-> Plane Issues): 0 defined",
		"nothing to sync.",
		"This was a dry run. No Plane API calls were made.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plane sync missing %q:\n%s", want, out)
		}
	}
}

// writeDelivery overwrites a docs/03-delivery file in an initialized project.
func writeDelivery(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, "docs", "03-delivery", name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// PyYAML-style serialization as produced by AI planning agents: indentless
// block sequences, anchors/aliases, wrapped plain scalars, nulls, Unicode.
const pyyamlMissions = `missions:
- id: MISSION-001
  title: Repository onboarding
  goal: Connect the real repository and resolve architecture
    decisions before implementation.
  status: active
  source_docs: &id001
  - docs/00-product/prd.md
  - docs/01-business/business-rules.yaml
  sprints:
  - SPRINT-001
  - SPRINT-002
- id: MISSION-002
  title: Core implementation
  status: draft
  source_docs: *id001
  sprints:
  - SPRINT-003
`

const pyyamlSprints = `sprints:
- id: SPRINT-001
  title: Sprint one
  mission_id: MISSION-001
  status: planned
  scope:
  - TEST-T-001
  - TEST-T-002
- id: SPRINT-002
  title: Sprint two
  mission_id: MISSION-001
  status: planned
  scope: []
- id: SPRINT-003
  title: "سبرنت ثلاثة"
  mission_id: MISSION-002
  status: planned
  plane_cycle:
    id: null
`

const pyyamlWork = `work_items:
- id: TEST-T-001
  title: First task
  state: todo
- id: TEST-T-002
  title: Second task
  state: todo
  description: |
    Literal block scalar
    over two lines.
- id: TEST-T-003
  title: Third task
  state: todo
- id: TEST-T-004
  title: Fourth task
  state: todo
- id: TEST-T-005
  title: Fifth task
  state: todo
`

func TestPlaneSyncReadsAIGeneratedYAML(t *testing.T) {
	dir := initializedDir(t)
	writeDelivery(t, dir, "missions.yaml", pyyamlMissions)
	writeDelivery(t, dir, "sprints.yaml", pyyamlSprints)
	writeDelivery(t, dir, "work-breakdown.yaml", pyyamlWork)

	out, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
	if code != 0 {
		t.Fatalf("plane sync exit %d:\n%s\n%s", code, out, errOut)
	}
	for _, want := range []string{
		"Missions (-> Plane Modules/Labels): 2 defined",
		"Sprints (-> Plane Cycles): 3 defined",
		"Tasks (-> Plane Issues): 5 defined",
		"would create/update: [MISSION-002] Core implementation (status: draft)",
		"would create/update: [SPRINT-003] سبرنت ثلاثة (status: planned)",
		"would create/update: [TEST-T-005] Fifth task (status: todo)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("plane sync missing %q:\n%s", want, out)
		}
	}
}

func TestPlaneSyncFailsClosedOnMalformedYAML(t *testing.T) {
	dir := initializedDir(t)
	writeDelivery(t, dir, "missions.yaml", "missions:\n- id: [unclosed\n")

	out, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
	if code == 0 {
		t.Fatalf("plane sync must fail on malformed YAML:\n%s", out)
	}
	if !strings.Contains(errOut, "ERROR: could not parse") || !strings.Contains(errOut, "missions.yaml") {
		t.Errorf("error must name the file and reason:\n%s", errOut)
	}
	if strings.Contains(out, "Missions (-> Plane Modules/Labels): 0 defined") {
		t.Errorf("must not print a misleading zero-item plan:\n%s", out)
	}
}

func TestPlaneSyncApplyRefusesWithoutConfig(t *testing.T) {
	dir := initializedDir(t)
	out, _, code := capture(Plane, "sync", "--target", dir, "--apply")
	if code == 0 {
		t.Fatal("plane sync --apply must fail without Plane config")
	}
	for _, want := range []string{
		"Cannot apply - missing requirements:",
		"plane.enabled is false",
		"plane.base_url is not configured",
		"$PLANE_API_KEY environment variable is not set",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("apply guard missing %q:\n%s", want, out)
		}
	}
}

func TestPlaneUnknownSubcommand(t *testing.T) {
	if _, _, code := capture(Plane, "destroy"); code == 0 {
		t.Error("unknown plane subcommand accepted")
	}
}

func TestOrchestratorPromptFromProject(t *testing.T) {
	dir := initializedDir(t)
	out, _, code := capture(Orchestrator, "prompt", "--target", dir)
	if code != 0 {
		t.Fatalf("orchestrator prompt exit %d", code)
	}
	if !strings.Contains(out, "# Source: project:") {
		t.Errorf("expected project source:\n%s", out[:min(len(out), 200)])
	}
	if !strings.Contains(out, "You are Hermes") {
		t.Error("Hermes prompt body missing")
	}
	if !strings.Contains(out, "No agent approves its own work.") {
		t.Error("hard rules missing from Hermes prompt")
	}
}

func TestOrchestratorPromptEmbeddedFallback(t *testing.T) {
	out, _, code := capture(Orchestrator, "prompt", "--target", t.TempDir())
	if code != 0 {
		t.Fatalf("orchestrator fallback exit %d", code)
	}
	if !strings.Contains(out, "# Source: embedded template") {
		t.Errorf("expected embedded fallback source:\n%s", out[:min(len(out), 200)])
	}
	if !strings.Contains(out, "You are Hermes") {
		t.Error("Hermes prompt body missing in fallback")
	}
}

func TestVersionCommand(t *testing.T) {
	out, _, code := capture(Version)
	if code != 0 {
		t.Fatalf("version exit %d", code)
	}
	for _, want := range []string{"MyFactory version: dev", "Git commit:        unknown", "Build date:        unknown"} {
		if !strings.Contains(out, want) {
			t.Errorf("version output missing %q:\n%s", want, out)
		}
	}
}
