package main

import (
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

//go:embed template.go.tpl
var tpl string

type templateData struct {
	Package  string
	Messages []messageData
}

type messageData struct {
	Name   string
	Fields []FieldData
}

type FieldData struct {
	Name        string
	IntField    *idl.IntField
	StringField *idl.StringField
	// Add other field types as needed
}

func (d *templateData) execute() (string, error) {
	tmpl, err := template.New("message").Parse(strings.TrimSpace(tpl))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	buf := new(strings.Builder)
	err = tmpl.Execute(buf, d)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}
