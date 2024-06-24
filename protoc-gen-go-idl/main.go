package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/tenz-io/gokit/genproto/go/custom/idl"
)

func main() {
	var flags flag.FlagSet
	opts := protogen.Options{
		ParamFunc: flags.Set,
	}

	opts.Run(func(plugin *protogen.Plugin) error {
		for _, file := range plugin.Files {
			if !file.Generate {
				continue
			}
			generateValidationMethod(plugin, file)
		}
		return nil
	})

}

func generateValidationMethod(plugin *protogen.Plugin, file *protogen.File) {
	filename := file.GeneratedFilenamePrefix + ".validation.go"

	msgTpl := &messageTemplate{
		Filename: filename,
		Package:  string(file.GoPackageName),
		Messages: []messageData{},
	}

	for _, msg := range file.Messages {
		mdata := messageData{
			Name:   string(msg.Desc.Name()),
			Fields: []fieldData{},
		}

		for _, field := range msg.Fields {
			options := proto.GetExtension(field.Desc.Options(), idl.E_Field)
			if options == nil {
				continue
			}
			fieldOpts, ok := options.(*idl.Field)
			if !ok {
				continue
			}

			fdata := fieldData{
				Name:        field.GoName,
				IntField:    fieldOpts.GetInt(),
				StringField: fieldOpts.GetStr(),
				BytesField:  fieldOpts.GetBytes(),
				ArrayField:  fieldOpts.GetArray(),
			}

			mdata.Fields = append(mdata.Fields, fdata)
		}

		msgTpl.Messages = append(msgTpl.Messages, mdata)

	}

	err := msgTpl.execute()
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	fmt.Printf("Generated file: %s\n", filename)

}
