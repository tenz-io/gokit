package retriever

import (
	"testing"
	"time"
)

func Test_pow(t *testing.T) {
	type args struct {
		base float64
		exp  int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "when exp is 0 then return 1",
			args: args{
				base: 2,
				exp:  0,
			},
			want: 1,
		},
		{
			name: "when exp is 1 then return base",
			args: args{
				base: 2,
				exp:  1,
			},
			want: 2,
		},
		{
			name: "when exp is 3 then return base^3",
			args: args{
				base: 2,
				exp:  3,
			},
			want: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pow(tt.args.base, tt.args.exp); got != tt.want {
				t.Errorf("pow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNoBackoff_Next(t *testing.T) {
	type args struct {
		in0 int
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "when failCount is 0 then return 0",
			args: args{
				in0: 0,
			},
			want: 0,
		},
		{
			name: "when failCount is 10 then return 0",
			args: args{
				in0: 10,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NoBackoff{}
			if got := n.Next(tt.args.in0); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinearBackoff_Next(t *testing.T) {
	type fields struct {
		duration time.Duration
	}
	type args struct {
		in0 int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Duration
	}{
		{
			name: "when failCount is 0 then return 100ms",
			fields: fields{
				duration: 100 * time.Millisecond,
			},
			args: args{
				in0: 0,
			},
			want: 100 * time.Millisecond,
		},
		{
			name: "when failCount is 10 then return 100ms",
			fields: fields{
				duration: 100 * time.Millisecond,
			},
			args: args{
				in0: 10,
			},
			want: 100 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LinearBackoff{
				duration: tt.fields.duration,
			}
			if got := l.Next(tt.args.in0); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExponentialBackoff_Next(t *testing.T) {
	type fields struct {
		base   float64
		factor float64
		jitter float64
	}
	type args struct {
		failCount int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Duration
	}{
		{
			name: "when failCount is 0 then return 100ms",
			fields: fields{
				base:   100,
				factor: 2,
				jitter: 0,
			},
			args: args{
				failCount: 0,
			},
			want: 100 * time.Millisecond,
		},
		{
			name: "when failCount is 1 then return 200ms",
			fields: fields{
				base:   100,
				factor: 2,
				jitter: 0,
			},
			args: args{
				failCount: 1,
			},
			want: 200 * time.Millisecond,
		},
		{
			name: "when failCount is 3 then return 800ms",
			fields: fields{
				base:   100,
				factor: 2,
				jitter: 0,
			},
			args: args{
				failCount: 3,
			},
			want: 800 * time.Millisecond,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExponentialBackoff{
				base:   tt.fields.base,
				factor: tt.fields.factor,
				jitter: tt.fields.jitter,
			}
			if got := e.Next(tt.args.failCount); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExponentialBackoffWithJitter_Next(t *testing.T) {
	type fields struct {
		base   float64
		factor float64
		jitter float64
	}
	type args struct {
		failCount int
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantRange [2]time.Duration // [min, max]
	}{
		{
			name: "when failCount is 0 with jitter 0.5 then return 100ms to 150ms",
			fields: fields{
				base:   100,
				factor: 2,
				jitter: 0.5,
			},
			args: args{
				failCount: 0,
			},
			wantRange: [2]time.Duration{100 * time.Millisecond, 150 * time.Millisecond},
		},
		{
			name: "when failCount is 1 with jitter 0.5 then return 200ms to 250ms",
			fields: fields{
				base:   100,
				factor: 2,
				jitter: 0.5,
			},
			args: args{
				failCount: 1,
			},
			wantRange: [2]time.Duration{200 * time.Millisecond, 250 * time.Millisecond},
		},
		{
			name: "when failCount is 3 with jitter 0.5 then return 800ms to 850ms",
			fields: fields{
				base:   100,
				factor: 2,
				jitter: 0.5,
			},
			args: args{
				failCount: 3,
			},
			wantRange: [2]time.Duration{800 * time.Millisecond, 850 * time.Millisecond},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExponentialBackoff{
				base:   tt.fields.base,
				factor: tt.fields.factor,
				jitter: tt.fields.jitter,
			}

			got := e.Next(tt.args.failCount)
			t.Logf("Next() = %v", got)
			if got < tt.wantRange[0] || got > tt.wantRange[1] {
				t.Errorf("Next() = %v, want range %v", got, tt.wantRange)
				return
			}

		})
	}
}
