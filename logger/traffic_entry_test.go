package logger

import "testing"

func Test_convertToMessage(t *testing.T) {
	type args struct {
		tb        *Traffic
		separator string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "when tb is nil then return empty string",
			args: args{
				tb:        nil,
				separator: "|",
			},
			want: "",
		},
		{
			name: "when tb is not nil then return string",
			args: args{
				tb: &Traffic{
					Typ: TrafficTypReq,
					Cmd: "test_command",
					Req: "request body",
				},
				separator: "|",
			},
			want: "req_to|test_command|-|-|-",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToMessage(tt.args.tb, tt.args.separator); got != tt.want {
				t.Errorf("convertToMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
