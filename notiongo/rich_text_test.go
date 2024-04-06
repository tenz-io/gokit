package notiongo

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRichText_JSON_marshal(t *testing.T) {
	type fields struct {
		richText *RichText
	}
	tests := []struct {
		name     string
		fields   fields
		wantJSON string
		wantErr  bool
	}{
		{
			name: "when rich text with plain text",
			fields: fields{
				richText: NewRichTextWithContent("some_text").WithPlainText("some_plain_text"),
			},
			wantJSON: `{"type":"text","text":{"content":"some_text"},"plain_text":"some_plain_text"}`,
			wantErr:  false,
		},
		{
			name: "when rich text with link",
			fields: fields{
				richText: NewRichTextWithContent("some_text").WithLink("some_link"),
			},
			wantJSON: `{"type":"text","text":{"content":"some_text","link":"some_link"}}`,
			wantErr:  false,
		},
		{
			name: "when rich text with annotations",
			fields: fields{
				richText: NewRichTextWithContent("some_text").WithAnnotations(NewAnnotations()),
			},
			wantJSON: `{"type":"text","text":{"content":"some_text"},"annotations":{}}`,
			wantErr:  false,
		},
		{
			name: "when rich text annotations bold",
			fields: fields{
				richText: NewRichTextWithContent("some_text").WithAnnotations(NewAnnotations().WithBold()),
			},
			wantJSON: `{"type":"text","text":{"content":"some_text"},"annotations":{"bold":true}}`,
			wantErr:  false,
		},
		{
			name: "when rich text with href",
			fields: fields{
				richText: NewRichTextWithContent("some_text").WithHref("some_href"),
			},
			wantJSON: `{"type":"text","text":{"content":"some_text"},"href":"some_href"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, err := json.Marshal(tt.fields.richText)
			if (err != nil) != tt.wantErr {
				t.Errorf("RichText.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("RichText.MarshalJSON() = %v, want %v", string(gotJSON), tt.wantJSON)
			}
		})
	}
}

func TestRichTextProperty_MarshalJSON(t *testing.T) {
	type fields struct {
		property *property
		RichText []*RichText
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "when rich text property with rich text",
			fields: fields{
				property: nil,
				RichText: []*RichText{
					NewRichTextWithContent("some_text"),
				},
			},
			want:    []byte(`{"rich_text":[{"text":{"content":"some_text"},"type":"text"}]}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &RichTextProperty{
				property: tt.fields.property,
				RichText: tt.fields.RichText,
			}
			got, err := r.MarshalJSON()
			t.Logf("err: %+v, got: %s", err, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf(" got: %s", got)
				t.Logf("want: %s", tt.want)
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}
