package notiongo

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCreateDatabaseRequest_MarshalJSON(t *testing.T) {
	type fields struct {
		Parent     *Parent
		Title      []*Title
		Properties Properties
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "when parent and title and properties are not nil",
			fields: fields{
				Parent: NewParent("some_database_id").WithType(ParentTypeDatabase),
				Title: []*Title{
					NewTitleWithContent("Grocery List"),
				},
				Properties: Properties{
					NewProperty("Name", PropertyTypeTitle),
					NewProperty("Description", PropertyTypeRichText),
					NewProperty("In Stock", PropertyTypeCheckbox),
					NewSelectProperty("Food Group").
						WithSelectOptions([]*SelectOption{
							NewSelectOption("🥦 Vegetables"),
							NewSelectOption("🍎 Fruits"),
							NewSelectOption("🍞 Grains"),
						}),
					NewNumberProperty("Price").WithNumber(&Number{Format: "dollar"}),
					NewProperty("Last Ordered", PropertyTypeDate),
					NewMultiSelectProperty("Store Availability").WithSelectOptions([]*SelectOption{
						NewSelectOption("Duc Loi Market").WithColor("blue"),
						NewSelectOption("Rainbow Grocery").WithColor("gray"),
						NewSelectOption("Nijiya Market").WithColor("purple"),
						NewSelectOption("Gus's Community Market").WithColor("yellow"),
					}),
					NewProperty("+1", PropertyTypePeople),
					NewProperty("Photo", PropertyTypeFile),
				},
			},
			want:    []byte(`{"parent":{"type":"database_id","database_id":"some_database_id"},"title":[{"type":"text","text":{"content":"Grocery List"}}],"properties":{"+1":{"people":{}},"Description":{"rich_text":{}},"Food Group":{"select":{"options":[{"name":"🥦 Vegetables"},{"name":"🍎 Fruits"},{"name":"🍞 Grains"}]}},"In Stock":{"checkbox":{}},"Last Ordered":{"date":{}},"Name":{"title":{}},"Photo":{"files":{}},"Price":{"number":{"format":"dollar"}},"Store Availability":{"multi_select":{"options":[{"name":"Duc Loi Market","color":"blue"},{"name":"Rainbow Grocery","color":"gray"},{"name":"Nijiya Market","color":"purple"},{"name":"Gus's Community Market","color":"yellow"}]}}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CreateDatabaseRequest{
				Parent:     tt.fields.Parent,
				Title:      tt.fields.Title,
				Properties: tt.fields.Properties,
			}
			got, err := json.Marshal(c)
			t.Logf("err: %+v, got: %s", err, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf(" got: %s", got)
				t.Logf("want: %s", tt.want)
				t.Errorf("got does not equal want")
			}
		})
	}
}
