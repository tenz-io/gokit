package tracer

// Flag 表示一个随请求 context 传递的流量模式标志位掩码。
//
// Flag 开销很小(uint8)且可安全共享;零值 FlagNone 不携带任何模式。使用
// 按位 OR 运算符组合多个 flag:
//
//	f := tracer.FlagDebug | tracer.FlagShadow
//
// 已知的 flag、其名称与别名都位于 flagTable 中 —— 这张表是解析
// ([ParseFlag]) 与渲染([Flag.String]、[Flag.Names])的唯一事实来源,因此
// 新增一个模式只需改动一行。使用 uint8(无符号)可提供 8 个可用 flag 位,
// 并避免有符号 int8 在第 7 位上的符号位溢出陷阱。
type Flag uint8

// 预定义的流量模式 flag。
const (
	// FlagNone 是零值 Flag:未设置任何模式。
	FlagNone Flag = 0
	// FlagDebug 启用 debug 模式:更详细的 logging / tracing。
	FlagDebug Flag = 1 << iota
	// FlagStress 标记压测流量:隔离 metrics、跳过副作用等。
	FlagStress
	// FlagShadow 标记影子流量(record/replay):请求不应产生真实副作用。
	FlagShadow
)

// flagDef 描述注册表中的一条已知 flag。
type flagDef struct {
	flag Flag
	// name 是规范字符串形式,供 ParseFlag 和 String 使用。
	name string
	// aliases 是除 name 之外 ParseFlag 也接受的别名(已做大小写折叠)。
	// 可为 nil。
	aliases []string
}

// flagTable 是已知 flag 的唯一事实来源。顺序很重要:String 按此顺序渲染
// 各位,且最高有效位在 String 的 "all known" 快速路径中胜出。在此新增一个
// flag,解析与渲染会自动识别它。
var flagTable = []flagDef{
	{flag: FlagDebug, name: "debug"},
	{flag: FlagStress, name: "stress"},
	{flag: FlagShadow, name: "shadow"},
}

// Is 报告 flag f 是否设置了 x 的所有位。FlagNone 不是真实 flag:
// Is(FlagNone) 返回 false(flag 集合绝不可能是 "none")。要测试是否完全
// 无模式,直接比较 f == FlagNone。
func (f Flag) Is(x Flag) bool {
	if x == FlagNone {
		return false
	}
	return f&x == x
}

// HasAll 报告 f 是否包含 flags 的所有位。它是 Is 的别名,便于在调用处
// 传入组合掩码时提升可读性。HasAll(FlagNone) 出于同样原因返回 false。
func (f Flag) HasAll(flags Flag) bool { return f.Is(flags) }

// IsDebug 报告是否设置了 FlagDebug。
func (f Flag) IsDebug() bool { return f.Is(FlagDebug) }

// IsStress 报告是否设置了 FlagStress。
func (f Flag) IsStress() bool { return f.Is(FlagStress) }

// IsShadow 报告是否设置了 FlagShadow。
func (f Flag) IsShadow() bool { return f.Is(FlagShadow) }

// Names 返回 f 中所有已设置位的规范名称,顺序遵循 flagTable。对 FlagNone
// 返回 nil。它是 String 的结构化对应版本(String 用 "|" 连接同名列表)。
func (f Flag) Names() []string {
	if f == FlagNone {
		return nil
	}
	var names []string
	for _, d := range flagTable {
		if f.Is(d.flag) {
			names = append(names, d.name)
		}
	}
	return names
}
