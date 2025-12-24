package converter

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	goldmarkText "github.com/yuin/goldmark/text"

	"md2word/internal/config"
	"md2word/internal/docx"
	"md2word/internal/parser"
)

// Converter Markdown到DOCX转换器
type Converter struct {
	config    *config.Config
	doc       *docx.Document
	parser    *parser.MarkdownParser
	source    []byte
	basePath  string
	tableMode bool
	elements  []Element

	// Chromedp 资源
	chromeCtx    context.Context
	chromeCancel context.CancelFunc
}

// Element 文档元素接口
type Element interface {
	ToXML() string
}

// NewConverter 创建新的转换器
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		config:   cfg,
		parser:   parser.NewMarkdownParser(),
		elements: make([]Element, 0),
	}
}

// Convert 转换Markdown到DOCX
func (c *Converter) Convert(content []byte, outputPath string) error {
	c.source = content
	c.basePath = filepath.Dir(outputPath)
	c.doc = docx.NewDocument(c.config)

	// 在转换结束时关闭浏览器
	defer c.Close()

	// 解析Markdown
	root := c.parser.Parse(content)

	// 遍历AST
	if err := c.walkNode(root); err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}

	// 保存文档
	return c.doc.Save(outputPath)
}

// Close 关闭转换器并释放资源
func (c *Converter) Close() {
	if c.chromeCancel != nil {
		c.chromeCancel()
		c.chromeCtx = nil
		c.chromeCancel = nil
	}
}

// ensureChrome 确保 chromedp 上下文已初始化
func (c *Converter) ensureChrome() (context.Context, error) {
	if c.chromeCtx != nil {
		return c.chromeCtx, nil
	}

	execPath, err := FindChromePath()
	if err != nil {
		return nil, err
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(execPath),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, cancel2 := chromedp.NewContext(allocCtx)
	c.chromeCtx = ctx
	c.chromeCancel = func() {
		cancel2()
		cancel()
	}

	// 启动浏览器
	chromedp.Run(c.chromeCtx, chromedp.Navigate("about:blank"))

	return c.chromeCtx, nil
}

// walkNode 遍历AST节点
func (c *Converter) walkNode(n ast.Node) error {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if err := c.processNode(child); err != nil {
			return err
		}
	}
	return nil
}

// processNode 处理单个节点
func (c *Converter) processNode(n ast.Node) error {
	switch node := n.(type) {
	case *ast.Heading:
		return c.processHeading(node)
	case *ast.Paragraph:
		return c.processParagraph(node)
	case *ast.TextBlock:
		return c.processTextBlock(node)
	case *ast.FencedCodeBlock:
		return c.processFencedCodeBlock(node)
	case *ast.CodeBlock:
		return c.processCodeBlock(node)
	case *ast.List:
		return c.processList(node, 0)
	case *ast.Blockquote:
		return c.processBlockquote(node)
	case *ast.ThematicBreak:
		return c.processThematicBreak()
	case *east.Table:
		return c.processTable(node)
	default:
		// 处理其他节点类型
		return c.walkNode(n)
	}
}

// processHeading 处理标题
func (c *Converter) processHeading(node *ast.Heading) error {
	level := node.Level
	if level > 9 {
		level = 9
	}

	styleID := fmt.Sprintf("Heading%d", level)
	p := docx.NewParagraph(styleID)
	p.LineHeight = c.config.Styles.Body.LineHeight

	c.processInlineNodes(node, p)
	c.doc.AddParagraph(p)

	return nil
}

// processParagraph 处理段落
func (c *Converter) processParagraph(node *ast.Paragraph) error {
	p := docx.NewParagraph("")

	// 应用正文配置
	p.SpacingA = c.config.Styles.Body.SpaceBefore
	p.SpacingB = c.config.Styles.Body.SpaceAfter
	p.LineHeight = c.config.Styles.Body.LineHeight
	p.FirstLineIndent = c.config.Styles.Body.FirstLineIndent

	// 处理内联节点
	c.processInlineNodes(node, p)

	// 如果段落有内容（子元素），则添加到文档
	if len(p.Children) > 0 {
		c.doc.AddParagraph(p)
	}

	return nil
}

