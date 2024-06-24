// template.go.tpl
package main

import (
	"context"
)

func (x *LoginRequest) Validate(ctx context.Context) error {
	return nil
}

func (x *LoginResponse) Validate(ctx context.Context) error {
	return nil
}

func (x *IndexRequest) Validate(ctx context.Context) error {
	return nil
}
