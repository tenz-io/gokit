package tracer

import (
	"fmt"
	"strings"
)

// flagByName 把每个被接受的字符串 token(规范名加别名,已小写化)映射到
// 其 flag。在 init 时从 flagTable 一次性构建,使解析热路径对每个 token
// 只需一次 map 查找,而非线性扫描。
var flagByName = func() map[string]Flag {
	m := make(map[string]Flag, len(flagTable)*2)
	for _, d := range flagTable {
		m[strings.ToLower(d.name)] = d.flag
		for _, a := range d.aliases {
			m[strings.ToLower(a)] = d.flag
		}
	}
	return m
}()

// allFlagBits 是所有已知 flag 位的 OR。String 用它把 Flag 中的命名位与
// 任何保留/未知位(Flag 上设置了但无 flagTable 条目认领的位 —— 例如更新
// 版本引入而当前构建未知的位)区分开。未知位被渲染为末尾的 0xNN 后缀,
// 而非被静默丢弃,使携带未知位的 Flag 不会被误认为 FlagNone。
var allFlagBits = func() Flag {
	var all Flag
	for _, d := range flagTable {
		all |= d.flag
	}
	return all
}()

// ParseFlag 把 "|" 分隔的 flag 字符串(如 "debug|shadow")解析成一个
// Flag。token 会做大小写折叠与 trim;未知 token 被忽略(因此
// "debug|bogus" 得到 FlagDebug,而非 FlagNone)。空字符串得到 FlagNone。
//
// 这是 lenient 解析器,为兼容此前调用方(如 ginext)手写的 per-transport
// switch 逻辑而保留。在 inbound 信任边界上,若拼错不得静默把 shadow/stress
// 流量降级为真实流量,请用 [ParseFlagStrict]。
//
// 它是 [Flag.String] 的逆运算:
//
//	tracer.ParseFlag(tracer.FlagDebug.String()) == tracer.FlagDebug
func ParseFlag(s string) Flag {
	f, _ := parseFlag(s, false)
	return f
}

// ParseFlagStrict 像 [ParseFlag] 一样解析 "|" 分隔的 flag 字符串,但返回
// 一个指出首个未知 token 的 error,而非静默丢弃它。用于 HTTP/gRPC 的
// inbound 边界 —— 在那里拼错的模式("shdow")必须暴露而非把 shadow 流量
// 降级为真实流量。no-op token("normal"、"none")与空 token 都被接受,
// 空字符串也被接受(得到 FlagNone、nil error)。"none" 是 FlagNone 的
// 字符串形式(见 [Flag.String])。
func ParseFlagStrict(s string) (Flag, error) {
	return parseFlag(s, true)
}

// parseFlag 是 ParseFlag(lenient)与 ParseFlagStrict 的共享引擎。当 strict
// 为 true 时,未知 token 产生非 nil error,并在首个此类 token 处停止解析
// (其已 OR 的位仍会被返回)。
func parseFlag(s string, strict bool) (Flag, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return FlagNone, nil
	}
	var f Flag
	for _, tok := range strings.Split(s, "|") {
		key := strings.ToLower(strings.TrimSpace(tok))
		if key == "" || key == "normal" || key == "none" {
			continue // "normal" / "none" / 空串显式表示 no-op 模式
		}
		bit, ok := flagByName[key]
		if !ok {
			if strict {
				return f, fmt.Errorf("tracer: unknown flag %q", tok)
			}
			continue // lenient: 忽略未知 token
		}
		f |= bit
	}
	return f, nil
}

// String 返回 f 已设置位的规范 "|" 连接形式,顺序遵循 flagTable,如
// "debug|shadow"。FlagNone 渲染为 "none"。
//
// f 上设置了但无 flagTable 条目认领的位(保留/未知位,例如更新版本引入
// 而当前构建未知的 flag)不会被丢弃:它们被追加为末尾 0xNN 后缀,使携带
// 未知位的 Flag 在版本不一致时不会被误认为 FlagNone。
func (f Flag) String() string {
	known := f & allFlagBits
	unknown := f &^ allFlagBits

	names := known.Names()
	switch {
	case len(names) == 0 && unknown == 0:
		return "none"
	case unknown == 0:
		// 只有已知位 —— 连接名称
		return joinNames(names)
	default:
		// 残留一些未知位;先渲染名称(若有)再接 0xNN。
		// 在交给 fmt 前先转换为 uint8:Flag 实现了 Stringer,
		// 因此 fmt.Sprintf("%x", unknown) 会重新进入 Flag.String
		// 并无限递归。原始 uint8 以 hex 打印,无方法派发。
		prefix := ""
		if len(names) > 0 {
			prefix = joinNames(names) + "|"
		}
		return prefix + fmt.Sprintf("0x%02x", uint8(unknown))
	}
}

// joinNames 以 "|" 按 flagTable 顺序连接 flag 名称。假定 Names 非空且
// 已排序。
func joinNames(names []string) string {
	if len(names) == 1 {
		return names[0]
	}
	var b strings.Builder
	b.Grow(len(names) * 8)
	b.WriteString(names[0])
	for _, n := range names[1:] {
		b.WriteByte('|')
		b.WriteString(n)
	}
	return b.String()
}

// GoString 返回 debug-printf 形式,如 "Flag(debug|shadow)" 或
// "Flag(none)",使 Flag 在 fmt.Printf("%#v", ...) 中可读打印。
func (f Flag) GoString() string { return "Flag(" + f.String() + ")" }
