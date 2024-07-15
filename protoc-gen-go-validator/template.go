package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

var (
	//go:embed template.go.tpl
	msgTpl string
	//go:embed init.go.tpl
	initTpl string
)

type (
	// messageTemplate is a template for generating validation code for a message.
	messageTemplate struct {
		MessageName string
		Fields      []fieldData
	}

	// initTemplate is a template for generating init code.
	initTemplate struct{}
)

type fieldData struct {
	MessageName string
	FieldName   string
	IsMessage   bool
	Int         *idl.IntField    // if field type is int/uint/*int/*uint
	Str         *idl.StringField // if field type is string/*string
	Bytes       *idl.BytesField  // if field type is bytes
	Array       *idl.ArrayField  // if field type is repeated
	Float       *idl.FloatField  // if field type is float32/float64
}

func (d *messageTemplate) execute() string {
	if d == nil || d.MessageName == "" {
		panic("message name is required")
	}

	if len(d.Fields) == 0 {
		return ""
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("message").Parse(strings.TrimSpace(msgTpl))
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(buf, d); err != nil {
		panic(err)
	}
	return buf.String()
}

func (d *initTemplate) execute() string {
	buf := new(bytes.Buffer)
	tmpl, err := template.New("init").Parse(strings.TrimSpace(initTpl))
	if err != nil {
		panic(err)
	}
	if err = tmpl.Execute(buf, d); err != nil {
		panic(err)
	}
	return buf.String()
}
