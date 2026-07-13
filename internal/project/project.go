// Package project resolves MyFactory project paths inside a target repo.
package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// MetadataDirName is the MyFactory metadata directory inside a project.
const MetadataDirName = ".ApplicationFactory"

// ResolveTarget resolves a --target argument (default: current directory).
func ResolveTarget(target string) (string, error) {
	if target == "" {
		return os.Getwd()
	}
	return filepath.Abs(expandHome(target))
}

func expandHome(path string) string {
	if path == "~" || len(path) >= 2 && path[:2] == "~/" || len(path) >= 2 && path[:2] == `~\` {
		home, err := os.UserHomeDir()
		if err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// EnsureDir returns an error when target is not an existing directory.
// --target always refers to a local path; MyFactory never accesses remote
// filesystems, so a nonexistent target is a user error, not an empty project.
func EnsureDir(target string) error {
	info, err := os.Stat(target)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("target directory does not exist: %s", target)
	}
	return nil
}

// EnsureInitialized returns an error when target is not an existing,
// MyFactory-initialized directory (marked by .ApplicationFactory/config.yaml).
func EnsureInitialized(target string) error {
	if err := EnsureDir(target); err != nil {
		return err
	}
	if info, err := os.Stat(Config(target)); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("target is not a MyFactory project (missing %s): %s\nRun `myfactory init --target %s` to initialize it",
			filepath.Join(MetadataDirName, "config.yaml"), target, target)
	}
	return nil
}

// MetadataDir returns <target>/.ApplicationFactory.
func MetadataDir(target string) string { return filepath.Join(target, MetadataDirName) }

// ProductManifest returns the product.yaml path.
func ProductManifest(target string) string { return filepath.Join(MetadataDir(target), "product.yaml") }

// Config returns the project config.yaml path.
func Config(target string) string { return filepath.Join(MetadataDir(target), "config.yaml") }

// TaskPackagesDir returns the task-packages directory.
func TaskPackagesDir(target string) string {
	return filepath.Join(MetadataDir(target), "task-packages")
}

// OrchestratorDir returns the orchestrator directory.
func OrchestratorDir(target string) string {
	return filepath.Join(MetadataDir(target), "orchestrator")
}
