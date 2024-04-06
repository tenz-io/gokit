package notiongo

import (
	"reflect"
	"testing"
)

func TestParent_MarshalJSON(t *testing.T) {
	type fields struct {
		Type ParentType
		ID   any
	}
	tests := []struct {
		name     string
		fields   fields
		wantJSON string
		wantErr  bool
	}{
		{
			name: "when database type",
			fields: fields{
				Type: ParentTypeDatabase,
				ID:   "some_database_id",
			},
			wantJSON: `{"type":"database_id","database_id":"some_database_id"}`,
			wantErr:  false,
		},
		{
			name: "when page type",
			fields: fields{
				Type: ParentTypePage,
				ID:   "some_page_id",
			},
			wantJSON: `{"type":"page_id","page_id":"some_page_id"}`,
			wantErr:  false,
		},
		{
			name: "when invalid type",
			fields: fields{
				Type: "invalid",
				ID:   "page_id",
			},
			wantJSON: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parent{
				Type: tt.fields.Type,
				ID:   tt.fields.ID,
			}
			gotJSON, err := p.MarshalJSON()
			t.Logf("err: %+v, gotJSON: %s", err, gotJSON)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parent.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(gotJSON) != tt.wantJSON {
				t.Errorf("Parent.MarshalJSON() = %v, want %v", string(gotJSON), tt.wantJSON)
			}
		})
	}
}

func TestParent_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Type ParentType
		ID   any
	}
	tests := []struct {
		name       string
		args       []byte
		wantParent *Parent
		wantErr    bool
	}{
		{
			name: "when database type",
			args: []byte(`{"type":"database_id","database_id":"some_database_id"}`),
			wantParent: &Parent{
				Type: ParentTypeDatabase,
				ID:   "some_database_id",
			},
			wantErr: false,
		},
		{
			name: "when page type",
			args: []byte(`{"type":"page_id","page_id":"some_page_id"}`),
			wantParent: &Parent{
				Type: ParentTypePage,
				ID:   "some_page_id",
			},
			wantErr: false,
		},
		{
			name:       "when invalid type",
			args:       []byte(`{"type":"invalid","page_id":"some_page_id"}`),
			wantParent: nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parent{}
			err := p.UnmarshalJSON(tt.args)
			t.Logf("err: %+v, gotParent: %+v", err, p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parent.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if !reflect.DeepEqual(p, tt.wantParent) {
				t.Errorf("Parent.UnmarshalJSON() = %v, want %v", p, tt.wantParent)
			}

		})
	}
}
