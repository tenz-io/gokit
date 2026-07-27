package app

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Expand 使用 lookup 解析变量值,替换 bs 中的 ${...} 占位符。
// 它是 ReadConfig 对原始
// config 字节在 unmarshal 之前执行的环境插值步骤,使敏感值(口令、DSN、
// API key)不进入 config 文件,而来自进程环境
// (由 WithDotEnvConfig 从 .env 加载,或直接设置)。
//
// 支持的语法(兼容 shell / docker-compose):
//
//	${VAR}            VAR 的值
//	${VAR:-default}   VAR 未设置或为空时取 default,否则取 VAR 的值
//	${VAR:?msg}       VAR 未设置或为空时报错(msg 可选)
//
// VAR 必须匹配 [A-Za-z_][A-Za-z0-9_]*。未跟 '{' 的裸 $ 保持
// 原样(因此 "$5" 或 "$HOME" 在正文中不会被篡改);只有花括号形式
// 被插值。default 本身可含 ${...},并针对同一 lookup 递归
// 展开。
//
// 默认严格展开:某 ${VAR} 的 VAR 未设置且无
// :-default / :?error 子句时为错误,因此缺失敏感值会
// 启动失败,而不会将字面量 "${VAR}" 泄漏到解码后的 config。
//
// 不含 '$' 的 bs 原样返回(快速路径,零分配)。
func Expand(bs []byte, lookup func(string) (string, bool)) ([]byte, error) {
	if !bytes.ContainsRune(bs, '$') {
		return bs, nil // 快速路径:无需展开
	}
	if lookup == nil {
		lookup = os.LookupEnv
	}
	ex := expander{lookup: lookup, depth: 0, maxDepth: 32}
	out, err := ex.expand(string(bs), nil)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// expander 携带 default-value 展开的递归守卫。
// `seen` 切片跟踪当前正在展开的占位符变量名,使
// "${A:-${A}}" 或互相引用的两个 default 无法无限循环。
type expander struct {
	lookup   func(string) (string, bool)
	seen     []string
	depth    int
	maxDepth int
}

// expand 解析 s 中的占位符。seen 累积当前展开路径上进入的变量名,
// 用于环检测。
func (ex *expander) expand(s string, seen []string) (string, error) {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		// 查找下一个 '$'。其前的内容均为字面量。
		dollar := strings.IndexByte(s[i:], '$')
		if dollar < 0 {
			b.WriteString(s[i:])
			break
		}
		// 拷贝 '$' 之前的字面量段。
		b.WriteString(s[i : i+dollar])
		j := i + dollar // '$' 的下标
		// 要求 '$' 之后紧跟 '{';否则 '$' 保持字面量。
		if j+1 >= len(s) || s[j+1] != '{' {
			b.WriteByte(s[j])
			i = j + 1
			continue
		}
		// 查找匹配的 '}'。嵌套很重要:一个 default 值本身可能
		// 含 ${...}(如 ${A:-${B}}),因此我们扫描与
		// 开始 ${ 配对的 '}' ,并计数嵌套的 ${ 开括号。
		close, ok := matchClose(s, j) // j 为 '$' 的下标
		if !ok {
			return "", fmt.Errorf("app: unterminated placeholder in config: %q", s[j:])
		}
		inner := s[j+2 : j+2+close] // ${ 与 } 之间的文本
		i = j + 2 + close + 1       // 从 '}' 之后继续

		val, err := ex.resolve(inner, seen)
		if err != nil {
			return "", err
		}
		b.WriteString(val)
	}
	return b.String(), nil
}

