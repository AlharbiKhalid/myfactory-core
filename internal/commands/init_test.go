package commands

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// expectedHiddenAssets are files `myfactory init` must create, all under
// hidden directories that naive embedding would drop.
var expectedHiddenAssets = []string{
	".ApplicationFactory/product.yaml",
	".ApplicationFactory/config.yaml",
	".ApplicationFactory/orchestrator/HERMES-CONTROLLER-PROMPT.md",
	".ApplicationFactory/orchestrator/SPRINT-RUN-LOOP.md",
	".ApplicationFactory/orchestrator/SERVER-REGISTRY-TEMPLATE.yaml",
	".ApplicationFactory/orchestrator/RUNTIME-STATE.yaml",
	".ApplicationFactory/task-packages/TASK-PACKAGE-TEMPLATE.md",
	".github/pull_request_template.md",
	".github/workflows/myfactory-checks.yml",
	".claude/commands/myfactory-discover.md",
	".claude/commands/myfactory-run-sprint.md",
	"AGENTS.md",
	"docs/00-product/idea-brief.md",
	"docs/03-delivery/missions.yaml",
	"docs/03-delivery/sprints.yaml",
	"docs/03-delivery/work-breakdown.yaml",
}

func runInit(t *testing.T, args ...string) (string, int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := Init(args, &out, &errOut)
	return out.String() + errOut.String(), code
}

func hashTree(t *testing.T, root string) map[string][32]byte {
	t.Helper()
	hashes := map[string][32]byte{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		hashes[rel] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatalf("hashTree: %v", err)
	}
	return hashes
}

func TestInitEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test", "--description", "Testing")
	if code != 0 {
		t.Fatalf("init failed (%d): %s", code, out)
	}
	for _, rel := range expectedHiddenAssets {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if info, err := os.Stat(p); err != nil || info.Size() == 0 {
			t.Errorf("expected file missing or empty after init: %s", rel)
		}
	}
	if !strings.Contains(out, "Files created:") || !strings.Contains(out, "Files overwritten: 0") {
		t.Errorf("summary missing from output:\n%s", out)
	}
	if !strings.Contains(out, "myfactory doctor") {
		t.Errorf("next steps missing from output:\n%s", out)
	}
}

func TestInitPreservesExistingReadme(t *testing.T) {
	dir := t.TempDir()
	readme := filepath.Join(dir, "README.md")
	original := []byte("# Existing App\nwith existing content\n")
	if err := os.WriteFile(readme, original, 0o644); err != nil {
		t.Fatal(err)
	}
	out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test")
	if code != 0 {
		t.Fatalf("init failed: %s", out)
	}
	after, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original, after) {
		t.Error("existing README.md was modified by init")
	}
}

