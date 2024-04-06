package notiongo

import (
	"encoding/json"
	"fmt"
)

type ParentType string

const (
	ParentTypeDatabase ParentType = "database_id"
	ParentTypePage     ParentType = "page_id"
)

type Parent struct {
	Type ParentType `json:"type"`
	ID   any        `json:"-"`
}

func NewParent(id string) *Parent {
	return &Parent{
		ID: id,
	}
}

func (p *Parent) WithType(parentType ParentType) *Parent {
	p.Type = parentType
	return p
}

// WithDatabaseType sets the parent type to database.
func (p *Parent) WithDatabaseType() {
	p.Type = ParentTypeDatabase
}

// WithPageType sets the parent type to page.
func (p *Parent) WithPageType() {
	p.Type = ParentTypePage
}

type parentAlias Parent

func (p *Parent) UnmarshalJSON(data []byte) error {
	temp := &parentAlias{}
	if err := json.Unmarshal(data, temp); err != nil {
		return err
	}

	*p = Parent(*temp)

	switch p.Type {
	case ParentTypeDatabase:
		var id struct {
			DatabaseID string `json:"database_id"`
		}
		if err := json.Unmarshal(data, &id); err != nil {
			return err
		}
		p.ID = id.DatabaseID
	case ParentTypePage:
		var id struct {
			PageID string `json:"page_id"`
		}
		if err := json.Unmarshal(data, &id); err != nil {
			return err
		}
		p.ID = id.PageID
	default:
		return fmt.Errorf("unknown type: %v", p.Type)
	}

	return nil
}

func (p *Parent) MarshalJSON() ([]byte, error) {
	switch p.Type {
	case ParentTypeDatabase:
		return json.Marshal(&struct {
			Type       ParentType `json:"type"`
			DatabaseID string     `json:"database_id"`
		}{
			Type:       p.Type,
			DatabaseID: p.ID.(string),
		})
	case ParentTypePage:
		return json.Marshal(&struct {
			Type   ParentType `json:"type"`
			PageID string     `json:"page_id"`
		}{
			Type:   p.Type,
			PageID: p.ID.(string),
		})
	default:
		return nil, fmt.Errorf("unknown type: %v", p.Type)
	}
}
