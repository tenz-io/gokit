package notiongo

import (
	"encoding/json"
	"testing"
)

func TestProperties_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name      string
		props     Properties
		args      args
		wantProps Properties
		wantErr   bool
	}{
		{
			name:  "UnmarshalJSON from create database response",
			props: Properties{},
			args: args{
				data: []byte(`{"Description":{"id":"%3EWW~","name":"Description","type":"rich_text","rich_text":{}},"Last ordered":{"id":"O%5C%3BK","name":"Last ordered","type":"date","date":{}},"In stock":{"id":"Pya%5C","name":"In stock","type":"checkbox","checkbox":{}},"+1":{"id":"%5CSky","name":"+1","type":"people","people":{}},"Photo":{"id":"dSrT","name":"Photo","type":"files","files":{}},"Store availability":{"id":"jRd%3E","name":"Store availability","type":"multi_select","multi_select":{"options":[{"id":"8e6441ee-8f17-4833-a2fe-68af5dced24f","name":"Duc Loi Market","color":"blue"},{"id":"64a9da77-9805-461f-9773-1e176fdbd203","name":"Rainbow Grocery","color":"gray"},{"id":"012d0436-66a1-4613-a1bd-314b1d1d059b","name":"Nijiya Market","color":"purple"},{"id":"63ab31f9-8cbd-4d02-8688-752376f455ea","name":"Gus's Community Market","color":"yellow"}]}},"Food group":{"id":"q%5DO%5B","name":"Food group","type":"select","select":{"options":[{"id":"392af858-f42f-43ea-a171-7c0ca5c0a683","name":"🥦Vegetable","color":"green"},{"id":"df461a24-14c6-494a-8c61-55775fedbdcd","name":"🍎Fruit","color":"red"},{"id":"0ff22aaa-348e-4194-83c2-67a76dfb10fc","name":"💪Protein","color":"yellow"}]}},"Price":{"id":"t%60jj","name":"Price","type":"number","number":{"format":"dollar"}},"Name":{"id":"title","name":"Name","type":"title","title":{}}}`),
			},
			wantProps: Properties{},
			wantErr:   false,
		},
		{
			name:  "UnmarshalJSON from append pages response",
			props: Properties{},
			args: args{
				data: []byte(`{"Score /5":{"id":")Y7%22","type":"select","select":{"id":"5c944de7-3f4b-4567-b3a1-fa2c71c540b6","name":"⭐️⭐️⭐️⭐️⭐️","color":"default"}},"Type":{"id":"%2F7eo","type":"select","select":{"id":"672b014a-2626-4ada-9211-fb3613d07ae2","name":"Article","color":"default"}},"Publisher":{"id":"%3E%24Pb","type":"select","select":{"id":"01f82d08-aa1f-4884-a4e0-3bc32f909ec4","name":"The Atlantic","color":"red"}},"Summary":{"id":"%3F%5C25","type":"rich_text","rich_text":[{"type":"text","text":{"content":"Some think chief ethics officers could help technology companies navigate political and social questions.","link":null},"annotations":{"bold":false,"italic":false,"strikethrough":false,"underline":false,"code":false,"color":"default"},"plain_text":"Some think chief ethics officers could help technology companies navigate political and social questions.","href":null}]},"Publishing/Release Date":{"id":"%3Fex%2B","type":"date","date":{"start":"2020-12-08T12:00:00.000+00:00","end":null,"time_zone":null}},"Link":{"id":"VVMi","type":"url","url":"https://www.nytimes.com/2018/10/21/opinion/who-will-teach-silicon-valley-to-be-ethical.html"},"Read":{"id":"_MWJ","type":"checkbox","checkbox":false},"Status":{"id":"%60zz5","type":"select","select":{"id":"8c4a056e-6709-4dd1-ba58-d34d9480855a","name":"Ready to Start","color":"yellow"}},"Author":{"id":"qNw_","type":"multi_select","multi_select":[]},"Name":{"id":"title","type":"title","title":[{"type":"text","text":{"content":"New Media Article","link":null},"annotations":{"bold":false,"italic":false,"strikethrough":false,"underline":false,"code":false,"color":"default"},"plain_text":"New Media Article","href":null}]}}`),
			},
			wantProps: Properties{},
			wantErr:   false,
		},
		{
			name:  "UnmarshalJSON from retrieve pages response",
			props: Properties{},
			args: args{
				data: []byte(`{"Score /5":{"id":")Y7%22","type":"select","select":{"id":"b7307e35-c80a-4cb5-bb6b-6054523b394a","name":"⭐️⭐️⭐️⭐️","color":"default"}},"Type":{"id":"%2F7eo","type":"select","select":{"id":"f96d0d0a-5564-4a20-ab15-5f040d49759e","name":"Article","color":"default"}},"Publisher":{"id":"%3E%24Pb","type":"select","select":{"id":"c5ee409a-f307-4176-99ee-6e424fa89afa","name":"NYT","color":"default"}},"Summary":{"id":"%3F%5C25","type":"rich_text","rich_text":[{"type":"text","text":{"content":"Some think chief ethics officers could help technology companies navigate political and social questions.","link":null},"annotations":{"bold":false,"italic":false,"strikethrough":false,"underline":false,"code":false,"color":"default"},"plain_text":"Some think chief ethics officers could help technology companies navigate political and social questions.","href":null}]},"Publishing/Release Date":{"id":"%3Fex%2B","type":"date","date":{"start":"2018-10-21","end":null,"time_zone":null}},"Link":{"id":"VVMi","type":"url","url":"https://www.nytimes.com/2018/10/21/opinion/who-will-teach-silicon-valley-to-be-ethical.html"},"Read":{"id":"_MWJ","type":"checkbox","checkbox":true},"Status":{"id":"%60zz5","type":"select","select":{"id":"5925ba22-0126-4b58-90c7-b8bbb2c3c895","name":"Reading","color":"red"}},"Author":{"id":"qNw_","type":"multi_select","multi_select":[{"id":"833e2c78-35ed-4601-badc-50c323341d76","name":"Kara Swisher","color":"default"}]},"Name":{"id":"title","type":"title","title":[{"type":"text","text":{"content":"Who Will Teach Silicon Valley to Be Ethical? ","link":null},"annotations":{"bold":false,"italic":false,"strikethrough":false,"underline":false,"code":false,"color":"default"},"plain_text":"Who Will Teach Silicon Valley to Be Ethical? ","href":null}]}}`),
			},
			wantProps: Properties{},
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.props.UnmarshalJSON(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			j, err := json.Marshal(tt.props)
			if err != nil {
				t.Errorf("Marshal() error = %+v", err)
				return
			}
			t.Logf("got: %s", j)

			//assert.Equal(t, len(tt.wantProps), len(tt.props))
		})
	}
}
