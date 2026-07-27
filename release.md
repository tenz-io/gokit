# 发布流程

本项目含多个 Go 子模块，通过 `go.work` 管理。所有模块统一为 `v3` 主版本轨道，共享同一个 `v3.0.x` 版本号。发布时通过 `scripts/tag-all.sh` 一键批量打 tag 与 GitHub Release（全程经 `gh` API，不依赖 `origin` 指向 GitHub）。

## 版本轨道规则

- **v3 轨道**：所有 `/v3` 模块共享同一个 `v3.0.x` 版本号。
- tag 前缀 = module path 去掉末尾 `/vN`：`annotation/v3` → tag `annotation/v3.0.x`，`logger/v3` → tag `logger/v3.0.x`。

## 一键批量发布

### 前提条件

- **[GitHub CLI](https://cli.github.com)** 已安装并登录：`brew install gh && gh auth login`
- 对目标仓库有 push 权限（默认 `tenz-io/gokit`，可用 `--repo` 覆盖）
- 脚本会自动执行 `gh auth setup-git`，让对 github.com 的 git 操作复用 gh token，无需手动添加 remote

> `gh release create <tag>` 在 tag 不存在时会自动在仓库默认分支上创建该 tag，因此 **tag 与 Release 在一条命令里同时打出**，无需先 `git push --tags`。

### 本地执行（一键 tag + release）

```bash
# 0. 发版前一致性校验（内部 require 版本一致、无遗漏）
./scripts/version-check.sh

# 1. 一键：为所有模块打 tag + 创建 GitHub Release
./scripts/tag-all.sh v3.0.1 --release

# 等价于：make release VERSION=v3.0.1   （内含 version-check + make test）
make release VERSION=v3.0.1

# 预览（不执行任何网络操作）
./scripts/tag-all.sh v3.0.1 --release --dry-run

# 只打 tag、不建 Release
./scripts/tag-all.sh v3.0.1 --push

# 只给特定模块
./scripts/tag-all.sh v3.0.1 --release annotation/v3,logger/v3

# 自定义 Release Notes（对所有模块共用）
./scripts/tag-all.sh v3.0.1 --release --notes-from-file NOTES.md

# tag 已存在时跳过，而非删后重建（默认是覆盖重建）
./scripts/tag-all.sh v3.0.1 --release --no-overwrite

# 指定其它 GitHub 仓库
./scripts/tag-all.sh v3.0.1 --release --repo owner/repo
```

#### 冲突策略

默认 **覆盖重建**：若某 tag 已存在，先 `gh release delete --cleanup-tag` 删掉旧 release+tag，再重建。`--no-overwrite` 则改为跳过已存在的 tag。

### GitHub Actions

1. 打开 [Actions → Tag Release](https://github.com/tenz-io/gokit/actions/workflows/tag-release.yml)
2. 点击 **Run workflow**
3. 填写参数：
   - `version`：版本号，如 `v3.0.1`（必填）
   - `modules`：限定模块（逗号分隔），留空 = 全部非 example 模块
   - `dry_run`：`true` = 仅预览不执行
4. 点击 **Run workflow**

> CI 路径走 `git push origin --tags`（CI 里 origin 即 GitHub），与本地 `gh` 路径独立、互不影响。

## Tag 与 GitHub Release 的区别

| | Git Tag | GitHub Release |
|---|---|---|
| 是什么 | 指向特定 commit 的引用 | 基于 tag + Release Notes + 附件 |
| 命令行 | `git tag` / `gh release create` | `gh release create` |
| 是否必须 | Go `go get` 需要 tag | 可选，方便人类阅读 changelog |
| 查看位置 | `https://github.com/<repo>/tags` | `https://github.com/<repo>/releases` |
| 本仓库 | `--release` 一键同时创建 | `--release` 一键同时创建（`--push` 仅 tag、Release 体为空） |

**注意**：本仓库 `gh release create` 在 tag 不存在时自动建 tag，故 tag 与 Release 同步出现。`--push` 模式走同一条命令但 release notes 为空，只起"打 tag"作用；要正式 Release 请用 `--release`。

## Tag 命名规则

tag 前缀 = module path 去掉末尾 `/vN`：

| go.work 路径 | Module 路径 | Tag 示例 |
|---|---|---|
| `./annotation/v3` | `github.com/tenz-io/gokit/annotation/v3` | `annotation/v3.0.1` |
| `./logger/v3` | `github.com/tenz-io/gokit/logger/v3` | `logger/v3.0.1` |
| `./cache/v3` | `github.com/tenz-io/gokit/cache/v3` | `cache/v3.0.1` |

规则：脚本将 go.work 路径中的 `/vN` 后缀移除作为 tag 前缀。所有模块均为 `v3`。

## 子模块清单

以下为 `go.work` 中所有非 example 子模块：

| # | 子模块 | Module Path |
|---|--------|-------------|
| 1 | `annotation` | `github.com/tenz-io/gokit/annotation/v3` |
| 2 | `app` | `github.com/tenz-io/gokit/app/v3` |
| 3 | `async` | `github.com/tenz-io/gokit/async/v3` |
| 4 | `cache` | `github.com/tenz-io/gokit/cache/v3` |
| 5 | `collection` | `github.com/tenz-io/gokit/collection/v3` |
| 6 | `functional` | `github.com/tenz-io/gokit/functional/v3` |
| 7 | `ginext` | `github.com/tenz-io/gokit/ginext/v3` |
| 8 | `gormext` | `github.com/tenz-io/gokit/gormext/v3` |
| 9 | `httpext` | `github.com/tenz-io/gokit/httpext/v3` |
| 10 | `logger` | `github.com/tenz-io/gokit/logger/v3` |
| 11 | `monitor` | `github.com/tenz-io/gokit/monitor/v3` |
| 12 | `retriever` | `github.com/tenz-io/gokit/retriever/v3` |
| 13 | `tracer` | `github.com/tenz-io/gokit/tracer/v3` |

## 发布后

Go 模块消费者通过 `go get` 拉取新版本：

```bash
go get github.com/tenz-io/gokit/logger/v3@v3.0.1
go get github.com/tenz-io/gokit/tracer/v3@v3.0.1
go get github.com/tenz-io/gokit/annotation/v3@v3.0.1
go get github.com/tenz-io/gokit/cache/v3@v3.0.1
```

消费者 `go.mod` 中版本引用示例：

```
require (
    github.com/tenz-io/gokit/logger/v3     v3.0.1
    github.com/tenz-io/gokit/tracer/v3     v3.0.1
    github.com/tenz-io/gokit/annotation/v3 v3.0.1
    github.com/tenz-io/gokit/cache/v3      v3.0.1
)
```

## 注意事项

- **tag 不可变**：Git tag 一旦推送不可修改（如需修正，需打新版本号）
- **Go 模块版本规则**：`v2+` 的 major version 要求 module path 以 `/vN` 结尾，本仓库所有模块已统一为 `v3`，module path 均以 `/v3` 结尾
- **Git tag ≠ GitHub Release**：tag 推送到 GitHub 后，需执行 `gh release create` 或在 Tags 页手动创建 Release
- **go.work 不参与发布**：`go.work` 仅用于本地开发，外部消费者通过 `go get` + tag 拉取
- **example 模块不打 tag**：`go.work` 中的 `*/example` 目录仅用于本地示例
- **gh CLI 认证**：使用 `--release` 前确保执行过 `gh auth login`，否则会报错