// processTextBlock 处理文本块
func (c *Converter) processTextBlock(node *ast.TextBlock) error {
	p := docx.NewParagraph("")
	c.processInlineNodes(node, p)
	c.doc.AddParagraph(p)
	return nil
}

// processInlineNodes 处理内联节点
func (c *Converter) processInlineNodes(parent ast.Node, p docx.RunContainer) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		c.processInlineNode(child, p, false, false, false, false)
	}
}

// processInlineNode 处理单个内联节点
func (c *Converter) processInlineNode(n ast.Node, p docx.RunContainer, bold, italic, code, strike bool) {
	switch node := n.(type) {
	case *ast.Text:
		text := string(node.Segment.Value(c.source))
		if code {
			run := p.AddRun(text)
			run.IsCode = true
			run.FontName = c.config.Styles.Code.Font
			if run.FontName == "" {
				run.FontName = "Consolas"
			}
			run.FontSize = c.config.Styles.Code.Size
			if run.FontSize == 0 {
				run.FontSize = 10.5
			}
			if c.config.Styles.Code.Color != "" {
				run.Color = strings.TrimPrefix(c.config.Styles.Code.Color, "#")
			}
		} else {
			c.processTextWithInlineFormulas(text, p, bold, italic, strike)
		}
	case *ast.Emphasis:
		level := node.Level
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			c.processInlineNode(child, p, bold || level == 2, italic || level == 1, code, strike)
		}
	case *ast.CodeSpan:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			c.processInlineNode(child, p, bold, italic, true, strike)
		}
	case *ast.Link:
		url := string(node.Destination)
		// 如果 p 是 Paragraph，则可以添加超链接
		if para, ok := p.(*docx.Paragraph); ok {
			rID := c.doc.AddHyperlink(url)
			link := para.AddHyperlink(rID)
			// 超链接内的文本通常由 Word 自动样式化，如果需要强制样式（如蓝色下划线），可以在这里传递 bold/italic 等
			// 但通常让 Word 处理即可，或者我们手动应用 "Hyperlink" 样式
			// 这里我们继续递归处理子节点，将它们添加到 link 中
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				// 将子节点添加进 link (link 也是 RunContainer)
				// 注意：Word 超链接内默认不自动变蓝，需要我们手动设置 Run 属性
				// 我们创建一个带默认超链接样式的 helper? 或者直接在 processTextWithInlineFormulas 里判断?
				// 既然我们控制 Run，我们就手动设置颜色和下划线
				// 我们需要一个 flag 告诉 processInlineNode "这是在链接里" 吗？
				// 上下文 p 变成了 link，我们可以针对 link 容器的所有新增 Run 设置样式
				// 但 RunContainer 接口没有 "SetStyle" 方法。
				// 简单办法：在添加完 Run 后，手动修改 Run 属性?
				// 这里的 processInlineNode 会调用 AddRun。
				// 我们可以在调用 processInlineNode 之前，或者之后?
				// 最好是 processInlineNode 里面的 AddRun 返回 run，然后我们无法拦截。
				// 解决方案：Hyperlink.AddRun 实现中可以默认给 Run 设置样式！
				// 或者：我们在 Hyperlink 结构体中添加 AddRun 实现时就加上样式。
				// 让我们修改 paragraph.go 中的 Hyperlink.AddRun 吗？
				// 或者在这里递归调用时，我们可以不用 processInlineNode，而是自己写循环？
				// 不，为了支持加粗斜体等嵌套，必须复用 processInlineNode。
				// 那么我们如何让 link 中的 run 变蓝？
				// 我们可以修改 Hyperlink.AddRun/AddFormattedRun 来默认应用样式。
				// 但这需要修改 paragraph.go。
				// 也可以：遍历 link.Runs 并修改。但 processInlineNode 不返回 runs。
				// 我们可以让 link.Runs 在处理完子节点后遍历修改。
				c.processInlineNode(child, link, bold, italic, code, strike)
			}
			// 处理完所有子节点后，统一给 link 的 Runs 加上超链接样式
			for _, run := range link.Runs {
				if run.Color == "" {
					run.Color = "0563C1" // Word 默认链接蓝
				}
				run.Underline = true
			}
		} else {
			// 嵌套链接（不被支持），作为普通文本处理
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
				c.processInlineNode(child, p, bold, italic, code, strike)
			}
		}

	case *ast.Image:
		c.processImage(node, p)
	case *east.Strikethrough:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			c.processInlineNode(child, p, bold, italic, code, true)
		}
	}
}

