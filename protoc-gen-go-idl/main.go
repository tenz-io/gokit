package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
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
		msgProto := *msg
		msgData := messageData{
			Name:         string(msg.Desc.Name()),
			MessageProto: &msgProto,
		}
		msgTpl.Messages = append(msgTpl.Messages, msgData)

	}

	err := msgTpl.execute()
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	fmt.Printf("Generated file: %s\n", filename)

}
