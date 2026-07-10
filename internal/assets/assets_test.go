package assets

import (
	"io/fs"
	"testing"
)

// TestHiddenAssetsAreEmbedded proves that go:embed's `all:` prefix captured
// the hidden template files. Plain directory embedding silently omits paths
// beginning with "." or "_", which would break `myfactory init`.
func TestHiddenAssetsAreEmbedded(t *testing.T) {
	t.Setenv(EnvOverride, "") // force embedded assets, not a dev override

	required := []string{
		"product-repo/.ApplicationFactory/product.yaml",
		"product-repo/.ApplicationFactory/config.yaml",
		"product-repo/.ApplicationFactory/orchestrator/HERMES-CONTROLLER-PROMPT.md",
		"product-repo/.ApplicationFactory/orchestrator/SPRINT-RUN-LOOP.md",
		"product-repo/.ApplicationFactory/orchestrator/SERVER-REGISTRY-TEMPLATE.yaml",
		"product-repo/.ApplicationFactory/orchestrator/RUNTIME-STATE.yaml",
		"product-repo/.ApplicationFactory/task-packages/TASK-PACKAGE-TEMPLATE.md",
		"product-repo/.github/pull_request_template.md",
		"product-repo/.github/workflows/myfactory-checks.yml",
		"product-repo/docs/00-product/idea-brief.md",
		"product-repo/docs/03-delivery/missions.yaml",
		"product-repo/docs/03-delivery/sprints.yaml",
		"product-repo/docs/03-delivery/work-breakdown.yaml",
		"project-overlays/codex/AGENTS.md",
		"project-overlays/claude/.claude/commands/myfactory-discover.md",
		"project-overlays/claude/.claude/commands/myfactory-business-rules.md",
		"project-overlays/claude/.claude/commands/myfactory-architecture.md",
		"project-overlays/claude/.claude/commands/myfactory-work-breakdown.md",
		"project-overlays/claude/.claude/commands/myfactory-plan-sprints.md",
		"project-overlays/claude/.claude/commands/myfactory-run-sprint.md",
		"project-overlays/claude/.claude/commands/myfactory-product.md",
	}
	base, err := Base()
	if err != nil {
		t.Fatalf("Base(): %v", err)
	}
	for _, rel := range required {
		data, err := fs.ReadFile(base, rel)
		if err != nil {
			t.Errorf("embedded asset missing: %s (%v)", rel, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("embedded asset is empty: %s", rel)
		}
	}
}

// TestSubFilesystems ensures the overlay/product accessors expose the hidden
// directories through fs.Sub as well.
func TestSubFilesystems(t *testing.T) {
	t.Setenv(EnvOverride, "")

	productRepo, err := ProductRepo()
	if err != nil {
		t.Fatalf("ProductRepo(): %v", err)
	}
	if _, err := fs.ReadFile(productRepo, ".ApplicationFactory/product.yaml"); err != nil {
		t.Errorf("product repo sub-FS missing .ApplicationFactory/product.yaml: %v", err)
	}

	claude, err := ClaudeOverlay()
	if err != nil {
		t.Fatalf("ClaudeOverlay(): %v", err)
	}
	entries, err := fs.ReadDir(claude, ".claude/commands")
	if err != nil {
		t.Fatalf("claude overlay missing .claude/commands: %v", err)
	}
	if len(entries) < 7 {
		t.Errorf("expected at least 7 claude commands, got %d", len(entries))
	}

	codex, err := CodexOverlay()
	if err != nil {
		t.Fatalf("CodexOverlay(): %v", err)
	}
	if _, err := fs.ReadFile(codex, "AGENTS.md"); err != nil {
		t.Errorf("codex overlay missing AGENTS.md: %v", err)
	}
}

// TestAssetsDirOverride checks the development override path.
func TestAssetsDirOverride(t *testing.T) {
	t.Setenv(EnvOverride, t.TempDir())
	base, err := Base()
	if err != nil {
		t.Fatalf("Base() with override: %v", err)
	}
	if _, err := fs.ReadFile(base, "product-repo/.ApplicationFactory/product.yaml"); err == nil {
		t.Error("override dir is empty; read should fail, proving the override is in effect")
	}

	t.Setenv(EnvOverride, `Z:\definitely\not\a\real\dir`)
	if _, err := Base(); err == nil {
		t.Error("expected error for nonexistent override directory")
	}
}