// processTextWithInlineFormulas 处理包含行内公式的文本
func (c *Converter) processTextWithInlineFormulas(text string, p docx.RunContainer, bold, italic, strike bool) {
	start := 0
	for {
		idx := strings.Index(text[start:], "$")
		if idx == -1 {
			break
		}

		absIdx := start + idx
		if absIdx > 0 && text[absIdx-1] == '\\' {
			run := p.AddRun(text[start:absIdx-1] + "$")
			run.Bold = bold
			run.Italic = italic
			start = absIdx + 1
			continue
		}

		endIdx := strings.Index(text[absIdx+1:], "$")
		if endIdx == -1 {
			break
		}

		absEndIdx := absIdx + 1 + endIdx
		if absIdx > start {
			run := p.AddRun(text[start:absIdx])
			run.Bold = bold
			run.Italic = italic
		}

		formula := text[absIdx+1 : absEndIdx]
		imgData, err := RenderMathJax(formula, false)
		if err == nil {
			width, height := c.getImageDimensions(imgData)
			rID := c.doc.AddImage(imgData, "image/png", width, height)
			p.AddImageRun(rID, int64(width)*9525, int64(height)*9525)
		} else {
			run := p.AddRun("$" + formula + "$")
			run.Bold = bold
			run.Italic = italic
		}

		start = absEndIdx + 1
	}

	if start < len(text) {
		run := p.AddRun(text[start:])
		run.Bold = bold
		run.Italic = italic
	}
}

// processImage 处理图片
func (c *Converter) processImage(node *ast.Image, p docx.RunContainer) {
	src := string(node.Destination)
	var data []byte
	var contentType string
	var err error

	if strings.HasPrefix(src, "http") {
		data, contentType, err = c.downloadImage(src)
	} else if strings.HasPrefix(src, "data:image") {
		data, contentType, err = c.parseBase64Image(src)
	} else {
		data, contentType, err = c.loadLocalImage(src)
	}

	if err != nil {
		fmt.Printf("图片加载失败: %s, %v\n", src, err)
		return
	}

	// 如果没有检测到 contentType，尝试从数据中检测
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = http.DetectContentType(data)
	}

	width, height := c.getImageDimensions(data)

	// 计算显示宽度（Word 使用 EMU，1英寸=914400 EMU，A4页宽约6.5英寸=5943600 EMU）
	// 这里我们使用配置文件中的 MaxWidth (默认为 600px)
	displayW := width
	displayH := height

	maxWidth := c.config.Images.MaxWidth
	if displayW > maxWidth {
		ratio := float64(maxWidth) / float64(displayW)
		displayW = maxWidth
		displayH = int(float64(displayH) * ratio)
	}

	rID := c.doc.AddImage(data, contentType, width, height)
	// Word使用EMU单位: 1 pixel 约等于 9525 EMUs
	p.AddImageRun(rID, int64(displayW)*9525, int64(displayH)*9525)
}

