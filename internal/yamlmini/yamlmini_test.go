package yamlmini

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleConfig = `# comment
project:
  key: "TEST"
  name: "Init Test"

plane:
  enabled: false
  base_url: CHANGE_ME
  api_key_env: PLANE_API_KEY
  workspace:
    id: CHANGE_ME
    name: CHANGE_ME

orchestration:
  hermes:
    enabled: true
    max_parallel_tasks: 3
`

func TestGetPath(t *testing.T) {
	data := Parse(sampleConfig)
	if got := GetString(data, "project.key", ""); got != "TEST" {
		t.Errorf("project.key = %q, want TEST", got)
	}
	if GetBool(data, "plane.enabled", true) {
		t.Error("plane.enabled should parse as false")
	}
	if !GetBool(data, "orchestration.hermes.enabled", false) {
		t.Error("orchestration.hermes.enabled should parse as true")
	}
	if got := GetPath(data, "orchestration.hermes.max_parallel_tasks", 0); got != 3 {
		t.Errorf("max_parallel_tasks = %v, want 3", got)
	}
	if got := GetString(data, "plane.api_key_env", ""); got != "PLANE_API_KEY" {
		t.Errorf("api_key_env = %q", got)
	}
	if got := GetString(data, "missing.path", "fallback"); got != "fallback" {
		t.Errorf("missing path = %q, want fallback", got)
	}
}

const sampleMissions = `schema:
  statuses:
    - draft
    - done

missions:
  - id: MISSION-001
    title: CHANGE_ME
    goal: CHANGE_ME
    status: draft
    source_docs:
      - docs/00-product/prd.md
    success_criteria: []
    sprints:
      - SPRINT-001
    risks: []
  - id: MISSION-002
    title: "Real mission"
    status: active
`

func TestListsOfMappings(t *testing.T) {
	data := Parse(sampleMissions)
	missions := Items(data, "missions")
	if len(missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(missions))
	}
	if missions[0]["id"] != "MISSION-001" {
		t.Errorf("first mission id = %v", missions[0]["id"])
	}
	if missions[0]["status"] != "draft" {
		t.Errorf("first mission status = %v", missions[0]["status"])
	}
	docs, ok := missions[0]["source_docs"].([]any)
	if !ok || len(docs) != 1 || docs[0] != "docs/00-product/prd.md" {
		t.Errorf("source_docs = %v", missions[0]["source_docs"])
	}
	if empty, ok := missions[0]["risks"].([]any); !ok || len(empty) != 0 {
		t.Errorf("risks should be an empty list, got %v", missions[0]["risks"])
	}
	if missions[1]["title"] != "Real mission" {
		t.Errorf("quoted title = %v", missions[1]["title"])
	}
}

func TestCommentsAndEmpty(t *testing.T) {
	if got := Parse(""); len(got) != 0 {
		t.Errorf("empty input should give empty map, got %v", got)
	}
	if got := Parse("# only a comment\n"); len(got) != 0 {
		t.Errorf("comment-only input should give empty map, got %v", got)
	}
	data := Parse("key: value # trailing comment\nurl: http://example.com/#anchor\n")
	if data["key"] != "value" {
		t.Errorf("trailing comment not stripped: %v", data["key"])
	}
	if data["url"] != "http://example.com/#anchor" {
		t.Errorf("mid-token # must not be treated as comment: %v", data["url"])
	}
}

func TestDocumentMarker(t *testing.T) {
	data := Parse("---\nkey: value\n")
	if data["key"] != "value" {
		t.Errorf("document start marker not accepted: %v", data)
	}
}

// PyYAML and AI agents emit block sequences at the same indentation as their
// parent key ("indentless"), which the legacy parser could not read.
func TestIndentlessSequences(t *testing.T) {
	data := Parse(`missions:
- id: MISSION-001
  title: First
  sprints:
  - SPRINT-001
  - SPRINT-002
- id: MISSION-002
  title: Second
`)
	missions := Items(data, "missions")
	if len(missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(missions))
	}
	if missions[0]["title"] != "First" || missions[1]["title"] != "Second" {
		t.Errorf("titles = %v, %v", missions[0]["title"], missions[1]["title"])
	}
	sprints, ok := missions[0]["sprints"].([]any)
	if !ok || len(sprints) != 2 || sprints[0] != "SPRINT-001" || sprints[1] != "SPRINT-002" {
		t.Errorf("nested indentless sequence = %v", missions[0]["sprints"])
	}
}

func TestAnchorsAndAliases(t *testing.T) {
	data := Parse(`missions:
- id: MISSION-001
  source_docs: &common
  - docs/a.md
  - docs/b.md
- id: MISSION-002
  source_docs: *common
`)
	missions := Items(data, "missions")
	if len(missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(missions))
	}
	for i, m := range missions {
		docs, ok := m["source_docs"].([]any)
		if !ok || len(docs) != 2 || docs[0] != "docs/a.md" || docs[1] != "docs/b.md" {
			t.Errorf("mission %d source_docs = %v", i, m["source_docs"])
		}
	}
}

