package main

import (
	"encoding/json"
	"fmt"

	"github.com/tenz-io/gokit/notionx"
)

func main() {
	md := "# hello **notionx**"
	blocks, err := notionx.MarkdownToNotionBlocks(md)
	if err != nil {
		fmt.Printf("MarkdownToNotionBlocks() error = %v", err)
		return
	}
	j, err := json.Marshal(blocks)
	if err != nil {
		fmt.Printf("json.Marshal() error = %v", err)
		return
	}
	fmt.Printf("blocks: %s", j)
}
