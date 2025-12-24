package parser

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// MarkdownParser Markdown解析器
type MarkdownParser struct {
	md goldmark.Markdown
}

// NewMarkdownParser 创建新的解析器
func NewMarkdownParser() *MarkdownParser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,           // GitHub Flavored Markdown
			extension.Table,         // 表格支持
			extension.Strikethrough, // 删除线
			extension.TaskList,      // 任务列表
			extension.Typographer,   // 排版优化
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
	return &MarkdownParser{md: md}
}

// Parse 解析Markdown内容为AST
func (p *MarkdownParser) Parse(content []byte) ast.Node {
	reader := text.NewReader(content)
	return p.md.Parser().Parse(reader)
}

// GetParser 获取底层Goldmark的Parser
func (p *MarkdownParser) GetParser() parser.Parser {
	return p.md.Parser()
}

// Render 渲染Markdown为HTML（用于测试）
func (p *MarkdownParser) Render(content []byte) (string, error) {
	var buf bytes.Buffer
	if err := p.md.Convert(content, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