func (c *Converter) downloadImage(url string) ([]byte, string, error) {
	client := &http.Client{
		Timeout: time.Duration(c.config.Images.DownloadTimeout) * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	return data, resp.Header.Get("Content-Type"), err
}

func (c *Converter) parseBase64Image(src string) ([]byte, string, error) {
	parts := strings.Split(src, ",")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid base64 image")
	}

	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", err
	}

	// 动态检测内容类型
	contentType := http.DetectContentType(data)
	return data, contentType, nil
}

func (c *Converter) loadLocalImage(path string) ([]byte, string, error) {
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.basePath, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	// 动态检测内容类型
	contentType := http.DetectContentType(data)
	return data, contentType, nil
}

func (c *Converter) getImageDimensions(data []byte) (int, int) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 300, 200
	}
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

// processFencedCodeBlock 处理围栏代码块
func (c *Converter) processFencedCodeBlock(node *ast.FencedCodeBlock) error {
	lang := string(node.Language(c.source))

	if strings.ToLower(lang) == "mermaid" && c.config.Mermaid.Enabled {
		return c.processMermaid(node)
	}

	if strings.ToLower(lang) == "math" || strings.ToLower(lang) == "latex" {
		return c.processMathBlock(node)
	}

	table := docx.NewTable()
	table.HasBorders = true
	row := table.AddRow(false)
	cell := row.AddCell()
	cell.Shading = "F6F8FA"

	var code strings.Builder
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		code.WriteString(string(line.Value(c.source)))
	}

	// 执行高亮渲染到单元格中
	fontName := c.config.Styles.CodeBlock.Font
	fontSize := c.config.Styles.CodeBlock.Size
	lineSpacing := c.config.Styles.CodeBlock.LineSpacing
	lineHeight := c.config.Styles.CodeBlock.LineHeight
	if fontName == "" {
		fontName = "Consolas"
	}
	if fontSize == 0 {
		fontSize = 9.5
	}

	if err := HighlightCodeNative(cell, code.String(), lang, fontName, fontSize, lineSpacing, lineHeight); err != nil {
		// 回退处理
		p := docx.NewParagraph("")
		p.SpacingA = lineSpacing / 2
		p.SpacingB = lineSpacing / 2
		p.LineHeight = lineHeight
		run := p.AddRun(code.String())
		run.FontName = fontName
		run.FontSize = fontSize
		cell.AddParagraph(p)
	}
	c.doc.AddParagraph(docx.NewTableElement(table))
	c.doc.AddParagraph(docx.NewParagraph(""))
	return nil
}

// processCodeBlock 处理缩进代码块 - 重新解析为Markdown
func (c *Converter) processCodeBlock(node *ast.CodeBlock) error {
	var lines []string
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		lineStr := string(line.Value(c.source))
		// 去除4空格缩进（如果存在）
		if len(lineStr) >= 4 && lineStr[:4] == "    " {
			lineStr = lineStr[4:]
		}
		lines = append(lines, lineStr)
	}

	// 合并为新的Markdown内容，重新解析
	content := strings.Join(lines, "")

	// 使用Goldmark重新解析这段内容
	mdParser := c.parser.GetParser()
	reader := goldmarkText.NewReader([]byte(content))
	subDoc := mdParser.Parse(reader)

	// 保存原始source，临时替换为新内容
	originalSource := c.source
	c.source = []byte(content)

	// 递归处理子AST
	for child := subDoc.FirstChild(); child != nil; child = child.NextSibling() {
		if err := c.processNode(child); err != nil {
			c.source = originalSource
			return err
		}
	}

	// 恢复原始source
	c.source = originalSource
	return nil
}

