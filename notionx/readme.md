# notionx

A simple utility to convert *markdown* files to **Notion** Blocks.

## Example

```go 
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

```

**Explanation**:

- The above code will convert the markdown string `# hello **notionx**` to Notion Blocks.
- The output will be a JSON string representing the Notion Blocks.
- The JSON string can be used to create a Notion page using the Notion API.

