func (x *{{.MessageName}}) Validate() error {
{{range .Fields}}
	if err := x.validate{{.FieldName}}(); err != nil {
        return err
    }
{{end}}
	return nil
}

{{range .Fields}}
func (x *{{.MessageName}}) validate{{.FieldName}}() error {
{{if .IsMessage}}
    return x.{{.FieldName}}.Validate()
{{- end}}

{{if .Int}}
    {{if .Int.Default}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        genproto.SetValue(x.{{.FieldName}}, {{.Int.Default}})
    }
    {{- end}}

    {{if .Int.Required}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is required",
        }
    }
    {{- end}}
    {{if .Int.Gt}}
    if x.Get{{.FieldName}}() <= {{.Int.Gt}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be greater than %d", {{.Int.Gt}}),
        }
    }
    {{- end}}
    {{if .Int.Gte}}
    if x.Get{{.FieldName}}() < {{.Int.Gte}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be greater than or equal to %d", {{.Int.Gte}}),
        }
    }
    {{- end}}
    {{if .Int.Lt}}
    if x.Get{{.FieldName}}() >= {{.Int.Lt}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be less than %d", {{.Int.Lt}}),
        }
    }
    {{- end}}
    {{if .Int.Lte}}
    if x.Get{{.FieldName}}() > {{.Int.Lte}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be less than or equal to %d", {{.Int.Lte}}),
        }
    }
    {{- end}}
    {{if .Int.Eq}}
    if x.Get{{.FieldName}}() != {{.Int.Eq}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be equal to %d", {{.Int.Eq}}),
        }
    }
    {{- end}}
    {{if .Int.Ne}}
    if x.Get{{.FieldName}}() == {{.Int.Ne}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must not be equal to %d", {{.Int.Ne}}),
        }
    }
    {{- end}}
    {{if .Int.In}}
    inList := []int64{ {{range $i, $v := .Int.In}}{{if $i}}, {{end}}{{$v}}{{end}} }
    if !genproto.IntIn(x.Get{{.FieldName}}(), inList) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be one of %v", inList),
        }
    }
    {{- end}}
    {{if .Int.NotIn}}
    notInList := []int64{ {{range $i, $v := .Int.NotIn}}{{if $i}}, {{end}}{{$v}}{{end}} }
    if genproto.IntIn(x.Get{{.FieldName}}(), notInList) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must not be one of %v", notInList),
        }
    }
    {{- end}}
{{- end}}

{{if .Str}}
    {{if .Str.Default}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        genproto.SetValue(x.{{.FieldName}}, "{{.Str.Default}}")
    }
    {{- end}}

    {{if .Str.Required}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is required",
        }
    }
    {{- end}}
    {{if .Str.NotBlank}}
    if strings.TrimSpace(x.Get{{.FieldName}}()) == "" {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "can not be blank",
        }
    }
    {{- end}}
    {{if .Str.MinLen}}
    if len(x.Get{{.FieldName}}()) < {{.Str.MinLen}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be at least %d characters long", {{.Str.MinLen}}),
        }
    }
    {{- end}}
    {{if .Str.MaxLen}}
    if len(x.Get{{.FieldName}}()) > {{.Str.MaxLen}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be at most %d characters long", {{.Str.MaxLen}}),
        }
    }
    {{- end}}
    {{if .Str.In}}
    inList := []string{ {{range $i, $v := .Str.In}}{{if $i}}, {{end}}"{{$v}}"{{end}} }
    if !genproto.StringIn(x.Get{{.FieldName}}(), inList) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be one of %v", inList),
        }
    }
    {{- end}}
    {{if .Str.NotIn}}
    notInList := []string{ {{range $i, $v := .Str.NotIn}}{{if $i}}, {{end}}"{{$v}}"{{end}} }
    if genproto.StringIn(x.Get{{.FieldName}}(), notInList) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must not be one of %v", notInList),
        }
    }
    {{- end}}
    {{if .Str.Pattern}}
    if !genproto.StringMatches(x.Get{{.FieldName}}(), "{{.Str.Pattern}}") {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is invalid",
        }
    }
    {{- end}}
{{- end}}

{{if .Float}}
    {{if .Float.Default}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        genproto.SetValue(x.{{.FieldName}}, {{.Float.Default}})
    }
    {{- end}}

    {{if .Float.Required}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is required",
        }
    }
    {{- end}}
    {{if .Float.Gt}}
    if x.Get{{.FieldName}}() <= {{.Float.Gt}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be greater than %f", {{.Float.Gt}}),
        }
    }
    {{- end}}
    {{if .Float.Gte}}
    if x.Get{{.FieldName}}() < {{.Float.Gte}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be greater than or equal to %f", {{.Float.Gte}}),
        }
    }
    {{- end}}
    {{if .Float.Lt}}
    if x.Get{{.FieldName}}() >= {{.Float.Lt}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be less than %f", {{.Float.Lt}}),
        }
    }
    {{- end}}
    {{if .Float.Lte}}
    if x.Get{{.FieldName}}() > {{.Float.Lte}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be less than or equal to %f", {{.Float.Lte}}),
        }
    }
    {{- end}}
 {{- end}}

 {{if .Bytes}}
    {{if .Bytes.Required}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is required",
        }
    }
    {{- end}}
    {{if .Bytes.MinLen}}
    if len(x.Get{{.FieldName}}()) < {{.Bytes.MinLen}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be at least %d bytes long", {{.Bytes.MinLen}}),
        }
    }
    {{- end}}
    {{if .Bytes.MaxLen}}
    if len(x.Get{{.FieldName}}()) > {{.Bytes.MaxLen}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must be at most %d bytes long", {{.Bytes.MaxLen}}),
        }
    }
    {{- end}}
{{- end}}

{{if .Array}}
    {{if .Array.Required}}
    if genproto.IsNilOrEmpty(x.{{.FieldName}}) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is required",
        }
    }
    {{- end}}
    {{if .Array.MinItems}}
    if len(x.Get{{.FieldName}}()) < {{.Array.MinItems}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must have at least %d items", {{.Array.MinItems}}),
        }
    }
    {{- end}}
    {{if .Array.MaxItems}}
    if len(x.Get{{.FieldName}}()) > {{.Array.MaxItems}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must have at most %d items", {{.Array.MaxItems}}),
        }
    }
    {{- end}}
    {{if .Array.Len}}
    if len(x.Get{{.FieldName}}()) != {{.Array.Len}} {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: fmt.Sprintf("must have exactly %d items", {{.Array.Len}}),
        }
    }
    {{- end}}
{{- end}}

	return nil
}
{{end}}


