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

func TestPlaneSyncNonexistentTarget(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "definitely-does-not-exist")
	out, errOut, code := capture(Plane, "sync", "--target", bad, "--dry-run")
	if code == 0 {
		t.Fatalf("plane sync must fail for a nonexistent target:\n%s", out)
	}
	if !strings.Contains(errOut, "ERROR: target directory does not exist: "+bad) {
		t.Errorf("expected target-directory error naming %s:\n%s", bad, errOut)
	}
	if strings.Contains(out, "defined") {
		t.Errorf("must not print a sync plan for a nonexistent target:\n%s", out)
	}
}

func TestPlaneSyncUninitializedTarget(t *testing.T) {
	dir := t.TempDir()
	out, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
	if code == 0 {
		t.Fatalf("plane sync must fail for an uninitialized target:\n%s", out)
	}
	if !strings.Contains(errOut, "not a MyFactory project") || !strings.Contains(errOut, "myfactory init") {
		t.Errorf("error must point at `myfactory init`:\n%s", errOut)
	}
	if strings.Contains(out, "defined") {
		t.Errorf("must not print a sync plan for an uninitialized target:\n%s", out)
	}
}

func TestPlaneSyncMissingDeliveryFiles(t *testing.T) {
	for _, name := range []string{"missions.yaml", "sprints.yaml", "work-breakdown.yaml"} {
		t.Run(name, func(t *testing.T) {
			dir := initializedDir(t)
			path := filepath.Join(dir, "docs", "03-delivery", name)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			out, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
			if code == 0 {
				t.Fatalf("plane sync must fail when %s is missing:\n%s", name, out)
			}
			if !strings.Contains(errOut, "required delivery file is missing: "+path) {
				t.Errorf("error must name the missing path %s:\n%s", path, errOut)
			}
			if strings.Contains(out, "defined") {
				t.Errorf("must not print a sync plan when %s is missing:\n%s", name, out)
			}
		})
	}
}

func TestPlaneSyncRejectsEmptyOrWrongStructure(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		dir := initializedDir(t)
		writeDelivery(t, dir, "missions.yaml", "")
		out, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
		if code == 0 {
			t.Fatalf("plane sync must fail for an empty missions.yaml:\n%s", out)
		}
		if !strings.Contains(errOut, "missions.yaml") || !strings.Contains(errOut, `missing top-level "missions" key`) {
			t.Errorf("error must explain the missing top-level key:\n%s", errOut)
		}
	})
	t.Run("missing top-level key", func(t *testing.T) {
		dir := initializedDir(t)
		writeDelivery(t, dir, "sprints.yaml", "schema: {}\n")
		_, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
		if code == 0 {
			t.Fatal("plane sync must fail when the sprints key is absent")
		}
		if !strings.Contains(errOut, `missing top-level "sprints" key`) {
			t.Errorf("error must name the sprints key:\n%s", errOut)
		}
	})
	t.Run("non-list value", func(t *testing.T) {
		dir := initializedDir(t)
		writeDelivery(t, dir, "work-breakdown.yaml", "work_items: 5\n")
		_, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
		if code == 0 {
			t.Fatal("plane sync must fail when work_items is not a list")
		}
		if !strings.Contains(errOut, `top-level "work_items" must be a list`) {
			t.Errorf("error must say work_items must be a list:\n%s", errOut)
		}
	})
}

// Files that exist, parse, and explicitly define their top-level lists may
// report zero items normally.
func TestPlaneSyncExplicitlyEmptyPlan(t *testing.T) {
	dir := initializedDir(t)
	writeDelivery(t, dir, "missions.yaml", "missions: []\n")
	writeDelivery(t, dir, "sprints.yaml", "sprints: []\n")
	writeDelivery(t, dir, "work-breakdown.yaml", "work_items: []\n")
	out, errOut, code := capture(Plane, "sync", "--target", dir, "--dry-run")
	if code != 0 {
		t.Fatalf("plane sync must accept an explicitly empty plan, exit %d:\n%s", code, errOut)
	}
	for _, want := range []string{
		"Missions (-> Plane Modules/Labels): 0 defined",
		"Sprints (-> Plane Cycles): 0 defined",
		"Tasks (-> Plane Issues): 0 defined",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("empty plan missing %q:\n%s", want, out)
		}
	}
}

func TestDoctorNonexistentTarget(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "definitely-does-not-exist")
	out, errOut, code := capture(Doctor, "--target", bad)
	if code == 0 {
		t.Fatalf("doctor must fail for a nonexistent target:\n%s", out)
	}
	if !strings.Contains(errOut, "ERROR: target directory does not exist: "+bad) {
		t.Errorf("expected target-directory error:\n%s", errOut)
	}
	if strings.Contains(out, "Readiness:") {
		t.Errorf("must not print a readiness score for a nonexistent target:\n%s", out)
	}
}

func TestPlanNonexistentTarget(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "definitely-does-not-exist")
	out, errOut, code := capture(Plan, "--target", bad, "--dry-run")
	if code == 0 {
		t.Fatalf("plan must fail for a nonexistent target:\n%s", out)
	}
	if !strings.Contains(errOut, "ERROR: target directory does not exist: "+bad) {
		t.Errorf("expected target-directory error:\n%s", errOut)
	}
}

func TestPlanUninitializedTargetNote(t *testing.T) {
	out, _, code := capture(Plan, "--target", t.TempDir(), "--dry-run")
	if code != 0 {
		t.Fatalf("plan remains a report for existing uninitialized dirs, exit %d", code)
	}
	if !strings.Contains(out, "not MyFactory-initialized") || !strings.Contains(out, "myfactory init") {
		t.Errorf("plan must call out an uninitialized project:\n%s", out)
	}
}

func TestOrchestratorPromptNonexistentExplicitTarget(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "definitely-does-not-exist")
	out, errOut, code := capture(Orchestrator, "prompt", "--target", bad)
	if code == 0 {
		t.Fatalf("orchestrator prompt must not fall back for an explicit nonexistent target:\n%s", out)
	}
	if !strings.Contains(errOut, "ERROR: target directory does not exist: "+bad) {
		t.Errorf("expected target-directory error:\n%s", errOut)
	}
	if strings.Contains(out, "embedded template") {
		t.Errorf("must not silently use the embedded fallback:\n%s", out)
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
