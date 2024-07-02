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
	Name          string
	Int           *idl.IntField    // if field type is int/uint/*int/*uint
	Str           *idl.StringField // if field type is string/*string
	Bytes         *idl.BytesField  // if field type is bytes
	Array         *idl.ArrayField  // if field type is repeated
	Float         *idl.FloatField  // if field type is float32/float64
	subFieldsData []fieldData      // for nested message type or pointer of message type
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
