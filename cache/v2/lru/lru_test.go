package lru

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestRemove(t *testing.T) {
	lru := New[string, int](0, nil, 0)
	lru.Set("myKey", 1234, 0)
	if val, ok := lru.Get("myKey"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != 1234 {
		t.Fatalf("TestRemove failed.  Expected %d, got %v", 1234, val)
	}

	lru.Remove("myKey")
	if _, ok := lru.Get("myKey"); ok {
		t.Fatal("TestRemove returned a removed item")
	}
}

func TestExpire(t *testing.T) {
	lru := New[string, int](0, nil, 0)
	lru.Set("myKey", 1234, time.Duration(1000)*time.Millisecond)
	if val, ok := lru.Get("myKey"); !ok {
		t.Fatal("TestExpire returned no match")
	} else if val != 1234 {
		t.Fatalf("TestExpire failed.  Expected %d, got %v", 1234, val)
	}

	time.Sleep(time.Duration(1000) * time.Millisecond)
	if _, ok := lru.Get("myKey"); ok {
		t.Fatal("TestExpire failed, got a expired item.")
	}
}

func TestEvict(t *testing.T) {
	evictedKeys := make([]string, 0)
	onEvictedFun := func(key string, value int) {
		evictedKeys = append(evictedKeys, key)
	}

	lru := New[string, int](20, nil, 0)
	lru.onEvicted = onEvictedFun
	for i := 0; i < 22; i++ {
		lru.Set(fmt.Sprintf("myKey%d", i), 1234, 0)
	}

	if len(evictedKeys) != 2 {
		t.Fatalf("got %d evicted keys; want 2", len(evictedKeys))
	}
	if evictedKeys[0] != ("myKey0") {
		t.Fatalf("got %v in first evicted key; want %s", evictedKeys[0], "myKey0")
	}
	if evictedKeys[1] != ("myKey1") {
		t.Fatalf("got %v in second evicted key; want %s", evictedKeys[1], "myKey1")
	}
}

func TestCache_Get(t *testing.T) {
	type args[K comparable] struct {
		key K
	}
	type testCase[K comparable, V any] struct {
		name         string
		c            *Cache[K, V]
		args         args[K]
		wantVal      V
		wantExisting bool
	}
	tests := []testCase[string, int]{
		{
			name: "when key exists then return value",
			c: func() *Cache[string, int] {
				c := New[string, int](10, nil, 0)
				c.Set("myKey", 1234, 0)
				return c
			}(),
			args:         args[string]{"myKey"},
			wantVal:      1234,
			wantExisting: true,
		},
		{
			name: "when key does not exist then return false",
			c: func() *Cache[string, int] {
				c := New[string, int](10, nil, 0)
				return c
			}(),
			args:         args[string]{"myKey"},
			wantVal:      0,
			wantExisting: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotExisting := tt.c.Get(tt.args.key)
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("Get() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
			if gotExisting != tt.wantExisting {
				t.Errorf("Get() gotExisting = %v, want %v", gotExisting, tt.wantExisting)
			}
		})
	}
}
