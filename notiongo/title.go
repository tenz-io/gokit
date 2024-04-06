package notiongo

import "encoding/json"

type TitleType string

const (
	TitleTypeText TitleType = "text"
)

type Title struct {
	Type        TitleType    `json:"type,omitempty"`
	Text        *Text        `json:"text,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
	PlainText   string       `json:"plain_text,omitempty"`
	Href        string       `json:"href,omitempty"`
}

func NewTitle() *Title {
	return &Title{}

}

func NewTitleWithContent(content string) *Title {
	return &Title{
		Type: TitleTypeText,
		Text: &Text{
			Content: content,
		},
	}
}

// WithType sets the type of the title.
func (t *Title) WithType(titleType TitleType) *Title {
	t.Type = titleType
	return t
}

// WithText sets the text of the title.
func (t *Title) WithText(text *Text) *Title {
	t.Text = text
	return t
}

// WithContent sets the content of the title.
func (t *Title) WithContent(content string) *Title {
	if t.Text == nil {
		t.Text = &Text{}
	}
	t.Text.Content = content
	return t
}

// WithLink sets the link of the title.
func (t *Title) WithLink(link string) *Title {
	if t.Text == nil {
		t.Text = &Text{}
	}
	t.Text.Link = link
	return t
}

func (t *Title) FromMap(titleMap map[string]any) *Title {
	t.Type = TitleType(ValueOrDefault(titleMap["type"], ""))
	t.PlainText = ValueOrDefault(titleMap["plain_text"], "")
	t.Href = ValueOrDefault(titleMap["href"], "")

	if textMap := ValueOrDefault(titleMap["text"], map[string]any{}); len(textMap) > 0 {
		t.Text = (&Text{}).FromMap(textMap)
	}

	if annotationsMap := ValueOrDefault(titleMap["annotations"], map[string]any{}); len(annotationsMap) > 0 {
		t.Annotations = (&Annotations{}).FromMap(annotationsMap)
	}

	return t
}

type TitleProperty struct {
	*property
	Title []*Title `json:"title,omitempty"`
}

func NewTitlePropertyWithBase(base *property) *TitleProperty {
	return &TitleProperty{
		property: base,
	}
}

func NewTitleProperty(name string) *TitleProperty {
	return &TitleProperty{
		property: &property{
			Type: PropertyTypeTitle,
			Name: name,
		},
	}
}

func (t *TitleProperty) WithTitle(title []*Title) *TitleProperty {
	t.Title = title
	return t
}

func (t *TitleProperty) AddTitle(title *Title) *TitleProperty {
	t.Title = append(t.Title, title)
	return t
}

func (t *TitleProperty) MarshalJSON() ([]byte, error) {
	propertyMap := make(map[string]any)
	if t.property != nil {
		propertyMap["id"] = t.property.ID
		propertyMap["type"] = t.property.Type
		propertyMap["name"] = t.property.Name
	}

	properties := make([]map[string]any, 0)
	for _, title := range t.Title {
		titleMap := make(map[string]any)
		titleMap["type"] = title.Type
		titleMap["text"] = title.Text
		properties = append(properties, titleMap)
	}
	if len(properties) > 0 {
		propertyMap["title"] = properties
	}
	return json.Marshal(propertyMap)
}

func (t *TitleProperty) FromMap(m map[string]any) *TitleProperty {
	titleVals := ValueOrDefault(m["title"], []any{})
	if len(titleVals) == 0 {
		return t
	}
	for _, titleVal := range titleVals {
		titleMap := ValueOrDefault(titleVal, map[string]any{})
		if len(titleMap) == 0 {
			continue
		}

		title := (&Title{}).FromMap(titleMap)
		t.Title = append(t.Title, title)
	}
	return t
}
