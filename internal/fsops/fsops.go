// Package fsops implements safe file operations for `myfactory init`:
// nothing is ever deleted, and existing files are never overwritten unless
// force is explicitly requested. Every action is tracked for the summary.
package fsops

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ignoredNames are never copied out of templates.
var ignoredNames = map[string]bool{
	".DS_Store":   true,
	"__pycache__": true,
	".git":        true,
}

// Actions records what happened to each file during an init/copy run.
// Paths are stored relative to the target directory using the OS separator.
type Actions struct {
	Created     []string
	Skipped     []string
	Overwritten []string
}

// SummaryLines renders the created/skipped/overwritten report exactly like
// the legacy Python CLI.
func (a *Actions) SummaryLines() []string {
	lines := []string{
		fmt.Sprintf("Files created:     %d", len(a.Created)),
		fmt.Sprintf("Files skipped:     %d", len(a.Skipped)),
		fmt.Sprintf("Files overwritten: %d", len(a.Overwritten)),
	}
	for _, group := range []struct {
		label string
		items []string
	}{
		{"created", a.Created},
		{"skipped", a.Skipped},
		{"overwritten", a.Overwritten},
	} {
		for _, p := range group.items {
			lines = append(lines, fmt.Sprintf("  [%s] %s", group.label, p))
		}
	}
	return lines
}

// YAMLString returns a YAML-safe string using JSON string escaping
// (YAML accepts JSON-style quoted strings).
func YAMLString(value string) string {
	b, _ := json.Marshal(value)
	return string(b)
}

// WriteFile writes content, respecting no-overwrite-by-default.
// rel is the path recorded in the summary.
func WriteFile(path, rel, content string, actions *Actions, force, dryRun bool) error {
	if _, err := os.Stat(path); err == nil {
		if !force {
			actions.Skipped = append(actions.Skipped, rel)
			return nil
		}
		if !dryRun {
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return err
			}
		}
		actions.Overwritten = append(actions.Overwritten, rel)
		return nil
	}
	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
	}
	actions.Created = append(actions.Created, rel)
	return nil
}

// CopyTree copies every file under srcFS into destDir.
// Existing destination files are skipped unless force is set. Nothing is
// ever deleted. exclude holds slash-separated relative paths to skip.
func CopyTree(srcFS fs.FS, destDir string, actions *Actions, force, dryRun bool, exclude map[string]bool) error {
	return fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		for _, part := range strings.Split(path, "/") {
			if ignoredNames[part] {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		if exclude != nil && exclude[path] {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		dest := filepath.Join(destDir, filepath.FromSlash(path))
		if d.IsDir() {
			if !dryRun {
				return os.MkdirAll(dest, 0o755)
			}
			return nil
		}
		content, err := fs.ReadFile(srcFS, path)
		if err != nil {
			return err
		}
		rel := filepath.FromSlash(path)
		if _, err := os.Stat(dest); err == nil {
			if !force {
				actions.Skipped = append(actions.Skipped, rel)
				return nil
			}
			if !dryRun {
				if err := os.WriteFile(dest, content, 0o644); err != nil {
					return err
				}
			}
			actions.Overwritten = append(actions.Overwritten, rel)
			return nil
		}
		if !dryRun {
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(dest, content, 0o644); err != nil {
				return err
			}
		}
		actions.Created = append(actions.Created, rel)
		return nil
	})
}

// EnsureDir creates a directory (and parents) unless dryRun.
func EnsureDir(path string, dryRun bool) error {
	if dryRun {
		return nil
	}
	return os.MkdirAll(path, 0o755)
}

// ReplacePlaceholders replaces the first occurrence of each placeholder that
// is present in the file. It never fails when a placeholder is absent (the
// file may have been customized already). Returns true if anything changed.
func ReplacePlaceholders(path string, replacements [][2]string, dryRun bool) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	text := string(data)
	changed := false
	for _, r := range replacements {
		if strings.Contains(text, r[0]) {
			text = strings.Replace(text, r[0], r[1], 1)
			changed = true
		}
	}
	if changed && !dryRun {
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			return false, err
		}
	}
	return changed, nil
}

// FileExists reports whether path is an existing regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// DirExists reports whether path is an existing directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ReadFileString reads a file as UTF-8 text.
func ReadFileString(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	return string(data), err
}
