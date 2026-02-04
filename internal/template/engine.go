package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Engine handles template rendering
type Engine struct {
	templates map[string]*template.Template
}

// NewEngine creates a new template engine
func NewEngine() *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
	}
}

// AddTemplate adds a template with the given name
func (e *Engine) AddTemplate(name, tmpl string) error {
	// Sanitize template name - Go templates don't allow hyphens in names
	// Use the name as-is for lookup, but sanitize for template.New()
	safeName := strings.ReplaceAll(name, "-", "_")
	t, err := template.New(safeName).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	e.templates[name] = t
	return nil
}

// Render renders a template with the given data
func (e *Engine) Render(name string, data map[string]interface{}) (string, error) {
	t, ok := e.templates[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}

// RenderAll renders all templates with the given data
func (e *Engine) RenderAll(data map[string]interface{}) (map[string]string, error) {
	results := make(map[string]string)
	for name := range e.templates {
		rendered, err := e.Render(name, data)
		if err != nil {
			return nil, err
		}
		results[name] = rendered
	}
	return results, nil
}
