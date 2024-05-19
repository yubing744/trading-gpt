package xtemplate

import (
	"bytes"
	"text/template"
)

// add is a custom template function that adds two integers.
func add(x, y int) int {
	return x + y
}

// Render function that takes a template string and a context map, and returns the rendered string or an error.
func Render(tmplStr string, ctx map[string]interface{}) (string, error) {
	// Parse the template string.
	tmpl, err := template.New("tmpl").Funcs(template.FuncMap{
		"add": add,
	}).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	// Create a buffer to hold the rendered output.
	var buf bytes.Buffer

	// Execute the template with the provided context and write the output to the buffer.
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	// Return the rendered string.
	return buf.String(), nil
}
