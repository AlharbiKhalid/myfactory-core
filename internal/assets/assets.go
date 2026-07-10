// Package assets provides access to the reusable MyFactory templates.
//
// Production builds always use the assets embedded in the binary. For local
// template development, MYFACTORY_ASSETS_DIR may point at a directory with
// the same layout as the repo's templates/ directory (containing
// product-repo/ and project-overlays/); it then overrides the embedded copy.
package assets

import (
	"fmt"
	"io/fs"
	"os"

	myfactorycore "github.com/AlharbiKhalid/myfactory-core"
)

// EnvOverride is the environment variable for local template development.
const EnvOverride = "MYFACTORY_ASSETS_DIR"

// Base returns the templates root: a directory containing product-repo/ and
// project-overlays/.
func Base() (fs.FS, error) {
	if dir := os.Getenv(EnvOverride); dir != "" {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("%s=%q is not a readable directory", EnvOverride, dir)
		}
		return os.DirFS(dir), nil
	}
	return fs.Sub(myfactorycore.TemplatesFS, "templates")
}

func sub(path string) (fs.FS, error) {
	base, err := Base()
	if err != nil {
		return nil, err
	}
	return fs.Sub(base, path)
}

// ProductRepo returns the product template copied by `myfactory init`.
func ProductRepo() (fs.FS, error) { return sub("product-repo") }

// CodexOverlay returns the Codex overlay (AGENTS.md).
func CodexOverlay() (fs.FS, error) { return sub("project-overlays/codex") }

// ClaudeOverlay returns the Claude overlay (.claude/commands).
func ClaudeOverlay() (fs.FS, error) { return sub("project-overlays/claude") }

// ReadFile reads a file by path relative to the templates root, e.g.
// "product-repo/.ApplicationFactory/orchestrator/HERMES-CONTROLLER-PROMPT.md".
func ReadFile(rel string) ([]byte, error) {
	base, err := Base()
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(base, rel)
}
