// Example: annotation/v3 — defaults, validation, custom rules and full error
// reporting.
package main

import (
	"errors"
	"log"
	"reflect"

	"github.com/tenz-io/gokit/annotation/v3"
)

// Config demonstrates default injection + validation together.
type Config struct {
	Host string `default:"localhost" validate:"required"`
	Port int    `default:"8080"      validate:"required,gt=0,lte=65535"`
	// Mode uses a custom rule registered below.
	Mode string `default:"release" validate:"oneof=debug release test"`
}

// User demonstrates slice validation (dive), email rule, and custom messages.
type User struct {
	Name  string   `validate:"required,non_blank,min_len=2"`
	Email string   `validate:"required,email"`
	Tags  []string `validate:"min_len=1,dive:non_blank"`
	Age   int      `validate:"gte=0,lte=150,msg=Age must be between 0 and 150"`
}

func main() {
	// 1) Defaults: unset fields become the tagged defaults; caller values win.
	cfg := &Config{Host: "override.example"}
	if err := annotation.ApplyDefaults(cfg); err != nil {
		log.Fatalf("apply defaults: %v", err)
	}
	log.Printf("cfg: %+v", cfg)

	if err := annotation.Validate(cfg); err != nil {
		log.Fatalf("validate cfg: %v", err)
	}

	// 2) Validation collects every failure, not just the first.
	user := &User{Name: " a ", Email: "not-an-email", Tags: []string{"go", "  "}, Age: 200}
	err := annotation.Validate(user)
	if err == nil {
		log.Fatal("expected validation errors, got none")
	}
	log.Print("validation errors:")
	var verrs annotation.ValidationErrors
	if errors.As(err, &verrs) {
		for _, e := range verrs {
			log.Printf("  %s", e)
		}
	}

	// 3) After fixing every error, validation passes.
	user2 := &User{Name: "Ada", Email: "ada@example.com", Tags: []string{"go", "rust"}, Age: 30}
	if err := annotation.Validate(user2); err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	log.Print("user2: valid")

	// 4) Register a custom rule at runtime.
	annotation.Register("even", func(_ string, _ reflect.StructField) (annotation.Rule, error) {
		return func(rv reflect.Value) (bool, string) {
			if rv.Kind() == reflect.Int && rv.Int()%2 != 0 {
				return false, "must be even"
			}
			return true, ""
		}, nil
	})

	type Counter struct {
		N int `validate:"required,even"`
	}
	if err := annotation.Validate(&Counter{N: 3}); err != nil {
		log.Printf("custom rule: %v", err)
	}
	if err := annotation.Validate(&Counter{N: 4}); err == nil {
		log.Print("counter: valid (even)")
	}
}
