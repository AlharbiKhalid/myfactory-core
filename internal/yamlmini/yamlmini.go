// Package yamlmini loads YAML into plain map[string]any / []any structures.
//
// It started as a hand-written parser for the narrow YAML subset MyFactory's
// own templates emit. Delivery files are now written by AI planning agents
// (Claude, Codex) and ordinary editors, which produce standards-compliant
// YAML — indentless block sequences, anchors and aliases, multiline plain
// scalars, literal/folded block scalars, and so on. Parsing therefore
// delegates to gopkg.in/yaml.v3. The dependency is compiled into the
// standalone binary; end users still need no runtime.
//
// The helper API (LoadFile, Parse, GetPath, GetBool, GetString, Items) is
// unchanged. ParseWithError exposes parse failures; LoadFile now returns
// them instead of silently yielding an empty mapping.
package yamlmini

import (
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// LoadFile parses a YAML file. Missing or empty files yield an empty map,
// mirroring the legacy Python implementation (many MyFactory files are
// optional). Unreadable or invalid YAML returns a descriptive error that
// includes the file path.
func LoadFile(path string) (map[string]any, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	data, err := ParseWithError(string(text))
	if err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}
	return data, nil
}

// Parse parses YAML text. Returns an empty map for blank input, input whose
// top level is not a mapping, or invalid YAML (legacy behavior). Callers
// that must distinguish invalid YAML from an empty document should use
// ParseWithError.
func Parse(text string) map[string]any {
	data, err := ParseWithError(text)
	if err != nil {
		return map[string]any{}
	}
	return data
}

// ParseWithError parses YAML text into a map. Blank or comment-only input
// yields an empty map. Invalid YAML, or a document whose top level is not a
// mapping, returns a non-nil error.
func ParseWithError(text string) (map[string]any, error) {
	if strings.TrimSpace(text) == "" {
		return map[string]any{}, nil
	}
	var doc any
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		return nil, err
	}
	switch v := doc.(type) {
	case nil:
		// Comment-only input or an explicit null document.
		return map[string]any{}, nil
	case map[string]any:
		return v, nil
	default:
		return nil, fmt.Errorf("top-level YAML value must be a mapping, got %T", doc)
	}
}

// GetPath fetches a nested key like "plane.enabled". Returns def when any
// segment is missing or not a mapping.
func GetPath(data map[string]any, dotted string, def any) any {
	var current any = data
	for _, part := range strings.Split(dotted, ".") {
		m, ok := current.(map[string]any)
		if !ok {
			return def
		}
		v, ok := m[part]
		if !ok {
			return def
		}
		current = v
	}
	return current
}

// GetBool interprets a config value as a boolean.
func GetBool(data map[string]any, dotted string, def bool) bool {
	v := GetPath(data, dotted, def)
	b, ok := v.(bool)
	if !ok {
		return def
	}
	return b
}

// GetString interprets a config value as a string.
func GetString(data map[string]any, dotted string, def string) string {
	v := GetPath(data, dotted, def)
	s, ok := v.(string)
	if !ok || s == "" {
		return def
	}
	return s
}

// Items returns the list entries under key that are mappings.
func Items(data map[string]any, key string) []map[string]any {
	var out []map[string]any
	list, ok := data[key].([]any)
	if !ok {
		return out
	}
	for _, v := range list {
		if m, ok := v.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