// resolve 将单个占位符内部文本(${ 与 } 之间)解析为
// 名字与可选 operator(:- 或 :?),再产出最终值。
func (ex *expander) resolve(inner string, seen []string) (string, error) {
	// 若存在 ':',则在首个 ':' 处分割,但仅当其后跟 '-' 或 '?' 时 ——
	// 否则 ':' 是值的一部分(如带端口的 URL)。查找 :- 或 :?。
	name, op, arg := "", byte(0), ""
	if idx := indexOperator(inner); idx >= 0 {
		name = inner[:idx]
		op = inner[idx+1] // '-' 或 '?'
		arg = inner[idx+2:]
	} else {
		name = inner
	}

	if !validName(name) {
		return "", fmt.Errorf("app: invalid variable name %q in placeholder", name)
	}

	// 环守卫:若该名字已在当前
	// 路径上展开(我们处于其自身的 default 表达式内),则为
	// 循环引用,如 ${A:-${A}} —— 报错而非递归。
	for _, n := range seen {
		if n == name {
			return "", fmt.Errorf("app: circular config placeholder reference to %q", name)
		}
	}

	val, set := ex.lookup(name)

	switch op {
	case 0: // 纯 ${VAR}
		if !set || val == "" {
			return "", fmt.Errorf("app: config placeholder ${%s} is unset or empty", name)
		}
		return val, nil
	case '-': // ${VAR:-default}
		if set && val != "" {
			return val, nil
		}
		return ex.expandDefault(arg, name, seen)
	case '?': // ${VAR:?msg}
		if set && val != "" {
			return val, nil
		}
		if strings.TrimSpace(arg) != "" {
			return "", fmt.Errorf("app: config variable %s: %s", name, arg)
		}
		return "", fmt.Errorf("app: config variable %s is required but unset or empty", name)
	}
	return val, nil
}

// expandDefault 递归展开 default 表达式,并防范
// 经由被取 default 的变量形成的环。
func (ex *expander) expandDefault(arg, name string, seen []string) (string, error) {
	if ex.depth >= ex.maxDepth {
		return "", fmt.Errorf("app: config placeholder nesting too deep (max %d)", ex.maxDepth)
	}
	for _, n := range seen {
		if n == name {
			return "", fmt.Errorf("app: circular config placeholder reference to %q", name)
		}
	}
	// 仅当 default 自身含占位符时才递归;否则
	// 返回字面量(常见情况:${DB_PASS:-fallback})。
	if !strings.Contains(arg, "${") {
		return arg, nil
	}
	ex.depth++
	defer func() { ex.depth-- }()
	return ex.expand(arg, append(seen, name))
}

// indexOperator 返回 inner 中 ':' 后跟 '-' 或 '?' 时
// 该 ':' 的下标,即 ':-' 或 ':?' 分隔符。inner 无此
// operator 时返回 -1。分隔符取首次出现位置,因此 default 值可
// 含冒号(如 ${URL:-http://host:8080})。
func indexOperator(inner string) int {
	for i := 0; i < len(inner)-1; i++ {
		if inner[i] == ':' && (inner[i+1] == '-' || inner[i+1] == '?') {
			return i
		}
	}
	return -1
}

// matchClose 从下标 dollarPos 处 '${' 之后开始扫描 s,
// 返回关闭该占位符的 '}' 相对于 s[dollarPos+2] 的字节偏移。
// 它跟踪嵌套的 '${' 开括号,使含自身
// 占位符的 default 值(如 ${A:-${B}})能匹配到
// 正确的闭花括号,而非内层的。
func matchClose(s string, dollarPos int) (int, bool) {
	depth := 1 // 我们已处于一个打开的 ${ 之内
	for k := dollarPos + 2; k < len(s); k++ {
		switch {
		case s[k] == '$' && k+1 < len(s) && s[k+1] == '{':
			depth++
			k++ // 消费 '{'
		case s[k] == '}':
			depth--
			if depth == 0 {
				return k - (dollarPos + 2), true
			}
		}
	}
	return 0, false
}

// validName 报告 name 是否为 shell 风格标识符:首字符为
// 字母或下划线,其余为字母/数字/下划线。空串非法。
func validName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
