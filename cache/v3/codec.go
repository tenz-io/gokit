package cache

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

// encodeBlob serializes val for storage. It uses MessagePack: compact
// binary, no repeated field names, and lighter allocations than JSON —
// noticeably smaller on the wire and in the map for typical structs.
//
// The codec is kept behind these helpers so a future pluggable Codec option
// can swap it without touching the backends.
func encodeBlob(val any) ([]byte, error) {
	bs, err := msgpack.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("cache: encode error: %w", err)
	}
	return bs, nil
}

// decodeBlob deserializes data (produced by encodeBlob) into output, which
// must be a pointer.
func decodeBlob(data []byte, output any) error {
	if err := msgpack.Unmarshal(data, output); err != nil {
		return fmt.Errorf("cache: decode error: %w", err)
	}
	return nil
}
