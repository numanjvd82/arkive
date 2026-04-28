package email

import (
	"bytes"
	"errors"
	"html/template"
	"strings"

	"arkive/pkg/email/templates"
)

func RenderHTMLTemplate(templateFile string, data any) (string, error) {
	templateFile = strings.TrimSpace(templateFile)
	if templateFile == "" {
		return "", errors.New("missing template file")
	}
	tmplBytes, err := templates.FS.ReadFile(templateFile)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(templateFile).Parse(string(tmplBytes))
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}
