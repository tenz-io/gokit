package tracer

import (
	"context"
	"testing"
)

func Test_TraceFlag(t *testing.T) {
	t.Run("test Flag", func(t *testing.T) {
		t.Logf("FlagDebug: %v", FlagDebug)
		t.Logf("FlagStress: %v", FlagStress)
		t.Logf("FlagShadow: %v", FlagShadow)
	})
}

func TestFlag_IsDebug(t *testing.T) {
	tests := []struct {
		name string
		f    Flag
		want bool
	}{
		{
			name: "test FlagDebug",
			f:    FlagDebug,
			want: true,
		},
		{
			name: "test FlagStress",
			f:    FlagStress,
			want: false,
		},
		{
			name: "test FlagDebug | FlagStress",
			f:    FlagDebug | FlagStress,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.IsDebug(); got != tt.want {
				t.Errorf("IsDebug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name     string
		args     args
		wantFlag Flag
	}{
		{
			name: "test nil context",
			args: args{
				ctx: nil,
			},
			wantFlag: FlagNone,
		},
		{
			name: "test FlagDebug",
			args: args{
				ctx: WithFlag(context.Background(), FlagDebug),
			},
			wantFlag: FlagDebug,
		},
		{
			name: "test FlagStress",
			args: args{
				ctx: WithFlag(context.Background(), FlagStress),
			},
			wantFlag: FlagStress,
		},
		{
			name: "test FlagShadow",
			args: args{
				ctx: WithFlag(context.Background(), FlagShadow),
			},
			wantFlag: FlagShadow,
		},
		{
			name: "test FlagDebug | FlagStress then is FlagDebug",
			args: args{
				ctx: WithFlag(context.Background(), FlagDebug|FlagStress),
			},
			wantFlag: FlagDebug,
		},
		{
			name: "test FlagDebug | FlagStress then is FlagStress",
			args: args{
				ctx: WithFlag(context.Background(), FlagDebug|FlagStress),
			},
			wantFlag: FlagStress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFlag := FromContext(tt.args.ctx); !gotFlag.Is(tt.wantFlag) {
				t.Errorf("FromContext() = %v, want %v", gotFlag, tt.wantFlag)
			}
		})
	}
}
