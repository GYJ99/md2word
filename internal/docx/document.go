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
	config      *config.Config
	elements    []Element
	images      map[string]*ImageData
	imageCount  int
	rels        []Relationship
	contentRels []Relationship
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
		config:   cfg,
		elements: make([]Element, 0),
		images:   make(map[string]*ImageData),
		rels:     make([]Relationship, 0),
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
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`)

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

	buf.WriteString(`
        <w:sectPr>
            <w:pgSz w:w="11906" w:h="16838"/>
            <w:pgMar w:top="1440" w:right="1800" w:bottom="1440" w:left="1800" w:header="851" w:footer="992" w:gutter="0"/>
        </w:sectPr>
    </w:body>
</w:document>`)

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
