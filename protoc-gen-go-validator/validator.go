package main

import (
	"github.com/tenz-io/gokit/genproto/go/custom/idl"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	genprotoPkg = protogen.GoImportPath("github.com/tenz-io/gokit/genproto")
	idlPkg      = protogen.GoImportPath("github.com/tenz-io/gokit/genproto/go/custom/idl")
	protoPkg    = protogen.GoImportPath("google.golang.org/protobuf/proto")
)

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Messages) == 0 {
		return nil
	}

	filename := file.GeneratedFilenamePrefix + "_validate.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by github.com/tenz-io/gokit/protoc-gen-go-validator. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the github.com/tenz-io/gokit/protoc-gen-go-validator package it is being compiled against.")
	g.P("// ", genprotoPkg.Ident(""))
	g.P("// ", idlPkg.Ident(""), protoPkg.Ident(""))
	g.P()

	genMessages(gen, file, g)

	return g
}

func genMessages(_ *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	msgTpl := &messageTemplate{
		Messages: []messageData{},
	}

	for _, msg := range file.Messages {
		deprecated := msg.Desc.Options().(*descriptorpb.MessageOptions).GetDeprecated()
		msgName := string(msg.Desc.Name())
		fields := msgFields(msgName, msg)
		msgData := messageData{
			deprecated: deprecated,
			Name:       msgName,
			Fields:     fields,
			FieldSet:   buildFieldSet(fields),
		}
		msgTpl.Messages = append(msgTpl.Messages, msgData)
	}

	g.P(msgTpl.execute())
}

func buildFieldSet(fields []fieldData) map[string]fieldData {
	fieldSet := make(map[string]fieldData)
	for _, field := range fields {
		fieldSet[field.Name] = field
	}
	return fieldSet
}

func msgFields(msgName string, msg *protogen.Message) []fieldData {
	var fields []fieldData
	for _, field := range msg.Fields {

		// check if is message type or pointer of message type
		if field.Desc.Kind() == protoreflect.MessageKind {
			continue
		}

		options := proto.GetExtension(field.Desc.Options(), idl.E_Field)
		if options == nil {
			continue
		}
		fieldOpts, ok := options.(*idl.Field)
		if !ok {
			continue
		}

		fields = append(fields, fieldData{
			MessageName: msgName,
			Name:        field.GoName,
			Int:         fieldOpts.GetInt(),
			Str:         fieldOpts.GetStr(),
			Bytes:       fieldOpts.GetBytes(),
			Array:       fieldOpts.GetArray(),
			Float:       fieldOpts.GetFloat(),
		})
	}

	return fields
}
