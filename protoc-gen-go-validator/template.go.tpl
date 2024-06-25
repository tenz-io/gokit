{{range .Messages}}
func (x *{{.Name}}) Validate(_ context.Context) error {
	return genproto.ValidateMessage(x.ValidateRule(), x)
}

func (x *{{.Name}}) ValidateRule() genproto.FieldRules {
	return genproto.FieldRules{
        {{range .Fields}}
        "{{.Name}}": &idl.Field{
            {{- if .Int}}
            Int: &idl.IntField{
                {{- if .Int.Default}}
                Default: proto.Int64({{.Int.Default}}),
                {{- end}}
                {{- if .Int.Required}}
                Required: proto.Bool({{.Int.Required}}),
                {{- end}}
                {{- if .Int.Gt}}
                Gt: proto.Int64({{.Int.Gt}}),
                {{- end}}
                {{- if .Int.Gte}}
                {{- end}}
                {{- if .Int.Gte}}
                Gte: proto.Int64({{.Int.Gte}}),
                {{- end}}
                {{- if .Int.Lt}}
                Lt: proto.Int64({{.Int.Lt}}),
                {{- end}}
                {{- if .Int.Lte}}
                Lte: proto.Int64({{.Int.Lte}}),
                {{- end}}
                {{- if .Int.Eq}}
                Eq: proto.Int64({{.Int.Eq}}),
                {{- end}}
                {{- if .Int.Ne}}
                Ne: proto.Int64({{.Int.Ne}}),
                {{- end}}
                {{- if .Int.In}}
                In: []int64{
                {{range .Int.In}}
                {{.}}, {{end}}
                },
                {{- end}}
                {{- if .Int.NotIn}}
                NotIn: []int64{
                {{range .Int.NotIn}}
                {{.}}, {{end}}
                },
                {{- end}}
            },
            {{- end}}
            {{- if .Str}}
            Str: &idl.StringField{
                {{- if .Str.Default}}
                Default: proto.String("{{.Str.Default}}"),
                {{- end}}
                {{- if .Str.Required}}
                Required: proto.Bool({{.Str.Required}}),
                {{- end}}
                {{- if .Str.NotBlank}}
                NotBlank: proto.Bool({{.Str.NotBlank}}),
                {{- end}}
                {{- if .Str.MinLen}}
                MinLen: proto.Int64({{.Str.MinLen}}),
                {{- end}}
                {{- if .Str.MaxLen}}
                MaxLen: proto.Int64({{.Str.MaxLen}}),
                {{- end}}
                {{- if .Str.In}}
                In: []string{
                {{range .Str.In}} "{{.}}",
                {{end}}
                },
                {{- end}}
                {{- if .Str.NotIn}}
                NotIn: []string{
                {{range .Str.NotIn}} "{{.}}",
                {{end}}
                },
                {{- end}}
                {{- if .Str.Pattern}}
                Pattern: proto.String("{{.Str.Pattern}}"),
                {{- end}}
            },
            {{- end}}
            {{- if .Bytes}}
            Bytes: &idl.BytesField{
                {{- if .Bytes.Required}}
                Required: proto.Bool({{.Bytes.Required}}),
                {{- end}}
                {{- if .Bytes.MinLen}}
                MinLen: proto.Int64({{.Bytes.MinLen}}),
                {{- end}}
                {{- if .Bytes.MaxLen}}
                MaxLen: proto.Int64({{.Bytes.MaxLen}}),
                {{- end}}
            },
            {{- end}}
            {{- if .Array}}
            Array: &idl.ArrayField{
                {{- if .Array.MinItems}}
                MinItems: proto.Int64({{.Array.MinItems}}),
                {{- end}}
                {{- if .Array.MaxItems}}
                MaxItems: proto.Int64({{.Array.MaxItems}}),
                {{- end}}
                {{- if .Array.Len}}
                Len: proto.Int64({{.Array.Len}}),
                {{- end}}
                {{- if .Array.Item}}
                Item: &idl.ItemField{
                    {{- if .Array.Item.Int}}
                    Int: &idl.IntField{
                        {{- if .Array.Item.Int.Default}}
                        Default: proto.Int64({{.Array.Item.Int.Default}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Required}}
                        Required: proto.Bool({{.Array.Item.Int.Required}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Gt}}
                        Gt: proto.Int64({{.Array.Item.Int.Gt}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Gte}}
                        Gte: proto.Int64({{.Array.Item.Int.Gte}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Lt}}
                        Lt: proto.Int64({{.Array.Item.Int.Lt}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Lte}}
                        Lte: proto.Int64({{.Array.Item.Int.Lte}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Eq}}
                        Eq: proto.Int64({{.Array.Item.Int.Eq}}),
                        {{- end}}
                        {{- if .Array.Item.Int.Ne}}
                        Ne: proto.Int64({{.Array.Item.Int.Ne}}),
                        {{- end}}
                        {{- if .Array.Item.Int.In}}
                        In: []int64{
                        {{range .Array.Item.Int.In}}
                        {{.}}, {{end}}
                        },
                        {{- end}}
                        {{- if .Array.Item.Int.NotIn}}
                        NotIn: []int64{
                        {{range .Array.Item.Int.NotIn}}
                        {{.}}, {{end}}
                        },
                        {{- end}}
                    },
                    {{- end}}
                    {{- if .Array.Item.Str}}
                    Str: &idl.StringField{
                        {{- if .Array.Item.Str.Default}}
                        Default: proto.String("{{.Array.Item.Str.Default}}"),
                        {{- end}}
                        {{- if .Array.Item.Str.Required}}
                        Required: proto.Bool({{.Array.Item.Str.Required}}),
                        {{- end}}
                        {{- if .Array.Item.Str.NotBlank}}
                        NotBlank: proto.Bool({{.Array.Item.Str.NotBlank}}),
                        {{- end}}
                        {{- if .Array.Item.Str.MinLen}}
                        MinLen: proto.Int64({{.Array.Item.Str.MinLen}}),
                        {{- end}}
                        {{- if .Array.Item.Str.MaxLen}}
                        MaxLen: proto.Int64({{.Array.Item.Str.MaxLen}}),
                        {{- end}}
                        {{- if .Array.Item.Str.In}}
                        In: []string{
                        {{range .Array.Item.Str.In}} "{{.}}",
                        {{end}}
                        },
                        {{- end}}
                        {{- if .Array.Item.Str.NotIn}}
                        NotIn: []string{
                        {{range .Array.Item.Str.NotIn}} "{{.}}",
                        {{end}}
                        },
                        {{- end}}
                        {{- if .Array.Item.Str.Pattern}}
                        Pattern: proto.String("{{.Array.Item.Str.Pattern}}"),
                        {{- end}}
                    },
                    {{- end}}
                },
                {{- end}}
            },
            {{- end}}
            {{- if .Float}}
            Float: &idl.FloatField{
                {{- if .Float.Default}}
                Default: proto.Float64({{.Float.Default}}),
                {{- end}}
                {{- if .Float.Required}}
                Required: proto.Bool({{.Float.Required}}),
                {{- end}}
                {{- if .Float.Gt}}
                Gt: proto.Float64({{.Float.Gt}}),
                {{- end}}
                {{- if .Float.Gte}}
                Gte: proto.Float64({{.Float.Gte}}),
                {{- end}}
                {{- if .Float.Lt}}
                Lt: proto.Float64({{.Float.Lt}}),
                {{- end}}
                {{- if .Float.Lte}}
                Lte: proto.Float64({{.Float.Lte}}),
                {{- end}}
            },
            {{- end}}
        },
        {{end}}
    }
}
{{end}}


