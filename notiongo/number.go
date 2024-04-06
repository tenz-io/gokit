package notiongo

type Number struct {
	Format string `json:"format,omitempty"`
}

type NumberProperty struct {
	*property
	Number *Number `json:"number,omitempty"`
}

func NewNumberProperty(name string) *NumberProperty {
	return &NumberProperty{
		property: &property{
			Type: PropertyTypeNumber,
			Name: name,
		},
	}
}

func NewNumberPropertyWithBase(base *property) *NumberProperty {
	return &NumberProperty{
		property: base,
	}
}

// WithNumber sets the number property.
func (n *NumberProperty) WithNumber(number *Number) *NumberProperty {
	n.Number = number
	return n
}

func (n *NumberProperty) FromMap(m map[string]any) *NumberProperty {
	n.Number = &Number{
		Format: ValueOrDefault(m["format"], ""),
	}
	return n
}
