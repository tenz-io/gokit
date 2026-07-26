package cache

import (
	"encoding/json"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

// The blob path was switched from JSON to MessagePack to shrink both the
// on-disk bytes held in the map and the per-op allocations. These benchmarks
// quantify that win so a regression (or a temptation to go back) is visible.

type benchUser struct {
	Name  string
	Email string
	Age   int
	Roles []string
}

func benchValue() benchUser {
	return benchUser{
		Name:  "tom",
		Email: "tom@example.com",
		Age:   38,
		Roles: []string{"admin", "ops", "viewer"},
	}
}

func BenchmarkBlobEncode_JSON(b *testing.B) {
	v := benchValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(v)
	}
}

func BenchmarkBlobEncode_Msgpack(b *testing.B) {
	v := benchValue()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = msgpack.Marshal(v)
	}
}

func BenchmarkBlobDecode_JSON(b *testing.B) {
	v := benchValue()
	data, _ := json.Marshal(v)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out benchUser
		_ = json.Unmarshal(data, &out)
	}
}

func BenchmarkBlobDecode_Msgpack(b *testing.B) {
	v := benchValue()
	data, _ := msgpack.Marshal(v)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var out benchUser
		_ = msgpack.Unmarshal(data, &out)
	}
}

// TestBlob_MsgpackSmallerThanJSON is not a perf assertion, it documents the
// headline win: the msgpack payload is strictly smaller than JSON for a
// typical struct, so each cached blob occupies less memory.
func TestBlob_MsgpackSmallerThanJSON(t *testing.T) {
	v := benchValue()
	jb, _ := json.Marshal(v)
	mb, _ := msgpack.Marshal(v)
	t.Logf("json=%dB msgpack=%dB  (%.0f%% of json)", len(jb), len(mb),
		float64(len(mb))/float64(len(jb))*100)
	if len(mb) >= len(jb) {
		t.Fatalf("msgpack (%dB) not smaller than json (%dB)", len(mb), len(jb))
	}
}
