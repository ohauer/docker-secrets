package template

import (
	"testing"
)

func FuzzRender(f *testing.F) {
	// Seed with valid templates
	f.Add("{{ .value }}")
	f.Add("{{ .key }}: {{ .value }}")
	f.Add("plain text")

	// Seed with attack vectors
	f.Add("{{ range . }}{{ . }}{{ end }}")
	f.Add("{{ . }}")
	f.Add("{{")
	f.Add("}}")
	f.Add("{{ .nonexistent }}")
	f.Add("{{ printf \"%s\" . }}")
	f.Add("{{ . | . | . }}")
	f.Add(string(make([]byte, 10000))) // Large template

	f.Fuzz(func(t *testing.T, tmpl string) {
		engine := NewEngine()
		data := map[string]interface{}{
			"value": "test",
			"key":   "testkey",
		}

		// Should not panic
		_, _ = engine.Render(tmpl, data)
	})
}