func TestWrappedPlainScalar(t *testing.T) {
	data := Parse(`goal: Resolve repository and architecture
  decisions before implementation begins.
`)
	want := "Resolve repository and architecture decisions before implementation begins."
	if data["goal"] != want {
		t.Errorf("wrapped plain scalar = %q, want %q", data["goal"], want)
	}
}

func TestLiteralAndFoldedScalars(t *testing.T) {
	data := Parse("literal: |\n  line one\n  line two\nfolded: >\n  joined into\n  one line\nfolded_stripped: >-\n  no trailing\n  newline\n")
	if data["literal"] != "line one\nline two\n" {
		t.Errorf("literal scalar = %q", data["literal"])
	}
	if data["folded"] != "joined into one line\n" {
		t.Errorf("folded scalar = %q", data["folded"])
	}
	if data["folded_stripped"] != "no trailing newline" {
		t.Errorf("stripped folded scalar = %q", data["folded_stripped"])
	}
}

func TestUnicodeContent(t *testing.T) {
	data := Parse("title: نظام إدارة المصنع\nnote: \"عربي and English mixed\"\n")
	if data["title"] != "نظام إدارة المصنع" {
		t.Errorf("Arabic plain scalar = %q", data["title"])
	}
	if data["note"] != "عربي and English mixed" {
		t.Errorf("Arabic quoted scalar = %q", data["note"])
	}
}

func TestNullAndEmptyValues(t *testing.T) {
	data := Parse(`plane_cycle:
  id: null
  name: ~
  pending:
empty_list: []
empty_map: {}
`)
	cycle, ok := data["plane_cycle"].(map[string]any)
	if !ok {
		t.Fatalf("plane_cycle = %v", data["plane_cycle"])
	}
	for _, key := range []string{"id", "name", "pending"} {
		if v, present := cycle[key]; !present || v != nil {
			t.Errorf("plane_cycle.%s = %v, want nil", key, v)
		}
	}
	if l, ok := data["empty_list"].([]any); !ok || len(l) != 0 {
		t.Errorf("empty_list = %v", data["empty_list"])
	}
	if m, ok := data["empty_map"].(map[string]any); !ok || len(m) != 0 {
		t.Errorf("empty_map = %v", data["empty_map"])
	}
}

func TestScalarTypes(t *testing.T) {
	data := Parse("s: plain\nq: 'single'\nd: \"double\"\nb1: true\nb2: false\ni: 42\nf: 3.5\nn: null\n")
	if data["s"] != "plain" || data["q"] != "single" || data["d"] != "double" {
		t.Errorf("strings = %v %v %v", data["s"], data["q"], data["d"])
	}
	if data["b1"] != true || data["b2"] != false {
		t.Errorf("bools = %v %v", data["b1"], data["b2"])
	}
	if data["i"] != 42 {
		t.Errorf("int = %v (%T)", data["i"], data["i"])
	}
	if data["f"] != 3.5 {
		t.Errorf("float = %v (%T)", data["f"], data["f"])
	}
	if v, present := data["n"]; !present || v != nil {
		t.Errorf("null = %v", v)
	}
}

func TestMultilineScalar(t *testing.T) {
	data := Parse("rule: >\n  Work items must be generated\n  from source-of-truth documents.\n")
	got, _ := data["rule"].(string)
	// Standard YAML clip chomping keeps the single trailing newline.
	if got != "Work items must be generated from source-of-truth documents.\n" {
		t.Errorf("multiline scalar = %q", got)
	}
}

func TestInvalidYAML(t *testing.T) {
	invalid := "missions:\n- id: [unclosed\n"
	if _, err := ParseWithError(invalid); err == nil {
		t.Error("ParseWithError should reject invalid YAML")
	}
	if got := Parse(invalid); len(got) != 0 {
		t.Errorf("Parse of invalid YAML should give empty map, got %v", got)
	}
	if _, err := ParseWithError("- a\n- b\n"); err == nil {
		t.Error("ParseWithError should reject non-mapping top level")
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()

	missing, err := LoadFile(filepath.Join(dir, "nope.yaml"))
	if err != nil || len(missing) != 0 {
		t.Errorf("missing file should give empty map, got %v, %v", missing, err)
	}

	good := filepath.Join(dir, "good.yaml")
	if err := os.WriteFile(good, []byte("key: value\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := LoadFile(good)
	if err != nil || data["key"] != "value" {
		t.Errorf("LoadFile(good) = %v, %v", data, err)
	}

	bad := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(bad, []byte("key: [unclosed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFile(bad); err == nil {
		t.Error("LoadFile must return an error for invalid YAML")
	} else if !strings.Contains(err.Error(), "could not parse") || !strings.Contains(err.Error(), bad) {
		t.Errorf("LoadFile error should name the file: %v", err)
	}
}

// All YAML shipped in the repo's templates must remain parseable.
func TestTemplatesStillParse(t *testing.T) {
	root := filepath.Join("..", "..", "templates")
	found := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}
		found++
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := ParseWithError(string(text)); err != nil {
			t.Errorf("template %s no longer parses: %v", path, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking templates: %v", err)
	}
	if found == 0 {
		t.Fatal("no template YAML files found; wrong path?")
	}
}
