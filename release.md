# 发布流程

本项目含多个 Go 子模块，通过 `go.work` 管理。子模块按 major 分版本轨道：大部分是 `v2`，`annotation` 已升级到 `v3`。发布时通过 `scripts/tag-all.sh` 一键批量打 tag，可同时推进两条版本轨道。

## 版本轨道规则

- **v2 轨道**：所有 `/v2` 模块共享同一个 `v2.0.x` 版本号。
- **v3 轨道**：`annotation/v3` 独立走 `v3.0.x`；引用它的模块（`ginext`、`app`、`cmd`）的 `require annotation/v3` 版本必须与 v3 tag 一致。
- tag 前缀 = module path 去掉末尾 `/vN`：`annotation/v3` → tag `annotation/v3.0.x`，`logger/v2` → tag `logger/v2.0.x`（两条轨道前缀不同，不冲突）。

## 一键批量发布

### 前提条件

- `git push` 权限
- （可选）[GitHub CLI](https://cli.github.com) (`brew install gh` + `gh auth login`)，用于创建 Release

### 本地执行

```bash
# 0. 发版前一致性校验（内部 require 版本一致、无遗漏）
./scripts/version-check.sh

# 1. 确认所有测试通过
make test

# 2. 预览即将创建的 tag（同时推进 v2/v3 两条轨道）
./scripts/tag-all.sh v2.0.1 v3.0.1 --dry-run

# 3. 打 tag（仅本地）
./scripts/tag-all.sh v2.0.1 v3.0.1

# 4. 打 tag + 推送到 GitHub
./scripts/tag-all.sh v2.0.1 v3.0.1 --push

# 5. 打 tag + 推送 + 创建 GitHub Release（需要 gh CLI）
./scripts/tag-all.sh v2.0.1 v3.0.1 --release

# 只发 v2 轨道（annotation/v3 会被跳过）
./scripts/tag-all.sh v2.0.1 --push

# 只发 v3 轨道
./scripts/tag-all.sh v3.0.1 --push

# 只给特定模块打 tag
./scripts/tag-all.sh v2.0.1 annotation/v3,logger/v2 --push
```

版本号按 `v2.`/`v3.` 前缀自动归到对应轨道，顺序无关：`./scripts/tag-all.sh v3.0.1 v2.0.1` 等价于上面。

### GitHub Actions

1. 打开 [Actions → Tag Release](https://github.com/tenz-io/gokit/actions/workflows/tag-release.yml)
2. 点击 **Run workflow**
3. 填写参数：
   - `version`：版本号，如 `v2.0.1`（必填）
   - `modules`：限定模块（逗号分隔），留空 = 全部非 example 模块
   - `dry_run`：`true` = 仅预览不执行
4. 点击 **Run workflow**

## Tag 与 GitHub Release 的区别

| | Git Tag | GitHub Release |
|---|---|---|
| 是什么 | 指向特定 commit 的引用 | 基于 tag + Release Notes + 附件 |
| 命令行 | `git tag -a` / `git push --tags` | `gh release create` |
| 是否必须 | Go `go get` 需要 tag | 可选，方便人类阅读 changelog |
| 查看位置 | `https://github.com/<repo>/tags` | `https://github.com/<repo>/releases` |
| 本仓库 | `--push` 创建 | `--release` 创建 |

**注意**：仅推送 tag 不会出现在 GitHub Releases 页面，必须在 Tags 页手动 "Create release" 或使用 `--release` 模式。

## Tag 命名规则

tag 前缀 = module path 去掉末尾 `/vN`：

| go.work 路径 | Module 路径 | Tag 示例 |
|---|---|---|
| `./annotation/v3` | `github.com/tenz-io/gokit/annotation/v3` | `annotation/v3.0.1` |
| `./logger/v2` | `github.com/tenz-io/gokit/logger/v2` | `logger/v2.0.1` |
| `./cache/v2` | `github.com/tenz-io/gokit/cache/v2` | `cache/v2.0.1` |

规则：脚本将 go.work 路径中的 `/vN` 后缀移除作为 tag 前缀。`annotation` 已升级到 `v3`，其余模块为 `v2`，两条版本轨道前缀不同、互不冲突。

## 子模块清单

以下为 `go.work` 中所有非 example 子模块：

| # | 子模块 | Module Path |
|---|--------|-------------|
| 1 | `annotation` (v3) | `github.com/tenz-io/gokit/annotation/v3` |
| 2 | `app` | `github.com/tenz-io/gokit/app/v2` |
| 3 | `async` | `github.com/tenz-io/gokit/async/v2` |
| 4 | `cache` | `github.com/tenz-io/gokit/cache/v2` |
| 5 | `cmd` | `github.com/tenz-io/gokit/cmd/v2` |
| 6 | `collection` | `github.com/tenz-io/gokit/collection/v2` |
| 7 | `functional` | `github.com/tenz-io/gokit/functional/v2` |
| 8 | `ginext` | `github.com/tenz-io/gokit/ginext/v2` |
| 9 | `gormext` | `github.com/tenz-io/gokit/gormext/v2` |
| 10 | `grpcext` | `github.com/tenz-io/gokit/grpcext/v2` |
| 11 | `httpext` | `github.com/tenz-io/gokit/httpext/v2` |
| 12 | `logger` | `github.com/tenz-io/gokit/logger/v2` |
| 13 | `monitor` | `github.com/tenz-io/gokit/monitor/v2` |
| 14 | `retriever` | `github.com/tenz-io/gokit/retriever/v2` |
| 15 | `tracer` | `github.com/tenz-io/gokit/tracer/v2` |

## 发布后

Go 模块消费者通过 `go get` 拉取新版本：

```bash
go get github.com/tenz-io/gokit/logger/v2@v2.0.1
go get github.com/tenz-io/gokit/tracer/v2@v2.0.1
go get github.com/tenz-io/gokit/annotation/v3@v3.0.1
go get github.com/tenz-io/gokit/cache/v2@v2.0.1
```

消费者 `go.mod` 中版本引用示例：

```
require (
    github.com/tenz-io/gokit/logger/v2     v2.0.1
    github.com/tenz-io/gokit/tracer/v2     v2.0.1
    github.com/tenz-io/gokit/annotation/v3 v3.0.1
    github.com/tenz-io/gokit/cache/v2      v2.0.1
)
```

## 注意事项

- **tag 不可变**：Git tag 一旦推送不可修改（如需修正，需打新版本号）
- **Go 模块版本规则**：`v2+` 版本要求 module path 以 `/v2` 结尾，本仓库所有模块已统一
- **Git tag ≠ GitHub Release**：tag 推送到 GitHub 后，需执行 `gh release create` 或在 Tags 页手动创建 Release
- **go.work 不参与发布**：`go.work` 仅用于本地开发，外部消费者通过 `go get` + tag 拉取
- **example 模块不打 tag**：`go.work` 中的 `*/example` 目录仅用于本地示例
- **gh CLI 认证**：使用 `--release` 前确保执行过 `gh auth login`，否则会报错
