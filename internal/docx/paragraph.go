package docx

import (
	"bytes"
	"fmt"
	"strings"
)

// Paragraph 段落
// ParagraphChild 段落子元素接口
type ParagraphChild interface {
	ToXML() string
}

// RunContainer 运行容器接口
type RunContainer interface {
	AddRun(text string) *Run
	AddFormattedRun(text string, bold, italic, code bool) *Run
	AddImageRun(relID string, width, height int64) *Run
}

// Paragraph 段落
type Paragraph struct {
	StyleID  string
	Children []ParagraphChild

	Align           string // left, center, right, justify
	Indent          int    // 缩进(twips)
	SpacingB        int    // 段后间距
	SpacingA        int    // 段前间距
	Shading         string // 背景色
	Border          bool   // 是否添加边框
	HorizontalRule  bool   // 是否是分隔线
	LineHeight      int    // 行高 (twips)
	FirstLineIndent int    // 首行缩进 (twips)
}

// Run 文本运行
type Run struct {
	Text        string
	Bold        bool
	Italic      bool
	Underline   bool
	Strike      bool
	FontName    string
	FontSize    float64
	Color       string
	Highlight   string
	IsCode      bool
	IsImage     bool
	ImageRelID  string
	ImageWidth  int64 // EMUs (English Metric Units)
	ImageHeight int64
}

// NewParagraph 创建新段落
func NewParagraph(styleID string) *Paragraph {
	return &Paragraph{
		StyleID:  styleID,
		Children: make([]ParagraphChild, 0),
	}
}

// AddRun 添加文本运行
func (p *Paragraph) AddRun(text string) *Run {
	run := &Run{Text: text}
	p.Children = append(p.Children, run)
	return run
}

// AddFormattedRun 添加格式化文本运行
func (p *Paragraph) AddFormattedRun(text string, bold, italic, code bool) *Run {
	run := &Run{
		Text:   text,
		Bold:   bold,
		Italic: italic,
		IsCode: code,
	}
	p.Children = append(p.Children, run)
	return run
}

// AddImageRun 添加图片运行
func (p *Paragraph) AddImageRun(relID string, width, height int64) *Run {
	run := &Run{
		IsImage:     true,
		ImageRelID:  relID,
		ImageWidth:  width,
		ImageHeight: height,
	}
	p.Children = append(p.Children, run)
	return run
}

// Hyperlink 超链接
type Hyperlink struct {
	ID   string
	Runs []*Run
}

// AddRun 添加文本运行
func (h *Hyperlink) AddRun(text string) *Run {
	run := &Run{Text: text}
	h.Runs = append(h.Runs, run)
	return run
}

// AddFormattedRun 添加格式化文本运行
func (h *Hyperlink) AddFormattedRun(text string, bold, italic, code bool) *Run {
	run := &Run{
		Text:   text,
		Bold:   bold,
		Italic: italic,
		IsCode: code,
	}
	h.Runs = append(h.Runs, run)
	return run
}

// AddImageRun 添加图片运行
func (h *Hyperlink) AddImageRun(relID string, width, height int64) *Run {
	run := &Run{
		IsImage:     true,
		ImageRelID:  relID,
		ImageWidth:  width,
		ImageHeight: height,
	}
	h.Runs = append(h.Runs, run)
	return run
}

// ToXML 转换为XML
func (h *Hyperlink) ToXML() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(`<w:hyperlink r:id="%s">`, h.ID))
	for _, run := range h.Runs {
		buf.WriteString(run.ToXML())
	}
	buf.WriteString(`</w:hyperlink>`)
	return buf.String()
}

// AddHyperlink 添加超链接
func (p *Paragraph) AddHyperlink(id string) *Hyperlink {
	link := &Hyperlink{
		ID:   id,
		Runs: make([]*Run, 0),
	}
	p.Children = append(p.Children, link)
	return link
}

// ToXML 转换为XML
func (p *Paragraph) ToXML() string {
	var buf bytes.Buffer

	buf.WriteString(`
        <w:p>`)

	// 段落属性
	if p.StyleID != "" || p.Align != "" || p.Indent > 0 || p.SpacingB > 0 || p.SpacingA > 0 || p.Shading != "" || p.Border || p.HorizontalRule || p.LineHeight > 0 || p.FirstLineIndent > 0 {
		buf.WriteString(`
            <w:pPr>`)
		if p.StyleID != "" {
			buf.WriteString(`
                <w:pStyle w:val="` + p.StyleID + `"/>`)
		}
		if p.Align != "" {
			jc := p.Align
			if jc == "left" {
				jc = "start"
			} else if jc == "right" {
				jc = "end"
			}
			buf.WriteString(`
                <w:jc w:val="` + jc + `"/>`)
		}
		if p.Indent > 0 || p.FirstLineIndent > 0 {
			buf.WriteString(fmt.Sprintf(`
                <w:ind w:left="%d" w:firstLine="%d"/>`, p.Indent, p.FirstLineIndent))
		}
		if p.SpacingB > 0 || p.SpacingA > 0 || p.LineHeight > 0 {
			line := 360
			if p.LineHeight > 0 {
				line = p.LineHeight
			}
			buf.WriteString(fmt.Sprintf(`
                <w:spacing w:before="%d" w:after="%d" w:line="%d" w:lineRule="auto"/>`, p.SpacingA, p.SpacingB, line))
		}
		if p.Shading != "" {
			shading := strings.TrimPrefix(p.Shading, "#")
			buf.WriteString(`
                <w:shd w:val="clear" w:color="auto" w:fill="` + shading + `"/>`)
		}
		if p.HorizontalRule {
			buf.WriteString(`
                <w:pBdr>
                    <w:bottom w:val="single" w:sz="6" w:space="1" w:color="A0A0A0"/>
                </w:pBdr>`)
		}
		if p.Border {
			buf.WriteString(`
                <w:pBdr>
                    <w:top w:val="single" w:sz="4" w:space="1" w:color="C0C0C0"/>
                    <w:left w:val="single" w:sz="4" w:space="4" w:color="C0C0C0"/>
                    <w:bottom w:val="single" w:sz="4" w:space="1" w:color="C0C0C0"/>
                    <w:right w:val="single" w:sz="4" w:space="4" w:color="C0C0C0"/>
                </w:pBdr>`)
		}
		buf.WriteString(`
            </w:pPr>`)
	}

	// 运行
	for _, child := range p.Children {
		buf.WriteString(child.ToXML())
	}

	buf.WriteString(`
        </w:p>`)

	return buf.String()
}

