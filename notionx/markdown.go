package notionx

import (
	"github.com/tenz-io/notionapi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const (
	header1 = 1
	header2 = 2
	header3 = 3

	emphasisItalic = 1
	emphasisBold   = 2
)

// MarkdownToNotionBlocks converts a Markdown string to a slice of Notion blocks.
func MarkdownToNotionBlocks(markdown string) ([]notionapi.Block, error) {
	var blocks []notionapi.Block

	mdParser := goldmark.New()
	source := []byte(markdown)
	doc := mdParser.Parser().Parse(text.NewReader(source))

	var walk func(node ast.Node) error
	walk = func(node ast.Node) error {
		var (
			block notionapi.Block
		)
		switch n := node.(type) {
		case *ast.Document:
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				if err := walk(child); err != nil {
					return err
				}
			}
		case *ast.Heading:
			richTexts := convertToRichTexts(node, source)
			switch n.Level {
			case header1:
				block = &notionapi.Heading1Block{
					BasicBlock: notionapi.BasicBlock{
						Object: notionapi.ObjectTypeBlock,
						Type:   notionapi.BlockTypeHeading1,
					},
					Heading1: notionapi.Heading{
						RichText: richTexts,
					},
				}
			case header2:
				block = &notionapi.Heading2Block{
					BasicBlock: notionapi.BasicBlock{
						Object: notionapi.ObjectTypeBlock,
						Type:   notionapi.BlockTypeHeading2,
					},
					Heading2: notionapi.Heading{
						RichText: richTexts,
					},
				}
			case header3:
				block = &notionapi.Heading3Block{
					BasicBlock: notionapi.BasicBlock{
						Object: notionapi.ObjectTypeBlock,
						Type:   notionapi.BlockTypeHeading3,
					},
					Heading3: notionapi.Heading{
						RichText: richTexts,
					},
				}
			default:
				// For levels > 3, treat as paragraph
				block = &notionapi.ParagraphBlock{
					BasicBlock: notionapi.BasicBlock{
						Object: notionapi.ObjectTypeBlock,
						Type:   notionapi.BlockTypeParagraph,
					},
					Paragraph: notionapi.Paragraph{
						RichText: richTexts,
					},
				}
			}
		case *ast.Paragraph:
			block = &notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: convertToRichTexts(node, source),
				},
			}
		case *ast.List:
			richText := convertToRichTexts(node, source)
			for item := n.FirstChild(); item != nil; item = item.NextSibling() {
				if n.IsOrdered() {
					block = &notionapi.NumberedListItemBlock{
						BasicBlock: notionapi.BasicBlock{
							Object: notionapi.ObjectTypeBlock,
							Type:   notionapi.BlockTypeNumberedListItem,
						},
						NumberedListItem: notionapi.ListItem{
							RichText: richText,
						},
					}
				} else {
					block = &notionapi.BulletedListItemBlock{
						BasicBlock: notionapi.BasicBlock{
							Object: notionapi.ObjectTypeBlock,
							Type:   notionapi.BlockTypeBulletedListItem,
						},
						BulletedListItem: notionapi.ListItem{
							RichText: richText,
						},
					}
				}
			}
		case *ast.Blockquote:
			block = &notionapi.QuoteBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
				},
				Quote: notionapi.Quote{
					RichText: convertToRichTexts(node, source),
				},
			}
		case *ast.FencedCodeBlock:
			block = &notionapi.CodeBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeCode,
				},
				Code: notionapi.Code{
					RichText: convertToRichTexts(node, source),
				},
			}
		default:
			//  treat as paragraph
			block = &notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: convertToRichTexts(node, source),
				},
			}
		}

		if block != nil {
			blocks = append(blocks, block)
		}

		return nil
	}

	if err := walk(doc); err != nil {
		return nil, err
	}

	return blocks, nil
}

// convertToRichTexts converts a Markdown AST node to Notion rich text.
func convertToRichTexts(node ast.Node, source []byte) []notionapi.RichText {
	var richTexts []notionapi.RichText

	var extractText func(n ast.Node)
	extractText = func(n ast.Node) {
		switch n := n.(type) {
		case *ast.Text:
			content := string(n.Segment.Value(source))
			richText := notionapi.RichText{
				Text: &notionapi.Text{
					Content: content,
				},
			}
			richTexts = append(richTexts, richText)
		case *ast.Emphasis:
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				extractText(child)
				if len(richTexts) > 0 {
					if richTexts[len(richTexts)-1].Annotations == nil {
						richTexts[len(richTexts)-1].Annotations = &notionapi.Annotations{}
					}
					if n.Level == emphasisItalic {
						richTexts[len(richTexts)-1].Annotations.Italic = true
					} else if n.Level == emphasisBold {
						richTexts[len(richTexts)-1].Annotations.Bold = true
					}
				}
			}
		case *ast.CodeSpan:
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				extractText(child)
				if len(richTexts) > 0 {
					if richTexts[len(richTexts)-1].Annotations == nil {
						richTexts[len(richTexts)-1].Annotations = &notionapi.Annotations{}
					}
					richTexts[len(richTexts)-1].Annotations.Code = true
				}
			}
		case *ast.FencedCodeBlock:
			content := string(n.Text(source))
			richText := notionapi.RichText{
				Text: &notionapi.Text{
					Content: content,
				},
				Annotations: &notionapi.Annotations{
					Code: true,
				},
			}
			richTexts = append(richTexts, richText)
		default:
			// treat as text
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				extractText(child)
			}
		}
	}

	extractText(node)

	return richTexts
}
