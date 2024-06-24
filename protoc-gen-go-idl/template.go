package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

//go:embed template.go.tpl
var tpl string

type messageTemplate struct {
	Filename string
	Package  string
	Messages []messageData
}

type messageData struct {
	Name   string
	Fields []fieldData
}

type fieldData struct {
	Name        string
	IntField    *idl.IntField
	StringField *idl.StringField
	BytesField  *idl.BytesField
	ArrayField  *idl.ArrayField
}

func (d *messageTemplate) execute() error {
	f, err := createFileIfNotExist(d.Filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("message").Parse(strings.TrimSpace(tpl))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = tmpl.Execute(f, d)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	return nil
}

// createFileIfNotExist creates a file if it does not exist
// including the directories
func createFileIfNotExist(filename string) (*os.File, error) {
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// create the directory
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return os.Create(filename)
	}
	return os.Create(filename)
}
