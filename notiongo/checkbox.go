package notiongo

type CheckBoxProperty struct {
	*property
	CheckBox bool `json:"checkbox,omitempty"`
}

// NewCheckBoxProperty creates a new checkbox property.
func NewCheckBoxProperty(name string) *CheckBoxProperty {
	return &CheckBoxProperty{
		property: &property{
			Type: PropertyTypeCheckbox,
			Name: name,
		},
	}
}

// NewCheckBoxPropertyWithBase creates a new checkbox property.
func NewCheckBoxPropertyWithBase(base *property) *CheckBoxProperty {
	return &CheckBoxProperty{
		property: base,
	}
}

// FromMap converts a map to a CheckBoxProperty.
func (c *CheckBoxProperty) FromMap(m map[string]any) *CheckBoxProperty {
	c.CheckBox = ValueOrDefault(m["checkbox"], false)
	return c
}
