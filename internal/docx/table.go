package docx

import (
	"bytes"
	"fmt"
)

// Table 表格
type Table struct {
	Rows       []*TableRow
	ColWidths  []int // 列宽(twips)
	HasBorders bool
}

// TableRow 表格行
type TableRow struct {
	Cells    []*TableCell
	IsHeader bool
}

// TableCell 表格单元格
type TableCell struct {
	Paragraphs []*Paragraph
	Width      int    // 单元格宽度(twips)
	Align      string // left, center, right
	VAlign     string // top, center, bottom
	Shading    string // 背景色
}

// NewTable 创建新表格
func NewTable() *Table {
	return &Table{
		Rows:       make([]*TableRow, 0),
		ColWidths:  make([]int, 0),
		HasBorders: true,
	}
}

// AddRow 添加行
func (t *Table) AddRow(isHeader bool) *TableRow {
	row := &TableRow{
		Cells:    make([]*TableCell, 0),
		IsHeader: isHeader,
	}
	t.Rows = append(t.Rows, row)
	return row
}

// AddCell 向行添加单元格
func (r *TableRow) AddCell() *TableCell {
	cell := &TableCell{
		Paragraphs: make([]*Paragraph, 0),
	}
	r.Cells = append(r.Cells, cell)
	return cell
}

// AddParagraph 向单元格添加段落
func (c *TableCell) AddParagraph(p *Paragraph) {
	c.Paragraphs = append(c.Paragraphs, p)
}

// SetText 设置单元格文本
func (c *TableCell) SetText(text string, bold bool) {
	p := NewParagraph("")
	run := p.AddRun(text)
	run.Bold = bold
	c.Paragraphs = append(c.Paragraphs, p)
}

// ToXML 表格转换为XML
func (t *Table) ToXML() string {
	var buf bytes.Buffer

	buf.WriteString(`
        <w:tbl>
            <w:tblPr>
                <w:tblStyle w:val="TableGrid"/>
                <w:tblW w:w="0" w:type="auto"/>
                <w:tblLook w:val="04A0" w:firstRow="1" w:lastRow="0" w:firstColumn="1" w:lastColumn="0" w:noHBand="0" w:noVBand="1"/>`)

	if t.HasBorders {
		buf.WriteString(`
                <w:tblBorders>
                    <w:top w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                    <w:left w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                    <w:bottom w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                    <w:right w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                    <w:insideH w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                    <w:insideV w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                </w:tblBorders>`)
	}

	buf.WriteString(`
            </w:tblPr>`)

	// 列宽定义
	if len(t.ColWidths) > 0 {
		buf.WriteString(`
            <w:tblGrid>`)
		for _, w := range t.ColWidths {
			buf.WriteString(fmt.Sprintf(`
                <w:gridCol w:w="%d"/>`, w))
		}
		buf.WriteString(`
            </w:tblGrid>`)
	}

	// 行
	for _, row := range t.Rows {
		buf.WriteString(`
            <w:tr>`)

		if row.IsHeader {
			buf.WriteString(`
                <w:trPr>
                    <w:tblHeader/>
                </w:trPr>`)
		}

		for _, cell := range row.Cells {
			buf.WriteString(`
                <w:tc>
                    <w:tcPr>`)

			if cell.Width > 0 {
				buf.WriteString(fmt.Sprintf(`
                        <w:tcW w:w="%d" w:type="dxa"/>`, cell.Width))
			}

			if cell.Shading != "" {
				buf.WriteString(`
                        <w:shd w:val="clear" w:color="auto" w:fill="` + cell.Shading + `"/>`)
			}

			if cell.VAlign != "" {
				buf.WriteString(`
                        <w:vAlign w:val="` + cell.VAlign + `"/>`)
			}

			buf.WriteString(`
                    </w:tcPr>`)

			if len(cell.Paragraphs) == 0 {
				// 空单元格也需要一个段落
				buf.WriteString(`
                    <w:p/>`)
			} else {
				for _, p := range cell.Paragraphs {
					if cell.Align != "" && p.Align == "" {
						p.Align = cell.Align
					}
					buf.WriteString(p.ToXML())
				}
			}

			buf.WriteString(`
                </w:tc>`)
		}

		buf.WriteString(`
            </w:tr>`)
	}

	buf.WriteString(`
        </w:tbl>`)

	return buf.String()
}

// TableElement 表格元素（用于添加到文档）
type TableElement struct {
	table *Table
}

// NewTableElement 创建表格元素
func NewTableElement(t *Table) *TableElement {
	return &TableElement{table: t}
}

// ToXML 表格元素转换为XML
func (te *TableElement) ToXML() string {
	return te.table.ToXML()
}