// processMermaid 处理Mermaid流程图
func (c *Converter) processMermaid(node *ast.FencedCodeBlock) error {
	fmt.Println("正在处理 Mermaid 流程图...")
	var lines []string
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		lines = append(lines, string(line.Value(c.source)))
	}
	mermaidCode := strings.Join(lines, "")

	ctx, err := c.ensureChrome()
	if err != nil {
		return fmt.Errorf("启动浏览器失败: %w", err)
	}

	imgData, err := RenderMermaidWithContext(ctx, mermaidCode, c.config.Mermaid.Theme)
	if err != nil {
		fmt.Printf("Mermaid 渲染错误: %v\n", err)
		p := docx.NewParagraph("")
		p.Shading = "FFF3CD"
		p.Border = true
		p.AddRun("[流程图渲染失败]\n").Bold = true
		p.AddRun("原始代码:\n" + mermaidCode).FontName = "Consolas"
		c.doc.AddParagraph(p)
		return nil
	}

	width, height := c.getImageDimensions(imgData)
	if width > c.config.Images.MaxWidth {
		ratio := float64(c.config.Images.MaxWidth) / float64(width)
		width = c.config.Images.MaxWidth
		height = int(float64(height) * ratio)
	}

	rID := c.doc.AddImage(imgData, "image/png", width, height)
	p := docx.NewParagraph("")
	p.Align = "center"
	p.AddImageRun(rID, int64(width)*9525, int64(height)*9525)
	c.doc.AddParagraph(p)
	return nil
}

func (c *Converter) processMathBlock(node *ast.FencedCodeBlock) error {
	var lines []string
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		lines = append(lines, string(line.Value(c.source)))
	}
	latex := strings.Join(lines, "")
	return c.renderMathAsImage(latex, true)
}

func (c *Converter) renderMathAsImage(latex string, display bool) error {
	imgData, err := RenderMathJax(latex, display)
	if err != nil {
		return err
	}
	width, height := c.getImageDimensions(imgData)
	rID := c.doc.AddImage(imgData, "image/png", width, height)
	p := docx.NewParagraph("")
	if display {
		p.Align = "center"
	}
	p.AddImageRun(rID, int64(width)*9525, int64(height)*9525)
	c.doc.AddParagraph(p)
	return nil
}

// processList 处理列表
func (c *Converter) processList(node *ast.List, level int) error {
	i := 1
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if item, ok := child.(*ast.ListItem); ok {
			c.processListItem(item, node.IsOrdered(), i, level)
			i++
		}
	}
	return nil
}

func (c *Converter) processListItem(node *ast.ListItem, isOrdered bool, index int, level int) {
	p := docx.NewParagraph("")
	p.Indent = (level + 1) * 360
	p.LineHeight = c.config.Styles.Body.LineHeight
	if isOrdered {
		p.AddRun(fmt.Sprintf("%d. ", index)).Bold = true
	} else {
		p.AddRun("• ").Bold = true
	}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok {
			c.processList(list, level+1)
			continue
		}
		c.processInlineNodes(child, p)
	}
	c.doc.AddParagraph(p)
}

func (c *Converter) processBlockquote(node *ast.Blockquote) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		tempP := docx.NewParagraph("")
		c.processInlineNodes(child, tempP)
		tempP.Shading = "F0F0F0"
		tempP.Border = true
		tempP.Indent = 360
		tempP.LineHeight = c.config.Styles.Body.LineHeight
		c.doc.AddParagraph(tempP)
	}
	return nil
}

func (c *Converter) processThematicBreak() error {
	p := docx.NewParagraph("")
	p.HorizontalRule = true
	p.SpacingA = 120
	p.SpacingB = 120
	c.doc.AddParagraph(p)
	return nil
}

func (c *Converter) processTable(node *east.Table) error {
	// 简单的表格占位符，可以稍后细化
	table := docx.NewTable()
	table.HasBorders = true
	for row := node.FirstChild(); row != nil; row = row.NextSibling() {
		r := table.AddRow(false) // 简化处理
		for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
			c_cell := r.AddCell()
			p := docx.NewParagraph("")
			c.processInlineNodes(cell, p)
			c_cell.AddParagraph(p)
		}
	}
	c.doc.AddParagraph(docx.NewTableElement(table))
	return nil
}
