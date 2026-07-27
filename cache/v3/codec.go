package cache

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

// encodeBlob 序列化 val 以便存储。它使用 MessagePack:紧凑的
// 二进制,无重复字段名,内存分配比 JSON 更少 ——
// 在网络传输与 map 中,对典型 struct 都明显更小。
//
// codec 被封装在这些 helper 之后,以便未来可插拔的 Codec option
// 能在不改动 backend 的情况下替换它。
func encodeBlob(val any) ([]byte, error) {
	bs, err := msgpack.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("cache: encode error: %w", err)
	}
	return bs, nil
}

// decodeBlob 将(由 encodeBlob 产生的)data 反序列化到 output,
// output 必须为指针。
func decodeBlob(data []byte, output any) error {
	if err := msgpack.Unmarshal(data, output); err != nil {
		return fmt.Errorf("cache: decode error: %w", err)
	}
	return nil
}
