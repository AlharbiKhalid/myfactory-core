package yamlmini

import "testing"

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
	data := Parse("key: value # trailing comment\nurl: http://example.com/#anchor\n")
	if data["key"] != "value" {
		t.Errorf("trailing comment not stripped: %v", data["key"])
	}
	if data["url"] != "http://example.com/#anchor" {
		t.Errorf("mid-token # must not be treated as comment: %v", data["url"])
	}
}

func TestMultilineScalar(t *testing.T) {
	data := Parse("rule: >\n  Work items must be generated\n  from source-of-truth documents.\n")
	got, _ := data["rule"].(string)
	if got != "Work items must be generated from source-of-truth documents." {
		t.Errorf("multiline scalar = %q", got)
	}
}
