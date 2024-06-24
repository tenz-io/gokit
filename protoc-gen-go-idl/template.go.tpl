// template.go.tpl
package main

import (
	"context"
)

{{range .Messages}}
func (x *{{.Name}}) Validate (ctx context.Context) error {
	return nil
}
{{end}}

