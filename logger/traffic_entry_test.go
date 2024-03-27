package logger

import (
	"testing"
	"time"
)

func TestTraffic_HeadString(t1 *testing.T) {
	type fields struct {
		Typ  TrafficTyp
		Cmd  string
		Cost time.Duration
		Code string
		Msg  string
		Req  any
		Resp any
	}
	type args struct {
		sep string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "when tb not nil then return string",
			fields: fields{
				Typ:  TrafficTypRecv,
				Cmd:  "test_command",
				Code: "200",
				Cost: 15 * time.Millisecond,
				Msg:  "test message",
			},
			args: args{
				sep: "|",
			},
			want: `recv|test_command|15ms|200|test message`,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Traffic{
				Typ:  tt.fields.Typ,
				Cmd:  tt.fields.Cmd,
				Cost: tt.fields.Cost,
				Code: tt.fields.Code,
				Msg:  tt.fields.Msg,
				Req:  tt.fields.Req,
				Resp: tt.fields.Resp,
			}
			if got := t.headString(tt.args.sep); got != tt.want {
				t1.Errorf("headString() = %v, want %v", got, tt.want)
			}
		})
	}
}
