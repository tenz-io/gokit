package monitor

import "github.com/prometheus/client_golang/prometheus"

// 每个 metric family 共享的 label 名。V3 在 counter/gauge/histogram/
// summary 之间统一了 label 集合(v2 的 histogram 丢弃了 `opt`,导致基数
// 策略不一致);现在每个 family 都携带相同的四个维度,从而使 dashboard
// 与查询保持一致。
const (
	labelCmd   = "cmd"
	labelDsCmd = "dsCmd"
	labelCode  = "code"
	labelOpt   = "opt"

	// defaultNamespace/Subsystem 是 metric 的前缀;可通过 Init 覆盖。
	defaultNamespace = "gokit"
	defaultSubsystem = "flight"

	// valNA 是空字符串 label 的占位符,使 Prometheus 永远不会看到未设置
	// 的 label 值(否则会静默地分裂序列)。
	valNA = "NA"

	// codeOK / codeErr 是归一化之后仅保留的两个结果码,无论调用方如何
	// 书写错误码都能保持基数受限。
	codeOK  = "0"
	codeErr = "1"

	// optActive 是活跃请求 gauge 的 opt 槽位:跟踪在途调用的 gauge 使用
	// opt="actives",从而不会与 hit/miss 这类业务 opt 维度冲突。
	optActive = "actives"
)

// latencyBuckets 是以毫秒为单位的延迟 histogram bucket 布局,覆盖
// 0.1ms 到 10s。复用自 v2(在代码库中久经实战)。
var latencyBuckets = []float64{
	1e-1,     // 0.1ms  factor 10
	1e0, 3e0, // 1ms    factor 3
	1e1, 2e1, 4e1, 8e1, // 10ms   factor 2
	1.6e2, 3.2e2, 6.4e2, // 160ms  factor 2
	1e3, 3e3, // 1000ms factor 3
	1e4, // 10000ms to infinite
}

// summaryObjectives 配置数据大小 summary 的 quantile 目标。
var summaryObjectives = map[float64]float64{
	0.5:  0.05,
	0.9:  0.01,
	0.95: 0.05,
	0.99: 0.001,
}

// normalizeOpt 将空 opt 映射为 NA 占位符,使被省略的 opt 值合并为
// 同一条序列而非多条。
func normalizeOpt(opt string) string {
	if opt == "" {
		return valNA
	}
	return opt
}

// normalizeCode 将任意非零 code 折叠为 "1"(err),将空 code 折叠为
// "0"(ok)。这限制了 observe/sample 中 code 的基数:业务在延迟与
// 数据大小 metric 上只会看到 ok/err,而精确 code 仍在 counter/gauge
// 上作为一等值被保留。
func normalizeCode(code string) string {
	if code == "" || code == codeOK {
		return codeOK
	}
	return codeErr
}

// labelsOf 为某个 metric 构造标准的四维 label 集合。cmd 是 Exporter 的
// 作用域;dsCmd/code/opt 是按调用的。
//
// 生产热路径使用 Exporter.WithLabelValues(无 map 分配的 slice 快速
// 路径);labelsOf 保留下来供测试与工具使用,它们需要一个现成的
// prometheus.Labels map 用于断言与采集。它不做归一化 —— 调用方需
// 传入已归一化的值。
func labelsOf(cmd, dsCmd, code, opt string) prometheus.Labels {
	return prometheus.Labels{
		labelCmd:   cmd,
		labelDsCmd: dsCmd,
		labelCode:  code,
		labelOpt:   opt,
	}
}
