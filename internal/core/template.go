package core

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// ExecuteTemplate, verilen içeriği (content) sağlanan veri (data) ile işler.
// data genellikle *core.SystemContext olacaktır.
func ExecuteTemplate(content string, data interface{}) (string, error) {
	// missingkey=zero allows optional variables (returning nil/zero), which works with Sprig's 'default'.
	// Use 'required' function from Sprig for mandatory variables.
	tmpl, err := template.New("veto").Funcs(sprig.TxtFuncMap()).Option("missingkey=zero").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
