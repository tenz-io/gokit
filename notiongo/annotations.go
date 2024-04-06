package notiongo

type Annotations struct {
	Bold          bool   `json:"bold,omitempty"`
	Italic        bool   `json:"italic,omitempty"`
	Strikethrough bool   `json:"strikethrough,omitempty"`
	Underline     bool   `json:"underline,omitempty"`
	Code          bool   `json:"code,omitempty"`
	Color         string `json:"color,omitempty"`
}

// WithBold sets the bold of the annotations.
func (a *Annotations) WithBold() *Annotations {
	a.Bold = true
	return a
}

// WithItalic sets the italic of the annotations.
func (a *Annotations) WithItalic() *Annotations {
	a.Italic = true
	return a
}

// WithStrikethrough sets the strikethrough of the annotations.
func (a *Annotations) WithStrikethrough() *Annotations {
	a.Strikethrough = true
	return a
}

// WithUnderline sets the underline of the annotations.
func (a *Annotations) WithUnderline() *Annotations {
	a.Underline = true
	return a
}

// WithCode sets the code of the annotations.
func (a *Annotations) WithCode() *Annotations {
	a.Code = true
	return a
}

// WithColor sets the color of the annotations.
func (a *Annotations) WithColor(color string) *Annotations {
	a.Color = color
	return a
}

func (a *Annotations) FromMap(annotations map[string]any) *Annotations {
	a.Bold = ValueOrDefault(annotations["bold"], false)
	a.Italic = ValueOrDefault(annotations["italic"], false)
	a.Strikethrough = ValueOrDefault(annotations["strikethrough"], false)
	a.Underline = ValueOrDefault(annotations["underline"], false)
	a.Code = ValueOrDefault(annotations["code"], false)
	a.Color = ValueOrDefault(annotations["color"], "")
	return a
}

// NewAnnotations creates a new annotations.
func NewAnnotations() *Annotations {
	return &Annotations{}
}
