package notiongo

import (
	"encoding/json"
	"fmt"
)

type PropertyType string

const (
	PropertyTypeTitle          PropertyType = "title"
	PropertyTypeRichText       PropertyType = "rich_text"
	PropertyTypeNumber         PropertyType = "number"
	PropertyTypeSelect         PropertyType = "select"
	PropertyTypeMultiSelect    PropertyType = "multi_select"
	PropertyTypeDate           PropertyType = "date"
	PropertyTypeCheckbox       PropertyType = "checkbox"
	PropertyTypeURL            PropertyType = "url"
	PropertyTypeEmail          PropertyType = "email"
	PropertyTypeCreatedTime    PropertyType = "created_time"
	PropertyTypeLastEditedTime PropertyType = "last_edited_time"
	PropertyTypeFile           PropertyType = "file"
	PropertyTypePeople         PropertyType = "people"
)

type Properties []Property

type Property interface {
	GetID() string
	GetType() PropertyType
	GetName() string
	WithID(id string) Property
	WithName(name string) Property
	WithType(propertyType PropertyType) Property
}

func NewProperty(title string, propertyType PropertyType) Property {
	return &property{
		Type: propertyType,
		Name: title,
	}
}

type property struct {
	ID   string       `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	Name string       `json:"name,omitempty"`
}

func (p *property) GetID() string {
	return p.ID
}

func (p *property) GetType() PropertyType {
	return p.Type
}

func (p *property) GetName() string {
	return p.Name
}

// WithID sets the ID of the property.
func (p *property) WithID(id string) Property {
	p.ID = id
	return p
}

// WithName sets the title of the property.
func (p *property) WithName(name string) Property {
	p.Name = name
	return p
}

// WithType sets the type of the property.
func (p *property) WithType(propertyType PropertyType) Property {
	p.Type = propertyType
	return p
}

func (p *property) FromMap(m map[string]any) *property {
	p.ID = ValueOrDefault(m["id"], "")
	p.Type = PropertyType(ValueOrDefault(m["type"], ""))
	p.Name = ValueOrDefault(m["name"], "")
	return p

}

// MarshalJSON marshals the properties into JSON.
func (pps *Properties) MarshalJSON() ([]byte, error) {
	propertyMap := make(map[string]any)
	for _, p := range *pps {
		switch p.GetType() {
		case PropertyTypeTitle:
			if titleProperty, ok := p.(*TitleProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"title": titleProperty.Title,
				}
			}
		case PropertyTypeRichText:
			if richTextProperty, ok := p.(*RichTextProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"rich_text": richTextProperty.RichText,
				}
			}
		case PropertyTypeNumber:
			if numberProperty, ok := p.(*NumberProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"number": numberProperty.Number,
				}
			}
		case PropertyTypeCheckbox:
			if checkboxProperty, ok := p.(*CheckBoxProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"checkbox": checkboxProperty.CheckBox,
				}
			}
		case PropertyTypeDate:
			if dateProperty, ok := p.(*DateProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"date": dateProperty.Date,
				}
			}
		case PropertyTypeCreatedTime:
			propertyMap[p.GetName()] = map[string]any{
				"created_time": map[string]any{},
			}
		case PropertyTypeLastEditedTime:
			propertyMap[p.GetName()] = map[string]any{
				"last_edited_time": map[string]any{},
			}
		case PropertyTypeURL:
			if urlProperty, ok := p.(*UrlProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"url": urlProperty.Url,
				}
			}
		case PropertyTypeEmail:
			if emailProperty, ok := p.(*EmailProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"email": emailProperty.Email,
				}
			}
		case PropertyTypeSelect:
			if selectProperty, ok := p.(*SelectProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"select": map[string]any{
						"options": selectProperty.SelectOptions,
					},
				}
			}
		case PropertyTypeMultiSelect:
			if selectProperty, ok := p.(*SelectProperty); ok {
				propertyMap[p.GetName()] = map[string]any{
					"multi_select": map[string]any{
						"options": selectProperty.SelectOptions,
					},
				}
			}
		case PropertyTypeFile:
			propertyMap[p.GetName()] = map[string]any{
				"files": map[string]any{},
			}
		case PropertyTypePeople:
			propertyMap[p.GetName()] = map[string]any{
				"people": map[string]any{},
			}
		}
	}

	return json.Marshal(propertyMap)
}

// UnmarshalJSON unmarshal the JSON data into properties.
func (pps *Properties) UnmarshalJSON(data []byte) error {
	var (
		propertyMap map[string]any
		props       Properties
	)

	// unmarshal the JSON data into a map.
	if err := json.Unmarshal(data, &propertyMap); err != nil {
		return err
	}
	// convert the map into properties.
	for name, val := range propertyMap {
		mapVal := ValueOrDefault(val, map[string]any{})
		if len(mapVal) == 0 {
			continue
		}

		id := ValueOrDefault(mapVal["id"], "")
		if nameVal := ValueOrDefault(mapVal["name"], ""); nameVal != "" {
			name = nameVal
		}
		typ := ValueOrDefault(mapVal["type"], "")
		if typ == "" {
			return fmt.Errorf("invalid property type: %v", typ)
		}

		base := &property{
			ID:   id,
			Type: PropertyType(typ),
			Name: name,
		}

		switch propType := PropertyType(typ); propType {
		case PropertyTypeFile, PropertyTypePeople:
			props = append(props, base)
		case PropertyTypeCheckbox:
			prop := NewCheckBoxPropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeURL:
			prop := NewUrlPropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeEmail:
			prop := NewEmailPropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeTitle:
			prop := NewTitlePropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeRichText:
			prop := NewRichTextPropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeDate, PropertyTypeCreatedTime, PropertyTypeLastEditedTime:
			prop := NewDatePropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeNumber:
			prop := NewNumberPropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		case PropertyTypeSelect, PropertyTypeMultiSelect:
			prop := NewSelectPropertyWithBase(base).FromMap(mapVal)
			props = append(props, prop)
		}

		*pps = props
	}

	return nil
}
