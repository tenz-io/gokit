package notiongo

type CreateDatabaseRequest struct {
	Parent     *Parent    `json:"parent,omitempty"`
	Title      []*Title   `json:"title,omitempty"`
	Properties Properties `json:"properties,omitempty"`
}

type CreateDatabaseResponse struct {
}

// NewCreateDatabaseRequest creates a new CreateDatabaseRequest.
func NewCreateDatabaseRequest(parent *Parent, title ...*Title) *CreateDatabaseRequest {
	return &CreateDatabaseRequest{
		Parent:     parent,
		Title:      title,
		Properties: make(Properties, 0),
	}
}

// WithParent sets the parent of the create database request.
func (c *CreateDatabaseRequest) WithParent(parent *Parent) *CreateDatabaseRequest {
	c.Parent = parent
	return c
}

// WithTitle sets the title of the create database request.
func (c *CreateDatabaseRequest) WithTitle(title ...*Title) *CreateDatabaseRequest {
	c.Title = title
	return c
}

// WithProperties sets the properties of the create database request.
func (c *CreateDatabaseRequest) WithProperties(properties Properties) *CreateDatabaseRequest {
	c.Properties = properties
	return c
}

// AddProperty adds a property to the create database request.
func (c *CreateDatabaseRequest) AddProperty(property *property) *CreateDatabaseRequest {
	if c.Properties == nil {
		c.Properties = make(Properties, 0)
	}
	c.Properties = append(c.Properties, property)
	return c
}
