package cache

import (
	"context"
	"testing"
	"time"
)

func Test_local_Get(t *testing.T) {
	type fields struct {
		m       map[string]*item
		nowFunc func() time.Time
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRaw string
		wantErr bool
	}{
		{
			name: "when key not found then return ErrNotFound",
			fields: fields{
				m: map[string]*item{},
				nowFunc: func() time.Time {
					return time.Now()
				},
			},
			args: args{
				ctx: context.Background(),
				key: "abc",
			},
			wantRaw: "",
			wantErr: true,
		},
		{
			name: "when key found but expired then return ErrNotFound",
			fields: fields{
				m: map[string]*item{
					"abc": {
						raw:    []byte("123"),
						expire: time.Now().Unix() - 100000,
					},
				},
				nowFunc: func() time.Time {
					return time.Now()
				},
			},
			args: args{
				ctx: context.Background(),
				key: "abc",
			},
			wantRaw: "",
			wantErr: true,
		},
		{
			name: "when key found and not expired then return value",
			fields: fields{
				m: map[string]*item{
					"abc": {
						raw:    []byte("123"),
						expire: time.Now().Unix() + 100000,
					},
				},
				nowFunc: func() time.Time {
					return time.Now()
				},
			},
			args: args{
				ctx: context.Background(),
				key: "abc",
			},
			wantRaw: "123",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &local{
				m:       tt.fields.m,
				nowFunc: tt.fields.nowFunc,
			}
			gotRaw, err := l.Get(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRaw != tt.wantRaw {
				t.Errorf("Get() gotRaw = %v, want %v", gotRaw, tt.wantRaw)
			}
		})
	}
}
