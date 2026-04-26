# notionx

将 Markdown 文本转换为 Notion API block 结构的工具库。

## 快速开始

```go
import "github.com/tenz-io/gokit/notionx"

md := `# 标题

这是一段**富文本**内容。

- 列表项 1
- 列表项 2

> 引用文字
`

blocks, err := notionx.MarkdownToNotionBlocks(md)
if err != nil {
    panic(err)
}
// blocks 可直接用于 Notion API 的页面追加操作
```

## API 参考

| 函数 | 签名 | 说明 |
|------|------|------|
| `MarkdownToNotionBlocks` | `func MarkdownToNotionBlocks(markdown string) ([]notionapi.Block, error)` | 将 Markdown 字符串解析为 Notion block 切片 |

## 支持的 Markdown 语法

| Markdown | Notion Block 类型 |
|----------|-------------------|
| `# heading` | `Heading1Block` |
| `## heading` | `Heading2Block` |
| `### heading` | `Heading3Block` |
| `#### heading+` (四级以上) | `ParagraphBlock`（降级处理） |
| 普通段落 | `ParagraphBlock` |
| `**bold**` `*italic*` | `RichText.Annotations.Bold/Italic` |
| `` `code` `` | `RichText.Annotations.Code` |
| ``` ```code block``` ``` | `CodeBlock` |
| `- item` / `* item` | `BulletedListItemBlock` |
| `1. item` | `NumberedListItemBlock` |
| `> quote` | `QuoteBlock` |

**不支持**：图片、表格、链接（URL 按纯文本处理）、嵌套列表、HTML 标签。不支持的节点类型默认降级为 `ParagraphBlock`。

## 最佳实践

### 空字符串

传入空字符串返回 `nil, nil`，不会 panic：

```go
blocks, _ := notionx.MarkdownToNotionBlocks("")
// blocks == nil
```

### 用于 Notion 页面创建

结合 Notion API 客户端将 Markdown 发布为页面内容：

```go
blocks, err := notionx.MarkdownToNotionBlocks(markdown)
if err != nil {
    return err
}
_, err = client.Block.AppendChildren(ctx, pageID, blocks)
```

## 与 v1 的区别

| 变更项 | v1 | v2 |
|--------|----|----|
| AST 遍历 | 递归深度优先，嵌套 walk 闭包 | `walkAST` 迭代式 + 独立 `convertNode` 分发 |
| 空输入 | 未处理 | 显式返回 nil |
| 代码块处理 | 文本内嵌 | `FencedCodeBlock` 正确映射为 `CodeBlock` |
| 列表项转换 | 仅提取第一个元素文本 | 所有子元素完整转换 |
| 富文本注解 | 仅 Bold/Italic | 支持 Bold、Italic、Code |
