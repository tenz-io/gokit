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
{{if .Int}}
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
{{- end}}

{{if .Str}}
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
    if !genproto.ContainsString(x.Get{{.FieldName}}(), []string{ {{range $i, $v := .Str.In}}{{if $i}}, {{end}}"{{$v}}"{{end}} }) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is invalid",
        }
    }
    {{- end}}
    {{if .Str.NotIn}}
    if genproto.ContainsString(x.Get{{.FieldName}}(), []string{ {{range $i, $v := .Str.NotIn}}{{if $i}}, {{end}}"{{$v}}"{{end}} }) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is invalid",
        }
    }
    {{- end}}
    {{if .Str.Pattern}}
    if !{{.Str.Pattern}}.MatchString(x.Get{{.FieldName}}()) {
        return &genproto.ValidationError{
            Key: "{{.FieldName}}",
            Message: "is invalid",
        }
    }
    {{- end}}
{{- end}}

	return nil
}
{{end}}


