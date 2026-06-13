package panel

import (
	"encoding/json"
	"strings"
	"unicode"
)

func prettyJSON(raw string, maxBytes int) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "(empty)"
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return truncateJSON(raw, maxBytes)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncateJSON(raw, maxBytes)
	}
	return truncateJSON(string(out), maxBytes)
}

func truncateJSON(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "\n… (truncated)"
}

func isJSON(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	var v any
	return json.Unmarshal([]byte(raw), &v) == nil
}

func jsonCodeBlock(raw string, maxBytes int) string {
	pretty := prettyJSON(raw, maxBytes)
	return `<pre class="code-block terminal-panel json-block"><code>` + highlightJSON(pretty) + `</code></pre>`
}

func plainCodeBlock(raw string) string {
	return `<pre class="code-block terminal-panel json-block"><code>` + htmlEscape(raw) + `</code></pre>`
}

// highlightJSON adds span classes for keys, strings, numbers, booleans, and null.
func highlightJSON(raw string) string {
	var b strings.Builder
	i := 0
	for i < len(raw) {
		c := raw[i]
		switch {
		case c == '"':
			j := i + 1
			for j < len(raw) {
				if raw[j] == '\\' {
					j += 2
					if j > len(raw) {
						break
					}
					continue
				}
				if raw[j] == '"' {
					j++
					break
				}
				j++
			}
			str := raw[i:j]
			cls := "json-str"
			k := j
			for k < len(raw) && (raw[k] == ' ' || raw[k] == '\t' || raw[k] == '\n' || raw[k] == '\r') {
				k++
			}
			if k < len(raw) && raw[k] == ':' {
				cls = "json-key"
			}
			writeJSONSpan(&b, cls, str)
			i = j
		case c == '-' || unicode.IsDigit(rune(c)):
			j := i + 1
			for j < len(raw) && isJSONNumberChar(raw[j]) {
				j++
			}
			writeJSONSpan(&b, "json-num", raw[i:j])
			i = j
		case strings.HasPrefix(raw[i:], "true"):
			writeJSONSpan(&b, "json-bool", "true")
			i += 4
		case strings.HasPrefix(raw[i:], "false"):
			writeJSONSpan(&b, "json-bool", "false")
			i += 5
		case strings.HasPrefix(raw[i:], "null"):
			writeJSONSpan(&b, "json-null", "null")
			i += 4
		case c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ':':
			writeJSONSpan(&b, "json-punct", string(c))
			i++
		default:
			b.WriteString(htmlEscape(string(c)))
			i++
		}
	}
	return b.String()
}

func isJSONNumberChar(c byte) bool {
	return unicode.IsDigit(rune(c)) || c == '.' || c == 'e' || c == 'E' || c == '+' || c == '-'
}

func writeJSONSpan(b *strings.Builder, cls, text string) {
	b.WriteString(`<span class="`)
	b.WriteString(cls)
	b.WriteString(`">`)
	b.WriteString(htmlEscape(text))
	b.WriteString(`</span>`)
}
