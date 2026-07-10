package cli

import (
	"bytes"
	"strings"
	"testing"
)

func run(args ...string) (string, string, int) {
	var out, errOut bytes.Buffer
	code := Run(args, &out, &errOut)
	return out.String(), errOut.String(), code
}

func TestHelp(t *testing.T) {
	for _, args := range [][]string{nil, {"--help"}, {"-h"}, {"help"}} {
		out, _, code := run(args...)
		if code != 0 {
			t.Errorf("help exit %d for %v", code, args)
		}
		for _, cmd := range []string{"init", "doctor", "discover", "plan", "plane", "orchestrator", "version"} {
			if !strings.Contains(out, cmd) {
				t.Errorf("help missing command %q for args %v", cmd, args)
			}
		}
	}
}

func TestVersionFlag(t *testing.T) {
	out, _, code := run("--version")
	if code != 0 {
		t.Fatalf("--version exit %d", code)
	}
	if !strings.HasPrefix(out, "myfactory ") {
		t.Errorf("--version output = %q", out)
	}
}

func TestVersionCommand(t *testing.T) {
	out, _, code := run("version")
	if code != 0 {
		t.Fatalf("version exit %d", code)
	}
	if !strings.Contains(out, "MyFactory version:") {
		t.Errorf("version output = %q", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	_, errOut, code := run("frobnicate")
	if code != 2 {
		t.Errorf("unknown command exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "unknown command") {
		t.Errorf("stderr = %q", errOut)
	}
}
