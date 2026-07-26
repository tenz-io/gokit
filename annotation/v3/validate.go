package annotation

import (
	"errors"
	"reflect"
)

// Validate runs every rule on every field of the struct behind ptr and returns
// all failures. ptr must be a non-nil pointer to a struct. nil means no
// failures.
func Validate(ptr any) error {
	p, err := PlanFor(ptr)
	if err != nil {
		return err
	}
	root := reflect.ValueOf(ptr).Elem()
	var errs ValidationErrors
	validateField(p.fields, root, &errs)
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// QuickValidate runs the same rules as Validate but returns on the first
// failure encountered, as a single FieldError. nil means no failures. Use it
// when you only need to know whether the input is valid (and one reason why
// not), without paying to collect every error.
func QuickValidate(ptr any) error {
	p, err := PlanFor(ptr)
	if err != nil {
		return err
	}
	root := reflect.ValueOf(ptr).Elem()
	if fe, ok := validateFieldFirst(p.fields, root); ok {
		return fe
	}
	return nil
}

func validateField(fields []*Field, parent reflect.Value, errs *ValidationErrors) {
	for _, f := range fields {
		fv := parent.FieldByIndex(f.Index)
		// Run this field's own rules against the raw (possibly nil) value so
		// "required" can catch a nil nested pointer.
		for _, r := range f.rules {
			ok, msg := r.run(fv)
			if ok {
				continue
			}
			fe := FieldError{Field: f.Path, Rule: r.name, Param: r.param}
			if r.msg != "" {
				fe.Msg = r.msg
			} else {
				fe.Msg = msg
			}
			*errs = append(*errs, fe)
		}

		// Recurse into nested structs only when present. We do NOT instantiate
		// nil nested pointers here (that would mask a "required" failure);
		// ApplyDefaults is responsible for wiring optional defaults.
		if len(f.children) == 0 {
			continue
		}
		cur := fv
		if cur.Kind() == reflect.Ptr {
			if cur.IsNil() {
				continue
			}
			cur = cur.Elem()
		}
		if cur.Kind() == reflect.Struct && cur.IsValid() && cur.CanInterface() {
			validateField(f.children, cur, errs)
		}
	}
}

// validateFieldFirst is the short-circuit twin of validateField: it stops at
// the first failing rule and returns it as a single FieldError.
func validateFieldFirst(fields []*Field, parent reflect.Value) (FieldError, bool) {
	for _, f := range fields {
		fv := parent.FieldByIndex(f.Index)
		for _, r := range f.rules {
			if ok, msg := r.run(fv); !ok {
				fe := FieldError{Field: f.Path, Rule: r.name, Param: r.param}
				if r.msg != "" {
					fe.Msg = r.msg
				} else {
					fe.Msg = msg
				}
				return fe, true
			}
		}
		if len(f.children) == 0 {
			continue
		}
		cur := fv
		if cur.Kind() == reflect.Ptr {
			if cur.IsNil() {
				continue
			}
			cur = cur.Elem()
		}
		if cur.Kind() == reflect.Struct && cur.IsValid() && cur.CanInterface() {
			if fe, ok := validateFieldFirst(f.children, cur); ok {
				return fe, true
			}
		}
	}
	return FieldError{}, false
}

// AsErrors extracts the collected failures from an error returned by Validate.
// ok is false when the error is not from the validator.
func AsErrors(err error) (errs ValidationErrors, ok bool) {
	if err == nil {
		return nil, false
	}
	if ve, ok := err.(ValidationErrors); ok {
		return ve, true
	}
	var target ValidationErrors
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}
