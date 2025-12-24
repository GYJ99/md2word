package converter

import (
	"strings"

	"md2word/internal/docx"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightCodeNative 使用Chroma将代码转换为具有高亮效果的DOCX段落并添加到单元格中
func HighlightCodeNative(cell *docx.TableCell, code, language, fontName string, fontSize float64, lineSpacing, lineHeight int) error {
	// 获取lexer
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// 获取样式
	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}

	// 迭代代码
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return err
	}

	// 创建初始段落
	p := docx.NewParagraph("")
	p.SpacingA = lineSpacing / 2
	p.SpacingB = lineSpacing / 2
	p.LineHeight = lineHeight
	cell.AddParagraph(p)

	for _, token := range iterator.Tokens() {
		entry := style.Get(token.Type)

		// 处理包含换行的 token
		lines := strings.Split(token.Value, "\n")
		for i, lineText := range lines {
			if i > 0 {
				// 换行，创建新段落
				p = docx.NewParagraph("")
				p.SpacingA = lineSpacing / 2
				p.SpacingB = lineSpacing / 2
				p.LineHeight = lineHeight
				cell.AddParagraph(p)
			}

			if lineText != "" {
				run := p.AddRun(lineText)
				run.FontName = fontName
				run.FontSize = fontSize

				// 映射Chroma颜色到RGB
				if entry.Colour.IsSet() {
					run.Color = strings.TrimPrefix(entry.Colour.String(), "#")
				}
				if entry.Bold == chroma.Yes {
					run.Bold = true
				}
				if entry.Italic == chroma.Yes {
					run.Italic = true
				}
			}
		}
	}

	return nil
}
