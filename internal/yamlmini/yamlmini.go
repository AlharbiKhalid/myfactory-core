// Package yamlmini parses the YAML subset that MyFactory itself generates:
// nested mappings, lists of scalars, lists of mappings, quoted strings,
// booleans, null, and numbers. It intentionally avoids a third-party YAML
// dependency; MyFactory only ever reads files it wrote from templates.
//
// It is a port of the legacy Python fallback parser in
// myfactory/core/config.py and preserves its semantics.
package yamlmini

import (
	"os"
	"strconv"
	"strings"
)

// LoadFile parses a YAML-subset file. Missing or empty files yield an empty
// map, mirroring the Python implementation.
func LoadFile(path string) (map[string]any, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	return Parse(string(text)), nil
}

// Parse parses YAML-subset text. Returns an empty map for blank input or
// input whose top level is not a mapping.
func Parse(text string) map[string]any {
	if strings.TrimSpace(text) == "" {
		return map[string]any{}
	}
	lines := collectLines(text)
	value, _ := parseBlock(lines, 0, 0)
	if m, ok := value.(map[string]any); ok {
		return m
	}
	return map[string]any{}
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

// ---------------------------------------------------------------------------
// Parser internals
// ---------------------------------------------------------------------------

type line struct {
	indent  int
	content string
}

func collectLines(text string) []line {
	var out []line
	for _, raw := range strings.Split(text, "\n") {
		stripped := strings.TrimRight(stripComment(raw), " \t\r")
		if strings.TrimSpace(stripped) == "" {
			continue
		}
		indent := len(stripped) - len(strings.TrimLeft(stripped, " "))
		out = append(out, line{indent: indent, content: strings.TrimSpace(stripped)})
	}
	return out
}

func stripComment(s string) string {
	inSingle, inDouble := false, false
	for i, ch := range s {
		switch {
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case ch == '#' && !inSingle && !inDouble:
			if i == 0 || s[i-1] == ' ' || s[i-1] == '\t' {
				return s[:i]
			}
		}
	}
	return s
}

func parseScalar(token string) any {
	token = strings.TrimSpace(token)
	switch token {
	case "", "~", "null":
		return nil
	case "true", "True":
		return true
	case "false", "False":
		return false
	case "[]":
		return []any{}
	case "{}":
		return map[string]any{}
	}
	if len(token) >= 2 {
		first, last := token[0], token[len(token)-1]
		if first == last && (first == '\'' || first == '"') {
			return token[1 : len(token)-1]
		}
	}
	if i, err := strconv.Atoi(token); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(token, 64); err == nil {
		return f
	}
	return token
}

func isListItem(content string) bool {
	return strings.HasPrefix(content, "- ") || content == "-"
}

func parseBlock(lines []line, index, indent int) (any, int) {
	if index >= len(lines) {
		return map[string]any{}, index
	}
	if isListItem(lines[index].content) {
		return parseList(lines, index, indent)
	}
	return parseMapping(lines, index, indent)
}

func parseMapping(lines []line, index, indent int) (map[string]any, int) {
	result := map[string]any{}
	for index < len(lines) {
		li := lines[index]
		if li.indent < indent || isListItem(li.content) {
			break
		}
		if li.indent > indent {
			break
		}
		colon := strings.Index(li.content, ":")
		if colon < 0 {
			break
		}
		key := strings.TrimSpace(li.content[:colon])
		rest := strings.TrimSpace(li.content[colon+1:])
		index++
		switch rest {
		case "", "|", ">", "|-", ">-":
			if rest != "" {
				// Multiline scalar: join deeper-indented lines with spaces.
				var parts []string
				for index < len(lines) && lines[index].indent > li.indent {
					parts = append(parts, lines[index].content)
					index++
				}
				result[key] = strings.Join(parts, " ")
			} else if index < len(lines) && lines[index].indent > li.indent {
				var value any
				value, index = parseBlock(lines, index, lines[index].indent)
				result[key] = value
			} else {
				result[key] = nil
			}
		default:
			result[key] = parseScalar(rest)
		}
	}
	return result, index
}

func parseList(lines []line, index, indent int) ([]any, int) {
	result := []any{}
	for index < len(lines) {
		li := lines[index]
		if li.indent != indent || !isListItem(li.content) {
			break
		}
		itemText := strings.TrimSpace(li.content[1:])
		index++
		switch {
		case itemText == "":
			if index < len(lines) && lines[index].indent > indent {
				var value any
				value, index = parseBlock(lines, index, lines[index].indent)
				result = append(result, value)
			} else {
				result = append(result, nil)
			}
		case strings.Contains(itemText, ":") && !strings.HasPrefix(itemText, "'") && !strings.HasPrefix(itemText, "\""):
			// Inline mapping start: '- key: value' with continuation lines at
			// indent+2 belonging to the same item.
			item := map[string]any{}
			colon := strings.Index(itemText, ":")
			key := strings.TrimSpace(itemText[:colon])
			rest := strings.TrimSpace(itemText[colon+1:])
			if rest != "" {
				item[key] = parseScalar(rest)
			} else if index < len(lines) && lines[index].indent > indent+2 {
				var value any
				value, index = parseBlock(lines, index, lines[index].indent)
				item[key] = value
			} else {
				item[key] = nil
			}
			for index < len(lines) && lines[index].indent == indent+2 && !isListItem(lines[index].content) {
				var sub map[string]any
				sub, index = parseMapping(lines, index, indent+2)
				for k, v := range sub {
					item[k] = v
				}
			}
			result = append(result, item)
		default:
			result = append(result, parseScalar(itemText))
		}
	}
	return result, index
}
