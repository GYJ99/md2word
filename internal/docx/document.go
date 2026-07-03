package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"md2word/internal/config"
)

// Element 文档元素接口
type Element interface {
	ToXML() string
}

// Document DOCX文档
type Document struct {
	config         *config.Config
	elements       []Element
	images         map[string]*ImageData
	imageCount     int
	rels           []Relationship
	contentRels    []Relationship
	numberingState *NumberingState
}

// ImageData 图片数据
type ImageData struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}

// Relationship 关系定义
type Relationship struct {
	ID         string
	Type       string
	Target     string
	TargetMode string
}

// NewDocument 创建新文档
func NewDocument(cfg *config.Config) *Document {
	return &Document{
		config:         cfg,
		elements:       make([]Element, 0),
		images:         make(map[string]*ImageData),
		rels:           make([]Relationship, 0),
		numberingState: NewNumberingState(),
	}
}

// AddParagraph 添加段落
func (d *Document) AddParagraph(p Element) {
	d.elements = append(d.elements, p)
}

// AddImage 添加图片并返回关系ID
func (d *Document) AddImage(data []byte, contentType string, width, height int) string {
	d.imageCount++
	rID := fmt.Sprintf("rId%d", d.imageCount+10) // 预留前10个ID给其他关系
	imgName := fmt.Sprintf("image%d", d.imageCount)

	ext := ".png"
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/gif":
		ext = ".gif"
	case "image/svg+xml":
		ext = ".svg"
	}

	d.images[imgName+ext] = &ImageData{
		Data:        data,
		ContentType: contentType,
		Width:       width,
		Height:      height,
	}

	d.contentRels = append(d.contentRels, Relationship{
		ID:     rID,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image",
		Target: "media/" + imgName + ext,
	})

	return rID
}

// AddHyperlink 添加超链接关系并返回ID
func (d *Document) AddHyperlink(target string) string {
	d.imageCount++ // 复用计数器或独立计数
	rID := fmt.Sprintf("rId%d", d.imageCount+1000)

	d.contentRels = append(d.contentRels, Relationship{
		ID:         rID,
		Type:       "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink",
		Target:     target,
		TargetMode: "External",
	})

	return rID
}

// GetNumberingState 获取编号状态
func (d *Document) GetNumberingState() *NumberingState {
	return d.numberingState
}

// Save 保存为DOCX文件
func (d *Document) Save(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建zip文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	// 写入[Content_Types].xml
	if err := d.writeContentTypes(w); err != nil {
		return err
	}

	// 写入_rels/.rels
	if err := d.writeRels(w); err != nil {
		return err
	}

	// 写入word/_rels/document.xml.rels
	if err := d.writeDocumentRels(w); err != nil {
		return err
	}

	// 写入word/styles.xml
	if err := d.writeStyles(w); err != nil {
		return err
	}

	// 写入word/numbering.xml
	if err := d.writeNumbering(w); err != nil {
		return err
	}

	// 写入word/document.xml
	if err := d.writeDocument(w); err != nil {
		return err
	}

	// 写入图片
	for name, img := range d.images {
		if err := d.writeImage(w, name, img); err != nil {
			return err
		}
	}

	return nil
}

// writeContentTypes 写入内容类型定义
func (d *Document) writeContentTypes(w *zip.Writer) error {
	f, err := w.Create("[Content_Types].xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
    <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
    <Default Extension="xml" ContentType="application/xml"/>
    <Default Extension="png" ContentType="image/png"/>
    <Default Extension="jpg" ContentType="image/jpeg"/>
    <Default Extension="jpeg" ContentType="image/jpeg"/>
    <Default Extension="gif" ContentType="image/gif"/>
    <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
    <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
    <Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>
</Types>`
	_, err = io.WriteString(f, content)
	return err
}

// writeRels 写入根关系
func (d *Document) writeRels(w *zip.Writer) error {
	f, err := w.Create("_rels/.rels")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	_, err = io.WriteString(f, content)
	return err
}

// writeDocumentRels 写入文档关系
func (d *Document) writeDocumentRels(w *zip.Writer) error {
	f, err := w.Create("word/_rels/document.xml.rels")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
    <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>`)

	for _, rel := range d.contentRels {
		if rel.TargetMode != "" {
			buf.WriteString(fmt.Sprintf(`
    <Relationship Id="%s" Type="%s" Target="%s" TargetMode="%s"/>`, rel.ID, rel.Type, rel.Target, rel.TargetMode))
		} else {
			buf.WriteString(fmt.Sprintf(`
    <Relationship Id="%s" Type="%s" Target="%s"/>`, rel.ID, rel.Type, rel.Target))
		}
	}

	buf.WriteString(`
</Relationships>`)

	_, err = f.Write(buf.Bytes())
	return err
}

// writeStyles 写入样式定义
func (d *Document) writeStyles(w *zip.Writer) error {
	f, err := w.Create("word/styles.xml")
	if err != nil {
		return err
	}

	styles := GenerateStyles(d.config)
	_, err = io.WriteString(f, styles)
	return err
}

// writeNumbering 写入编号定义
func (d *Document) writeNumbering(w *zip.Writer) error {
	f, err := w.Create("word/numbering.xml")
	if err != nil {
		return err
	}

	numbering := GenerateNumberingXML(d.numberingState.GetNumberingInstances())
	_, err = io.WriteString(f, numbering)
	return err
}

// writeDocument 写入文档内容
func (d *Document) writeDocument(w *zip.Writer) error {
	f, err := w.Create("word/document.xml")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"
            xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
            xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
            xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"
            xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
    <w:body>`)

	for _, elem := range d.elements {
		buf.WriteString(elem.ToXML())
	}

	buf.WriteString(fmt.Sprintf(`
        <w:sectPr>
            <w:pgSz w:w="%d" w:h="%d"/>
            <w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="851" w:footer="992" w:gutter="0"/>
        </w:sectPr>
    </w:body>
</w:document>`, PageWidthTwips, PageHeightTwips, MarginTop, MarginRight, MarginBottom, MarginLeft))

	_, err = f.Write(buf.Bytes())
	return err
}

// writeImage 写入图片文件
func (d *Document) writeImage(w *zip.Writer, name string, img *ImageData) error {
	f, err := w.Create("word/media/" + name)
	if err != nil {
		return err
	}
	_, err = f.Write(img.Data)
	return err
}

// XMLEscape 转义XML特殊字符
func XMLEscape(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// Word 页面布局常量（twips，1 inch = 1440 twips，1 cm ≈ 567 twips）
// 与 writeDocument 中 sectPr 的 pgSz/pgMar 保持一致。
const (
	PageWidthTwips  = 11906 // A4 宽度 ≈ 21cm
	PageHeightTwips = 16838 // A4 高度 ≈ 29.7cm
	MarginTop       = 1440  // 2.54cm
	MarginBottom    = 1440
	MarginLeft      = 1800  // 3.17cm
	MarginRight     = 1800
)

// ContentWidthTwips 返回页面内容区可用宽度（页面宽度 - 左右边距）
func ContentWidthTwips() int {
	return PageWidthTwips - MarginLeft - MarginRight
}

// ContentWidthPx 把内容宽度换算为 96DPI 下的像素数。
// 1 twip = 1/1440 inch，96 DPI 下 1 inch = 96 px，故 1 twip = 96/1440 px。
func ContentWidthPx() int {
	return ContentWidthTwips() * 96 / 1440
}

// ContentHeightPx 把内容区高度换算为 96DPI 下的像素数。
func ContentHeightPx() int {
	return (PageHeightTwips - MarginTop - MarginBottom) * 96 / 1440
}