func TestInitPlaceholdersReplaced(t *testing.T) {
	dir := t.TempDir()
	out, code := runInit(t, "--target", dir, "--key", "SHOP", "--name", "Shop App", "--description", "A shop")
	if code != 0 {
		t.Fatalf("init failed: %s", out)
	}
	product, err := os.ReadFile(filepath.Join(dir, ".ApplicationFactory", "product.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(product)
	for _, want := range []string{`key: "SHOP"`, `name: "Shop App"`, `description: "A shop"`} {
		if !strings.Contains(text, want) {
			t.Errorf("product.yaml missing %s:\n%s", want, text)
		}
	}
	wb, err := os.ReadFile(filepath.Join(dir, "docs", "03-delivery", "work-breakdown.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wb), `key: "SHOP"`) {
		t.Error("work-breakdown.yaml placeholder not replaced")
	}
}

func TestInitDryRunMakesNoChanges(t *testing.T) {
	dir := t.TempDir()
	out, code := runInit(t, "--target", dir, "--dry-run")
	if code != 0 {
		t.Fatalf("dry-run failed: %s", out)
	}
	if !strings.Contains(out, "DRY RUN") || !strings.Contains(out, "planned") {
		t.Errorf("dry-run output unexpected:\n%s", out)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dry-run created %d entries; want 0", len(entries))
	}
}

func TestInitIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	if out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test", "--description", "d"); code != 0 {
		t.Fatalf("first init failed: %s", out)
	}
	before := hashTree(t, dir)

	out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test", "--description", "d")
	if code != 0 {
		t.Fatalf("second init failed: %s", out)
	}
	if !strings.Contains(out, "Files created:     0") {
		t.Errorf("second init should create zero files:\n%s", out)
	}
	if strings.Contains(out, "Files overwritten: 0") == false {
		t.Errorf("second init must not overwrite:\n%s", out)
	}
	after := hashTree(t, dir)
	if len(before) != len(after) {
		t.Fatalf("file count changed: %d -> %d", len(before), len(after))
	}
	for rel, h := range before {
		if after[rel] != h {
			t.Errorf("file modified by repeat init: %s", rel)
		}
	}
}

func TestInitForceOverwritesManagedFiles(t *testing.T) {
	dir := t.TempDir()
	if out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test"); code != 0 {
		t.Fatalf("first init failed: %s", out)
	}
	// User customizes a managed file and owns an unrelated file.
	managed := filepath.Join(dir, "docs", "00-product", "prd.md")
	if err := os.WriteFile(managed, []byte("user edits"), 0o644); err != nil {
		t.Fatal(err)
	}
	unrelated := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(unrelated, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Init Test", "--force")
	if code != 0 {
		t.Fatalf("force init failed: %s", out)
	}
	if !strings.Contains(out, "[overwritten]") {
		t.Errorf("force run should report overwritten files:\n%s", out)
	}
	data, _ := os.ReadFile(managed)
	if string(data) == "user edits" {
		t.Error("--force did not overwrite managed file")
	}
	mine, _ := os.ReadFile(unrelated)
	if string(mine) != "mine" {
		t.Error("--force touched a file MyFactory does not manage")
	}
}

func TestInitPathWithSpacesAndNonASCII(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my factory pröjéct 日本")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	out, code := runInit(t, "--target", dir, "--key", "TEST", "--name", "Test")
	if code != 0 {
		t.Fatalf("init in unicode path failed: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".ApplicationFactory", "product.yaml")); err != nil {
		t.Errorf("product.yaml missing in unicode path: %v", err)
	}
}

func TestInitOverlayFlags(t *testing.T) {
	dir := t.TempDir()
	out, code := runInit(t, "--target", dir, "--no-codex", "--no-claude", "--key", "TEST", "--name", "T")
	if code != 0 {
		t.Fatalf("init failed: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Error("--no-codex still created AGENTS.md")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		t.Error("--no-claude still created .claude/")
	}
}

func TestInitInvalidKey(t *testing.T) {
	dir := t.TempDir()
	out, code := runInit(t, "--target", dir, "--key", "bad key!")
	if code == 0 {
		t.Fatalf("invalid key accepted:\n%s", out)
	}
}

func TestInitGitlabProvider(t *testing.T) {
	dir := t.TempDir()
	out, code := runInit(t, "--target", dir, "--git-provider", "gitlab", "--key", "TEST", "--name", "T")
	if code != 0 {
		t.Fatalf("init failed: %s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitlab", "merge_request_templates", "Default.md")); err != nil {
		t.Error("gitlab provider did not create MR template")
	}
	if _, err := os.Stat(filepath.Join(dir, ".github")); err == nil {
		t.Error("gitlab provider should not create .github without --with-github")
	}
	config, err := os.ReadFile(filepath.Join(dir, ".ApplicationFactory", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(config), "provider: gitlab") {
		t.Error("config.yaml provider not set to gitlab")
	}
}

func TestDeriveKeyAndName(t *testing.T) {
	cases := []struct{ folder, key, name string }{
		{"clinic-booking", "CLINIC_BOOKING", "Clinic Booking"},
		{"my_app", "MY_APP", "My App"},
		{"9lives", "P_9LIVES", "9lives"},
		{"###", "APP", "###"},
		{"averyveryverylongfoldernamehere", "AVERYVERYVERYLON", "Averyveryverylongfoldernamehere"},
	}
	for _, c := range cases {
		if got := DeriveKey(c.folder); got != c.key {
			t.Errorf("DeriveKey(%q) = %q, want %q", c.folder, got, c.key)
		}
		if got := DeriveName(c.folder); got != c.name {
			t.Errorf("DeriveName(%q) = %q, want %q", c.folder, got, c.name)
		}
	}
}
