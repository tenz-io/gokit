package notionx

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tenz-io/notionapi"
)

func TestMarkdownDoc(t *testing.T) {
	t.Run("markdown", func(t *testing.T) {
		// read markdown file

		mdFile, err := os.ReadFile("readme.md")
		if err != nil {
			t.Errorf("os.ReadFile() error = %v", err)
			return
		}

		// convert markdown to notion blocks
		blocks, err := MarkdownToNotionBlocks(string(mdFile))
		if err != nil {
			t.Errorf("MarkdownToNotionBlocks() error = %v", err)
			return
		}

		// convert notion blocks to json
		j, err := json.Marshal(blocks)
		if err != nil {
			t.Errorf("json.Marshal() error = %v", err)
			return
		}

		t.Logf("blocks: %s", j)

	})
}

func TestMarkdownToNotionBlocks(t *testing.T) {
	type args struct {
		markdown string
	}
	tests := []struct {
		name    string
		args    args
		want    []notionapi.Block
		wantErr bool
	}{
		{
			name: "paragraph",
			args: args{markdown: "**hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "heading 1",
			args: args{markdown: "# **hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "heading 2",
			args: args{markdown: "## **hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "heading 3",
			args: args{markdown: "### **hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "numbered list",
			args: args{markdown: "1. **hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "bulleted list",
			args: args{markdown: "- **hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "quote",
			args: args{markdown: "> **hello** *world* golang <u>developer</u> `channel`"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
		{
			name: "code",
			args: args{markdown: "```go\nconst test = \"test\"\n```"},
			want: []notionapi.Block{
				&notionapi.ParagraphBlock{
					Paragraph: notionapi.Paragraph{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkdownToNotionBlocks(tt.args.markdown)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarkdownToNotionBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			j, err := json.Marshal(got)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}

			t.Logf("blocks: %s", j)
		})
	}
}
