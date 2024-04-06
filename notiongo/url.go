package notiongo

type UrlProperty struct {
	*property
	Url string `json:"url,omitempty"`
}

// NewUrlProperty creates a new url property.
func NewUrlProperty(name string) *UrlProperty {
	return &UrlProperty{
		property: &property{
			Type: PropertyTypeURL,
			Name: name,
		},
	}
}

// NewUrlPropertyWithBase creates a new url property.
func NewUrlPropertyWithBase(base *property) *UrlProperty {
	return &UrlProperty{
		property: base,
	}
}

// FromMap converts a map to a UrlProperty.
func (u *UrlProperty) FromMap(m map[string]any) *UrlProperty {
	u.Url = ValueOrDefault(m["url"], "")
	return u
}
