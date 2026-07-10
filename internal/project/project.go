// Package project resolves MyFactory project paths inside a target repo.
package project

import (
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
