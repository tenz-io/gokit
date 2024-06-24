// main.go
package main

import (
	_ "embed"
	"fmt"
	"os"
)

func main() {
	// Define the data to be used in the template
	td := &templateData{
		Package: "",
		Messages: []messageData{
			{
				Name: "LoginRequest",
			},
			{
				Name: "LoginResponse",
			},
			{
				Name: "IndexRequest",
			},
		},
	}

	// Create the output file
	outputFile, err := os.Create("generated.go")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Execute the template with the data and write to the output file
	out, err := td.execute()
	if err != nil {
		fmt.Println("Error executing template:", err)
		return
	}

	_, err = outputFile.WriteString(out)
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		return
	}

	fmt.Println("Generated Go source code in generated.go")
}
