// Package notionx converts Markdown text into Notion API block structures.
package notionx

import (
	"fmt"

	"github.com/tenz-io/notionapi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	header1 = 1
	header2 = 2
	header3 = 3
)

// MarkdownToNotionBlocks parses a Markdown string and returns the corresponding Notion blocks.
// Supports: headings (1-3), paragraphs, bulleted/numbered lists, blockquotes, and fenced code blocks.
// Unsupported node types fall back to paragraph blocks.
func MarkdownToNotionBlocks(markdown string) ([]notionapi.Block, error) {
	if markdown == "" {
		return nil, nil
	}

	source := []byte(markdown)
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	var blocks []notionapi.Block
	err := walkAST(doc, source, &blocks)
	if err != nil {
		return nil, fmt.Errorf("notionx: %w", err)
	}
	return blocks, nil
}

func walkAST(node ast.Node, source []byte, blocks *[]notionapi.Block) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		block := convertNode(child, source)
		if block != nil {
			*blocks = append(*blocks, block)
		}
		if err := walkAST(child, source, blocks); err != nil {
			return err
		}
	}
	return nil
}

func convertNode(n ast.Node, source []byte) notionapi.Block {
	richTexts := extractRichTexts(n, source)

	switch n := n.(type) {
	case *ast.Heading:
		return newHeadingBlock(n.Level, richTexts)
	case *ast.Paragraph:
		return newParagraphBlock(richTexts)
	case *ast.FencedCodeBlock:
		return &notionapi.CodeBlock{
			BasicBlock: notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeCode},
			Code:       notionapi.Code{RichText: richTexts},
		}
	case *ast.Blockquote:
		return &notionapi.QuoteBlock{
			BasicBlock: notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock},
			Quote:      notionapi.Quote{RichText: richTexts},
		}
	case *ast.List:
		return convertListItem(n, source)
	}
	return nil
}

func newHeadingBlock(level int, richTexts []notionapi.RichText) notionapi.Block {
	switch level {
	case header1:
		return &notionapi.Heading1Block{
			BasicBlock: notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeHeading1},
			Heading1:   notionapi.Heading{RichText: richTexts},
		}
	case header2:
		return &notionapi.Heading2Block{
			BasicBlock: notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeHeading2},
			Heading2:   notionapi.Heading{RichText: richTexts},
		}
	case header3:
		return &notionapi.Heading3Block{
			BasicBlock: notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeHeading3},
			Heading3:   notionapi.Heading{RichText: richTexts},
		}
	default:
		return newParagraphBlock(richTexts)
	}
}

func newParagraphBlock(richTexts []notionapi.RichText) notionapi.Block {
	return &notionapi.ParagraphBlock{
		BasicBlock: notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeParagraph},
		Paragraph:  notionapi.Paragraph{RichText: richTexts},
	}
}

func convertListItem(n *ast.List, source []byte) notionapi.Block {
	richText := extractRichTexts(n, source)
	if n.IsOrdered() {
		return &notionapi.NumberedListItemBlock{
			BasicBlock:       notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeNumberedListItem},
			NumberedListItem: notionapi.ListItem{RichText: richText},
		}
	}
	return &notionapi.BulletedListItemBlock{
		BasicBlock:      notionapi.BasicBlock{Object: notionapi.ObjectTypeBlock, Type: notionapi.BlockTypeBulletedListItem},
		BulletedListItem: notionapi.ListItem{RichText: richText},
	}
}

func extractRichTexts(node ast.Node, source []byte) []notionapi.RichText {
	var richTexts []notionapi.RichText
	collectText(node, source, &richTexts)
	return richTexts
}

func collectText(n ast.Node, source []byte, richTexts *[]notionapi.RichText) {
	switch n := n.(type) {
	case *ast.Text:
		*richTexts = append(*richTexts, notionapi.RichText{
			Text: &notionapi.Text{Content: string(n.Segment.Value(source))},
		})
	case *ast.Emphasis:
		before := len(*richTexts)
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, richTexts)
		}
		for i := before; i < len(*richTexts); i++ {
			if (*richTexts)[i].Annotations == nil {
				(*richTexts)[i].Annotations = &notionapi.Annotations{}
			}
			switch n.Level {
			case 1:
				(*richTexts)[i].Annotations.Italic = true
			case 2:
				(*richTexts)[i].Annotations.Bold = true
			}
		}
	case *ast.CodeSpan:
		before := len(*richTexts)
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, richTexts)
		}
		for i := before; i < len(*richTexts); i++ {
			if (*richTexts)[i].Annotations == nil {
				(*richTexts)[i].Annotations = &notionapi.Annotations{}
			}
			(*richTexts)[i].Annotations.Code = true
		}
	default:
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, richTexts)
		}
	}
}
