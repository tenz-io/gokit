package notiongo

// "Status": {
//            "select": {
//                "id": "8c4a056e-6709-4dd1-ba58-d34d9480855a",
//                "name": "Ready to Start",
//                "color": "yellow"
//            }
//        },

type SelectOption struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// FromMap sets the select option from a map.
func (s *SelectOption) FromMap(m map[string]any) *SelectOption {
	s.ID = ValueOrDefault(m["id"], "")
	s.Name = ValueOrDefault(m["name"], "")
	s.Color = ValueOrDefault(m["color"], "")
	return s
}

type SelectProperty struct {
	*property
	SelectOptions []*SelectOption `json:"options,omitempty"`
	SelectOption  *SelectOption   `json:"select,omitempty"`
}

// NewSelectPropertyWithBase creates a new select property.
func NewSelectPropertyWithBase(base *property) *SelectProperty {
	return &SelectProperty{
		property: base,
	}
}

func NewSelectOption(name string) *SelectOption {
	return &SelectOption{
		Name: name,
	}
}

func (s *SelectOption) WithID(id string) *SelectOption {
	s.ID = id
	return s
}

func (s *SelectOption) WithColor(color string) *SelectOption {
	s.Color = color
	return s
}

func NewSelectProperty(name string) *SelectProperty {
	return &SelectProperty{
		property: &property{
			Type: PropertyTypeSelect,
			Name: name,
		},
	}
}

func NewMultiSelectProperty(name string) *SelectProperty {
	return &SelectProperty{
		property: &property{
			Type: PropertyTypeMultiSelect,
			Name: name,
		},
	}
}

func (s *SelectProperty) WithSelectOptions(options []*SelectOption) *SelectProperty {
	s.SelectOptions = options
	return s
}

func (s *SelectProperty) WithSelectOption(option *SelectOption) *SelectProperty {
	s.SelectOption = option
	return s
}

// FromMap sets the select property from a map.
func (s *SelectProperty) FromMap(m map[string]any) *SelectProperty {
	s.property = (&property{}).FromMap(m)

	key := ifThen(s.property.Type == PropertyTypeSelect, "select", "multi_select")

	if options, ok := m["options"].([]any); ok {
		for _, option := range options {
			optionMap := option.(map[string]any)
			s.SelectOptions = append(s.SelectOptions, (&SelectOption{}).FromMap(optionMap))
		}
	}
	if selectOption, ok := m[key].(map[string]any); ok {
		s.SelectOption = (&SelectOption{}).FromMap(selectOption)
	}
	return s
}
