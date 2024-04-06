package notiongo

import (
	"encoding/json"
	"fmt"
)

type RichTextType string

const (
	RichTextTypeText RichTextType = "text"
)

type Text struct {
	Content string `json:"content,omitempty"`
	Link    string `json:"link,omitempty"`
}

// NewTextWithContent creates a new text.
func NewTextWithContent(content string) *Text {
	return &Text{
		Content: content,
	}
}

// NewText creates a new text.
func NewText() *Text {
	return &Text{}
}

// WithContent sets the content of the text.
func (t *Text) WithContent(content string) *Text {
	t.Content = content
	return t
}

// WithLink sets the link of the text.
func (t *Text) WithLink(link string) *Text {
	t.Link = link
	return t
}

func (t *Text) FromMap(text map[string]any) *Text {
	t.Content = ValueOrDefault(text["content"], "")
	t.Link = ValueOrDefault(text["link"], "")
	return t
}

type RichText struct {
	Type        RichTextType `json:"type,omitempty"`
	Text        *Text        `json:"text,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
	PlainText   string       `json:"plain_text,omitempty"`
	Href        string       `json:"href,omitempty"`
}

// NewRichText creates a new rich text.
func NewRichText() *RichText {
	return &RichText{}
}

func NewRichTextWithContent(content string) *RichText {
	return &RichText{
		Type: RichTextTypeText,
		Text: NewTextWithContent(content),
	}
}

// WithType sets the type of the rich text.
func (r *RichText) WithType(richTextType RichTextType) *RichText {
	r.Type = richTextType
	return r
}

// WithText sets the text of the rich text.
func (r *RichText) WithText(text *Text) *RichText {
	r.Text = text
	return r
}

// WithAnnotations sets the annotations of the rich text.
func (r *RichText) WithAnnotations(annotations *Annotations) *RichText {
	r.Annotations = annotations
	return r
}

// WithPlainText sets the plain text of the rich text.
func (r *RichText) WithPlainText(plainText string) *RichText {
	r.PlainText = plainText
	return r
}

// WithHref sets the href of the rich text.
func (r *RichText) WithHref(href string) *RichText {
	r.Href = href
	return r
}

// WithLink sets the link of the rich text.
func (r *RichText) WithLink(link string) *RichText {
	if r.Text == nil {
		r.Text = NewText()
	}
	r.Text.Link = link
	return r
}

// WithContent sets the content of the rich text.
func (r *RichText) WithContent(content string) *RichText {
	if r.Text == nil {
		r.Text = NewText()
	}
	r.Text.Content = content
	return r
}

type RichTextProperty struct {
	*property
	RichText []*RichText `json:"rich_text,omitempty"`
}

// NewRichTextPropertyWithBase creates a new rich text property.
func NewRichTextPropertyWithBase(base *property) *RichTextProperty {
	return &RichTextProperty{
		property: base,
	}
}

// WithRichText sets the rich text of the rich text property.
func (r *RichTextProperty) WithRichText(richText []*RichText) *RichTextProperty {
	r.RichText = richText
	return r
}

// MarshalJSON marshals the rich text property into JSON.
func (r *RichTextProperty) MarshalJSON() ([]byte, error) {
	propertyMap := make(map[string]any)
	if r.property != nil {
		propertyMap["id"] = r.property.ID
		propertyMap["type"] = r.property.Type
		propertyMap["name"] = r.property.Name
	}

	texts := make([]map[string]any, 0)
	for _, rt := range r.RichText {
		textMap := make(map[string]any)
		textMap["type"] = rt.Type
		textMap["text"] = rt.Text
		if rt.Annotations != nil {
			textMap["annotations"] = rt.Annotations
		}
		if rt.PlainText != "" {
			textMap["plain_text"] = rt.PlainText
		}
		if rt.Href != "" {
			textMap["href"] = rt.Href
		}
		texts = append(texts, textMap)
	}

	if len(texts) > 0 {
		propertyMap["rich_text"] = texts
	}

	return json.Marshal(propertyMap)
}

// FromMap converts a map to a rich text property.
func (r *RichTextProperty) FromMap(m map[string]any) *RichTextProperty {
	richTextsVal := ValueOrDefault(m["rich_text"], []any{})
	if len(richTextsVal) == 0 {
		return r
	}

	r.RichText = make([]*RichText, 0)
	for _, richTextVal := range richTextsVal {
		richText := ValueOrDefault(richTextVal, map[string]any{})
		if len(richText) == 0 {
			continue
		}
		rt := &RichText{}
		rt.Type = RichTextType(fmt.Sprint(richText["type"]))
		if textMap := ValueOrDefault(richText["text"], map[string]any{}); len(textMap) > 0 {
			rt.Text = (&Text{}).FromMap(textMap)
		}
		if annotationsMap := ValueOrDefault(richText["annotations"], map[string]any{}); len(annotationsMap) > 0 {
			rt.Annotations = (&Annotations{}).FromMap(annotationsMap)
		}

		rt.PlainText = ValueOrDefault(richText["plain_text"], "")
		rt.Href = ValueOrDefault(richText["href"], "")
		r.RichText = append(r.RichText, rt)
	}

	return r
}
