{{range .Messages}}
func (x *{{.Name}}) Validate(_ context.Context) error {
	return genproto.ValidateMessage(x.ValidateRule(), x)
}

func (x *{{.Name}}) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{
        {{range .Fields}}
        "{{.Name}}": &idl.Field{
            {{- if .IntField}}
            Int: &idl.IntField{
                {{- if .IntField.Default}}
                Default: proto.Int64({{.IntField.Default}}),
                {{- end}}
                {{- if .IntField.Required}}
                Required: proto.Bool({{.IntField.Required}}),
                {{- end}}
                {{- if .IntField.Gt}}
                Gt: proto.Int64({{.IntField.Gt}}),
                {{- end}}
                {{- if .IntField.Gte}}
                {{- end}}
                {{- if .IntField.Gte}}
                Gte: proto.Int64({{.IntField.Gte}}),
                {{- end}}
                {{- if .IntField.Lt}}
                Lt: proto.Int64({{.IntField.Lt}}),
                {{- end}}
                {{- if .IntField.Lte}}
                Lte: proto.Int64({{.IntField.Lte}}),
                {{- end}}
                {{- if .IntField.Eq}}
                Eq: proto.Int64({{.IntField.Eq}}),
                {{- end}}
                {{- if .IntField.Ne}}
                Ne: proto.Int64({{.IntField.Ne}}),
                {{- end}}
                {{- if .IntField.In}}
                In: []int64{
                {{range .IntField.In}}
                {{.}}, {{end}}
                },
                {{- end}}
            },
            {{- end}}
            {{- if .StringField}}
            Str: &idl.StringField{
                {{- if .StringField.Default}}
                Default: proto.String("{{.StringField.Default}}"),
                {{- end}}
                {{- if .StringField.Required}}
                Required: proto.Bool({{.StringField.Required}}),
                {{- end}}
                {{- if .StringField.NotBlank}}
                NotBlank: proto.Bool({{.StringField.NotBlank}}),
                {{- end}}
                {{- if .StringField.MinLen}}
                MinLen: proto.Int64({{.StringField.MinLen}}),
                {{- end}}
                {{- if .StringField.MaxLen}}
                MaxLen: proto.Int64({{.StringField.MaxLen}}),
                {{- end}}
                {{- if .StringField.In}}
                In: []string{
                {{range .StringField.In}} "{{.}}",
                {{end}}
                },
                {{- end}}
                {{- if .StringField.Pattern}}
                Pattern: proto.String("{{.StringField.Pattern}}"),
                {{- end}}
            },
            {{- end}}
            {{- if .BytesField}}
            Bytes: &idl.BytesField{
                {{- if .BytesField.Required}}
                Required: proto.Bool({{.BytesField.Required}}),
                {{- end}}
                {{- if .BytesField.MinLen}}
                MinLen: proto.Int64({{.BytesField.MinLen}}),
                {{- end}}
                {{- if .BytesField.MaxLen}}
                MaxLen: proto.Int64({{.BytesField.MaxLen}}),
                {{- end}}
            },
            {{- end}}
            {{- if .ArrayField}}
            Array: &idl.ArrayField{
                {{- if .ArrayField.Default}}
                Required: proto.Bool({{.ArrayField.Required}}),
                {{- end}}
                {{- if .ArrayField.MinItems}}
                MinItems: proto.Int64({{.ArrayField.MinItems}})
                {{- end}}
                {{- if .ArrayField.MaxItems}}
                MaxItems: proto.Int64({{.ArrayField.MaxItems}})
                {{- end}}
                {{- if .ArrayField.Len}}
                Len: proto.Int64({{.ArrayField.Len}})
                {{- end}}
            },
            {{- end}}
        },
        {{end}}
    }
}
{{end}}


