package notiongo

type EmailProperty struct {
	*property
	Email string `json:"email,omitempty"`
}

// NewEmailProperty creates a new email property.
func NewEmailProperty(name string) *EmailProperty {
	return &EmailProperty{
		property: &property{
			Type: PropertyTypeEmail,
			Name: name,
		},
	}
}

// NewEmailPropertyWithBase creates a new email property.
func NewEmailPropertyWithBase(base *property) *EmailProperty {
	return &EmailProperty{
		property: base,
	}
}

// FromMap converts a map to a EmailProperty.
func (e *EmailProperty) FromMap(m map[string]any) *EmailProperty {
	e.Email = ValueOrDefault(m["email"], "")
	return e
}
