// Package gitutil provides read-only git detection helpers.
// It never mutates a repository and degrades gracefully when git is absent.
package gitutil

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// IsGitRepo reports whether dir is inside a git work tree.
func IsGitRepo(dir string) bool {
	if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info != nil {
		return true
	}
	out, err := run(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// RemoteURLs lists the URLs of all configured remotes.
func RemoteURLs(dir string) []string {
	out, err := run(dir, "remote", "-v")
	if err != nil {
		return nil
	}
	var urls []string
	for _, l := range strings.Split(out, "\n") {
		fields := strings.Fields(l)
		if len(fields) >= 2 {
			urls = append(urls, fields[1])
		}
	}
	return urls
}

// DetectProvider returns "github", "gitlab", or "unknown" based on remotes.
func DetectProvider(dir string) string {
	for _, url := range RemoteURLs(dir) {
		lowered := strings.ToLower(url)
		if strings.Contains(lowered, "github.com") {
			return "github"
		}
		if strings.Contains(lowered, "gitlab") {
			return "gitlab"
		}
	}
	return "unknown"
}

func run(dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
