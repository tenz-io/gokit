package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

//go:embed template.go.tpl
var tpl string

type messageTemplate struct {
	Messages []messageData
}

type messageData struct {
	deprecated bool
	Name       string
	Fields     []fieldData
}

type fieldData struct {
	Name  string
	Int   *idl.IntField
	Str   *idl.StringField
	Bytes *idl.BytesField
	Array *idl.ArrayField
	Float *idl.FloatField
}

func (d *messageTemplate) execute() string {
	if len(d.Messages) == 0 {
		return ""
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("message").Parse(strings.TrimSpace(tpl))
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(buf, d); err != nil {
		panic(err)
	}
	return buf.String()
}
