package notiongo

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTitle_JSON_marshal(t *testing.T) {
	type fields struct {
		title *Title
	}
	tests := []struct {
		name     string
		fields   fields
		wantJSON string
		wantErr  bool
	}{
		{
			name: "when title with link",
			fields: fields{
				title: NewTitleWithContent("some_title").WithLink("some_link"),
			},
			wantJSON: `{"type":"text","text":{"content":"some_title","link":"some_link"}}`,
		},
		{
			name: "when title without link",
			fields: fields{
				title: NewTitleWithContent("some_title"),
			},
			wantJSON: `{"type":"text","text":{"content":"some_title"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, err := json.Marshal(tt.fields.title)
			if (err != nil) != tt.wantErr {
				t.Errorf("Name.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("Name.MarshalJSON() = %v, want %v", string(gotJSON), tt.wantJSON)
			}
		})
	}
}

func TestTitleProperty_MarshalJSON(t1 *testing.T) {
	type fields struct {
		property *property
		Title    []*Title
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "when title property with title",
			fields: fields{
				property: nil,
				Title: []*Title{
					NewTitleWithContent("some_title"),
				},
			},
			want:    []byte(`{"title":[{"text":{"content":"some_title"},"type":"text"}]}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TitleProperty{
				property: tt.fields.property,
				Title:    tt.fields.Title,
			}
			got, err := t.MarshalJSON()
			t1.Logf("err: %+v, got: %s", err, got)
			if (err != nil) != tt.wantErr {
				t1.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Logf(" got: %s", got)
				t1.Logf("want: %s", tt.want)
				t1.Errorf("MarshalJSON() got = %s, want %s", got, tt.want)
			}
		})
	}
}
