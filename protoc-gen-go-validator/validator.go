package main

import (
	"github.com/tenz-io/gokit/genproto/go/custom/idl"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	fmtPkg      = protogen.GoImportPath("fmt")
	stringsPkg  = protogen.GoImportPath("strings")
	genprotoPkg = protogen.GoImportPath("github.com/tenz-io/gokit/genproto")
	//idlPkg      = protogen.GoImportPath("github.com/tenz-io/gokit/genproto/go/custom/idl")
	//protoPkg    = protogen.GoImportPath("google.golang.org/protobuf/proto")
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
	g.P("// ", fmtPkg.Ident(""))
	g.P("// ", stringsPkg.Ident(""))
	g.P("// ", genprotoPkg.Ident(""))
	g.P()

	genMessages(gen, file, g)

	return g
}

func genMessages(_ *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	for _, msg := range file.Messages {
		msgName := string(msg.Desc.Name())
		fields := msgFields(msgName, msg)
		msgTpl := messageTemplate{
			MessageName: msgName,
			Fields:      fields,
		}
		g.P(msgTpl.execute())
	}
}

func msgFields(msgName string, msg *protogen.Message) []fieldData {
	var fields []fieldData
	for _, field := range msg.Fields {
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
			FieldName:   field.GoName,
			Int:         fieldOpts.GetInt(),
			Str:         fieldOpts.GetStr(),
			Bytes:       fieldOpts.GetBytes(),
			Array:       fieldOpts.GetArray(),
			Float:       fieldOpts.GetFloat(),
		})
	}

	return fields
}
