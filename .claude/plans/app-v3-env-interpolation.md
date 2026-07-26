# app/v3 配置占位符插值（environment interpolation）

## 目标
配置文件（YAML/JSON）里写 `${VAR}` 占位符，实际值来自进程环境变量（由 `WithDotEnvConfig` 加载的 `.env` 或真实环境提供）。敏感配置不落盘。

## 语法（shell/docker-compose 通行）
- `${VAR}` — 替换为环境变量 VAR 的值
- `${VAR:-default}` — VAR 未设**或为空**时用 `default`
- `${VAR:?error}` — VAR 未设**或为空**时报错（错误信息为 `error`，可省略）

## 语义（已确认）
- **默认严格**：配置里写了 `${VAR}` 而 VAR 未设且无 `:-` 兜底 / `:?` 报错时，`ReadConfig` 返回 error，启动失败。"漏配敏感配置"就该启动失败。
- 不含 `${}` 的配置完全不受影响，零行为变化、零性能开销（快速路径：无 `$` 直接返回原字节）。
- 替换在**原始字节层面、unmarshal 之前**进行 —— 类型转换（int/bool/duration）交给现有 unmarshal 逻辑，无需特例。
- 变量名规则：`[A-Za-z_][A-Za-z0-9_]*`。`$` 后非 `{` 或非字母下划线 → 原样保留（不误吞 shell 风格的 `$VAR`，避免歧义；文档明确只支持 `${}` 形式）。
- 嵌套替换：默认值 `default` 内部若含 `${...}` 也递归展开（同一 lookup 集合，防无限递归用"已展开占位符集合 + 最大深度"双保险）。

## 集成点
`ReadConfig(path, conf, unmarshal)`：在 `ApplyDefaults` 之后、`unmarshal` 之前插入一步 `Expand(bs, os.LookupEnv)`。
- 顺序：`ApplyDefaults`（结构体默认值，作用于 conf 指针，不碰字节）→ `Expand`（字节层占位符替换）→ `unmarshal`（字节→conf）→ `Validate`。
- 注意 `ApplyDefaults` 操作的是 conf 结构体而非字节，所以 Expand 在 unmarshal 前是对的；defaults 与 env 各管一层不冲突。
- `Expand` 用 `os.LookupEnv` 取值，这样 `.env` 已注入的环境变量自然被读到（前提：`WithDotEnvConfig()` 排在 `WithYamlConfig()` 之前 —— example 已是这个顺序，README 注明）。

## 新增文件 / 改动
1. **`app/v3/expand.go`**（新）— 纯函数，无包内依赖，可独立测试：
   - `Expand(bs []byte, lookup func(string)(string,bool)) ([]byte, error)` — 主入口
   - 内部：扫描 `bs` 找 `${`，解析到匹配的 `}`，按 `:-` / `:?` 分支处理。
   - 快速路径：`bytes.IndexByte(bs, '$') == -1` → 直接返回原 bs。
   - 用 `strings.Builder`/`bytes.Buffer` 拼接，避免反复分配。
   - 递归展开默认值时用 `seen` map 防环 + 深度计数上限（如 32）防恶意嵌套。
2. **`app/v3/expand_test.go`**（新）— 表驱动测试：
   - `${VAR}` 存在/不存在（严格报错）
   - `${VAR:-default}` 未设/设非空/设空字符串
   - `${VAR:?msg}` 未设/设空报错、已设正常
   - 无 `$` 快速路径原样返回
   - `$VAR`（无花括号）原样保留
   - 嵌套 `${A:-${B}}`
   - JSON/YAML 多值
   - 未闭合 `${VAR` 报错
3. **`app/v3/init.go`** — `ReadConfig` 插入 Expand 调用（约 3 行）。
4. **`app/v3/lifecycle_test.go`** — 加一个 `TestWithYamlConfig_ExpandsEnvPlaceholders`：临时配置写 `db_password: ${DB_PASSWORD}`，`t.Setenv` 设值，断言解码后字段 == 该值；再一个未设场景断言 `ReadConfig` 报错。
5. **`app/v3/example/`** — `main.go` 的 `MyConfig` 加 `DBPassword string \`yaml:"db_password"\``；`config/app.yaml` 写 `db_password: ${DB_PASSWORD:-dev-secret}`；`.env` 加 `DB_PASSWORD=prod-secret`。演示占位符 + env 兜底。
6. **`app/v3/README.md`** — "配置加载"能力加一行说明占位符插值；能力清单/API 速查加 `Expand` + 语法说明；差异表加一条。

## 不做
- 不支持 `$VAR` 无花括号形式（避免歧义，文档明确）。
- 不支持命令替换 / 算术（不是 shell）。
- 不改变 `WithDotEnvConfig` 行为（它已注入 env，Expand 自然读到）。
- 不为宽松模式加 option（用户已选"默认严格"，保持 API 简单；若以后需要再加 `ReadConfigOption`）。

## 验证
- `go vet ./...`、`gofmt -l`、`go build ./...`、`go test ./... -cover`（覆盖率应仍 ≥74%）。
- example 实跑：`db_password` 被替换为 `.env` 里的 `prod-secret`，`/ping` 与 SIGTERM 优雅退出仍正常。