// ToXML 运行转换为XML
func (r *Run) ToXML() string {
	var buf bytes.Buffer

	buf.WriteString(`
            <w:r>`)

	// 运行属性
	if r.Bold || r.Italic || r.Underline || r.Strike || r.FontName != "" || r.FontSize > 0 || r.Color != "" || r.IsCode {
		buf.WriteString(`
                <w:rPr>`)

		if r.FontName != "" {
			buf.WriteString(`
                    <w:rFonts w:ascii="` + r.FontName + `" w:eastAsia="` + r.FontName + `" w:hAnsi="` + r.FontName + `"/>`)
		}
		if r.FontSize > 0 {
			sz := int(r.FontSize * 2)
			buf.WriteString(fmt.Sprintf(`
                    <w:sz w:val="%d"/>
                    <w:szCs w:val="%d"/>`, sz, sz))
		}
		if r.Bold {
			buf.WriteString(`
                    <w:b/>`)
		}
		if r.Italic {
			buf.WriteString(`
                    <w:i/>`)
		}
		if r.Underline {
			buf.WriteString(`
                    <w:u w:val="single"/>`)
		}
		if r.Strike {
			buf.WriteString(`
                    <w:strike/>`)
		}
		if r.Color != "" {
			color := strings.TrimPrefix(r.Color, "#")
			buf.WriteString(`
                    <w:color w:val="` + color + `"/>`)
		}
		if r.IsCode {
			buf.WriteString(`
                    <w:rFonts w:ascii="Consolas" w:hAnsi="Consolas"/>
                    <w:shd w:val="clear" w:color="auto" w:fill="E8E8E8"/>`)
		}

		buf.WriteString(`
                </w:rPr>`)
	}

	// 内容
	if r.IsImage {
		buf.WriteString(fmt.Sprintf(`
                <w:drawing>
                    <wp:inline distT="0" distB="0" distL="0" distR="0">
                        <wp:extent cx="%d" cy="%d"/>
                        <wp:effectExtent l="0" t="0" r="0" b="0"/>
                        <wp:docPr id="1" name="Picture"/>
                        <wp:cNvGraphicFramePr>
                            <a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/>
                        </wp:cNvGraphicFramePr>
                        <a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
                            <a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">
                                <pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture">
                                    <pic:nvPicPr>
                                        <pic:cNvPr id="0" name="Picture"/>
                                        <pic:cNvPicPr/>
                                    </pic:nvPicPr>
                                    <pic:blipFill>
                                        <a:blip r:embed="%s"/>
                                        <a:stretch>
                                            <a:fillRect/>
                                        </a:stretch>
                                    </pic:blipFill>
                                    <pic:spPr>
                                        <a:xfrm>
                                            <a:off x="0" y="0"/>
                                            <a:ext cx="%d" cy="%d"/>
                                        </a:xfrm>
                                        <a:prstGeom prst="rect">
                                            <a:avLst/>
                                        </a:prstGeom>
                                    </pic:spPr>
                                </pic:pic>
                            </a:graphicData>
                        </a:graphic>
                    </wp:inline>
                </w:drawing>`, r.ImageWidth, r.ImageHeight, r.ImageRelID, r.ImageWidth, r.ImageHeight))
	} else if r.Text != "" {
		escapedText := XMLEscape(r.Text)
		// 处理换行和空格
		lines := strings.Split(escapedText, "\n")
		for i, line := range lines {
			if i > 0 {
				buf.WriteString(`
                <w:br/>`)
			}
			if strings.HasPrefix(line, " ") || strings.HasSuffix(line, " ") || strings.Contains(line, "  ") {
				buf.WriteString(`
                <w:t xml:space="preserve">` + line + `</w:t>`)
			} else {
				buf.WriteString(`
                <w:t>` + line + `</w:t>`)
			}
		}
	}

	buf.WriteString(`
            </w:r>`)

	return buf.String()
}
