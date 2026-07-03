package parser

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// maxATXHeadingLevel 是本解析器支持的最高 ATX 标题层级。
// goldmark 内置 atxHeadingParser 把上限写死为 6，这里放宽到 9，
// 让 7-9 级 `#` 标题也能产出 Heading 节点，进而在 docx 中应用 Heading7/8/9 样式。
const maxATXHeadingLevel = 9

// deepATXHeadingParser 与 goldmark 内置的 atxHeadingParser 行为一致，
// 仅把层级上限从 6 放宽到 maxATXHeadingLevel。
type deepATXHeadingParser struct {
	parser.HeadingConfig
}

// NewDeepATXHeadingParser 返回一个支持 1-9 级的 ATX Heading BlockParser。
func NewDeepATXHeadingParser(opts ...parser.HeadingOption) parser.BlockParser {
	p := &deepATXHeadingParser{}
	for _, o := range opts {
		o.SetHeadingOption(&p.HeadingConfig)
	}
	return p
}

func (b *deepATXHeadingParser) Trigger() []byte {
	return []byte{'#'}
}

func (b *deepATXHeadingParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	pos := pc.BlockOffset()
	if pos < 0 {
		return nil, parser.NoChildren
	}

	// 统计连续的 `#` 个数作为 level
	i := pos
	for ; i < len(line) && line[i] == '#'; i++ {
	}
	level := i - pos
	if i == pos || level > maxATXHeadingLevel {
		return nil, parser.NoChildren
	}
	if i == len(line) {
		// 整行只有 `#`, 如 `###`
		return ast.NewHeading(level), parser.NoChildren
	}

	// `#` 后必须有空白分隔
	l := util.TrimLeftSpaceLength(line[i:])
	if l == 0 {
		return nil, parser.NoChildren
	}

	start := i + l
	if start >= len(line) {
		start = len(line) - 1
	}
	stop := len(line) - util.TrimRightSpaceLength(line)

	node := ast.NewHeading(level)

	// 剥离尾部可选的 `###` 闭合标记（如 `## 标题 ##`）
	if stop > start {
		j := stop - 1
		for j >= start && line[j] == '#' {
			j--
		}
		// 如果尾部 `#` 紧贴文字（如 `##abc##`），则视为正文的一部分，不剥离
		if j != stop-1 && !util.IsSpace(line[j]) {
			j = stop - 1
		}
		stop = j + 1
	}

	if len(util.TrimRight(line[start:stop], []byte{'#'})) != 0 {
		node.Lines().Append(text.NewSegment(
			segment.Start+start-segment.Padding,
			segment.Start+stop-segment.Padding,
		))
	}
	return node, parser.NoChildren
}

func (b *deepATXHeadingParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *deepATXHeadingParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	if !b.AutoHeadingID {
		return
	}
	heading, ok := node.(*ast.Heading)
	if !ok || heading.Lines().Len() == 0 {
		return
	}
	seg := heading.Lines().At(heading.Lines().Len() - 1)
	lastLine := seg.Value(reader.Source())
	heading.SetAttributeString("id", string(pc.IDs().Generate(lastLine, ast.KindHeading)))
}

func (b *deepATXHeadingParser) CanInterruptParagraph() bool {
	return true
}

func (b *deepATXHeadingParser) CanAcceptIndentedLine() bool {
	return false
}
