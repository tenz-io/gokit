# v3 源码注释改写为简体中文

## 目标
将 10 个 v3 模块的**源码文件**(不含 `_test.go` 与 `example/`)中的英文注释改写为简体中文,保留英文技术术语。

## 范围(已与用户确认)
- ✅ 源码文件:共 **60 个**(`*.go`,排除 `*_test.go` 与 `*/example/*.go`)
- ❌ 不改测试文件、不改 example 示例文件
- ✅ 保留英文技术术语:API 名、函数/方法名、标准库类型名(如 `context.Context`、`RoundTripper`、`goroutine`、`TTL`、`LRU`、`flag`、`iter.Seq` 等)
- ❌ 不改动任何代码逻辑、不调整 import 顺序、不动 `go.mod`

## 翻译风格约定
1. **godoc 首行格式**:导出符号(类型、函数、方法、常量、变量)的注释首行保持 `符号名 — 中文说明` 或 `符号名 中文说明` 形式,确保 `go doc` 渲染正常、信息完整。
2. **段落与缩进**:保持原 godoc 的列表缩进(空格)、代码块缩进(单个 tab),不破坏 markdown 风格的渲染。
3. **代码示例注释**:示例代码块内的 `//` 注释也译为中文;但示例代码本身的字符串字面量(`t.Errorf` 的 want/goal 等)不动——源码文件里这类字面量本就少。
4. **断句**:自然语言用简体中文,标点用中文全角(。、,;)或保持英文逗点;数字与英文术语两侧不加多余空格改动,贴近原文排版。
5. **链接/引用**:保留 `[pkg.Symbol]`、`go doc` 引用格式不变,只翻译周围叙述。

## 按模块文件清单(60 个)
- **annotation/v3**(8): builtin, convert, defaults, errors, patterns, plan, validate, validator
- **app/v3**(7): admin, app, context, expand, flag, init, util
- **async/v3**(2): async, group
- **cache/v3**(4): cache, codec, local, lru
- **collection/v3**(7): doc, iter, priority_queue, queue, set_ops, set, stack
- **functional/v3**(10): aggregate, chain, conditional, dedupe_group, doc, heap, predicate, seq, set, transform
- **httpext/v3**(5): config, doc, interceptor, traffic, transport
- **logger/v3**(7): context, encoder, entry, logger, sink, traffic, trim
- **monitor/v3**(5): context, exporter, labels, monitor, recorder
- **tracer/v3**(5): context, doc, flag, parse, requestid

## 执行方式
逐文件用 Edit 工具替换英文注释为中文,每文件:
- 先 Read 确认当前注释,逐段 Edit 替换(保留代码行原样)。
- 仅替换 `//` 注释与 `/* */` 块注释内容,不动可执行代码。
- 内联尾随注释(`x := 1 // foo`)同样翻译 `foo` 部分。

## 验证
每个模块改完后:
1. `gofmt -l <模块>/v3/*.go` —— 应无输出(格式未变)。
2. `cd <模块>/v3 && go build ./...` —— 编译通过。
3. 最后 `git diff --stat` 汇总改动量,抽样抽查若干文件确认中英风格一致。

## 不做
- 不翻译测试文件、example 文件(用户已排除)。
- 不翻译字符串字面量、错误信息文案(保留英文便于日志检索)。
- 不改动代码逻辑、签名、import。
- 不新建/删除文件。
