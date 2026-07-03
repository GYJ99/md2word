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

	// 编号状态跟踪
	numberingState *docx.NumberingState
}

// Element 文档元素接口
type Element interface {
	ToXML() string
}

// NewConverter 创建新的转换器
func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		config:         cfg,
		parser:         parser.NewMarkdownParser(),
		elements:       make([]Element, 0),
		numberingState: docx.NewNumberingState(),
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

	// 提取标题文本 - 修复文本提取逻辑
	var headingText strings.Builder
	c.extractTextFromNode(node, &headingText)

	// 解析标题中的编号
	originalText := headingText.String()
	parsedNum := docx.ParseHeadingNumber(originalText)

	if parsedNum != nil {
		// 找到了编号，设置Word自动编号
		// 获取或创建编号实例
		numState := c.doc.GetNumberingState()
		instance := numState.GetOrCreateNumberingInstance(parsedNum)
		
		if instance != nil {
			// 计算Word的编号级别 (0-based)
			ilvl := parsedNum.Level - 1
			if ilvl < 0 {
				ilvl = 0
			}
			if ilvl > 8 {
				ilvl = 8
			}

			// 设置段落的编号属性
			p.NumberingXML = docx.GetNumberingXMLForParagraph(ilvl, instance.NumId)
		}

		// 使用移除编号后的标题文本
		cleanRun := p.AddRun(parsedNum.Text)
		cleanRun.Bold = c.config.GetHeadingStyle(level).Bold

		// 更新编号状态
		numState.UpdateNumberingState(parsedNum)
	} else {
		// 没有编号，正常处理内联节点
		c.processInlineNodes(node, p)
	}

	c.doc.AddParagraph(p)
	return nil
}

// extractTextFromNode 递归提取节点中的所有文本
func (c *Converter) extractTextFromNode(node ast.Node, builder *strings.Builder) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			builder.WriteString(string(n.Segment.Value(c.source)))
		case *ast.Emphasis:
			// 处理加粗/斜体标记，继续提取内部文本
			c.extractTextFromNode(n, builder)
		default:
			// 递归处理其他节点
			c.extractTextFromNode(n, builder)
		}
	}
}

// processParagraph 处理段落
func (c *Converter) processParagraph(node *ast.Paragraph) error {
	p := docx.NewParagraph("")

	// 应用正文配置
	p.SpacingA = c.config.Styles.Body.SpaceBefore
	p.SpacingB = c.config.Styles.Body.SpaceAfter
	p.LineHeight = c.config.Styles.Body.LineHeight
	p.FirstLineIndent = c.config.Styles.Body.FirstLineIndent

	// 首先检查段落是否包含数学公式
	paragraphText := c.extractParagraphText(node)
	if strings.Contains(paragraphText, "$") {
		// 包含公式，使用特殊处理
		c.processParagraphWithFormulas(node, p)
	} else {
		// 不包含公式，正常处理内联节点
		c.processInlineNodes(node, p)
	}

	// 如果段落有内容（子元素），则添加到文档
	if len(p.Children) > 0 {
		c.doc.AddParagraph(p)
	}

	return nil
}

// extractParagraphText 提取段落的完整文本内容
func (c *Converter) extractParagraphText(node ast.Node) string {
	var builder strings.Builder
	c.extractTextFromNode(node, &builder)
	return builder.String()
}

// processParagraphWithFormulas 处理包含公式的段落
func (c *Converter) processParagraphWithFormulas(node ast.Node, p *docx.Paragraph) {
	// 获取段落的完整文本
	fullText := c.extractParagraphText(node)
	
	// 解析公式位置
	formulas := c.parseInlineFormulas(fullText)
	
	if len(formulas) == 0 {
		// 没有找到公式，正常处理
		c.processInlineNodes(node, p)
		return
	}
	
	// 按位置处理文本和公式
	lastEnd := 0
	for _, formula := range formulas {
		// 添加公式前的文本
		if formula.Start > lastEnd {
			beforeText := fullText[lastEnd:formula.Start]
			if beforeText != "" {
				p.AddRun(beforeText)
			}
		}
		
		// 处理公式
		imgData, err := RenderMathJax(formula.Formula, false)
		if err == nil && len(imgData) > 0 {
			width, height := c.getImageDimensions(imgData)
			if width > 0 && height > 0 {
				// 计算适合的显示尺寸
				displayW, displayH := c.calculateFormulaSize(width, height, true) // true表示行内公式
				
				rID := c.doc.AddImage(imgData, "image/png", width, height)
				p.AddImageRun(rID, int64(displayW)*9525, int64(displayH)*9525)
			} else {
				// 尺寸异常，作为文本处理
				p.AddRun("$" + formula.Formula + "$")
			}
		} else {
			// 渲染失败，作为文本处理
			p.AddRun("$" + formula.Formula + "$")
		}
		
		lastEnd = formula.End
	}
	
	// 添加剩余的文本
	if lastEnd < len(fullText) {
		remainingText := fullText[lastEnd:]
		if remainingText != "" {
			p.AddRun(remainingText)
		}
	}
}

// FormulaInfo 公式信息
type FormulaInfo struct {
	Start   int
	End     int
	Formula string
}

// parseInlineFormulas 解析行内公式
func (c *Converter) parseInlineFormulas(text string) []FormulaInfo {
	var formulas []FormulaInfo
	start := 0
	
	for {
		idx := strings.Index(text[start:], "$")
		if idx == -1 {
			break
		}
		
		absIdx := start + idx
		
		// 检查是否是转义的 $
		if absIdx > 0 && text[absIdx-1] == '\\' {
			start = absIdx + 1
			continue
		}
		
		// 检查是否是块级公式 $$
		if absIdx+1 < len(text) && text[absIdx+1] == '$' {
			start = absIdx + 2
			continue
		}
		
		// 查找结束的 $
		endIdx := strings.Index(text[absIdx+1:], "$")
		if endIdx == -1 {
			break
		}
		
		absEndIdx := absIdx + 1 + endIdx
		formula := strings.TrimSpace(text[absIdx+1 : absEndIdx])
		
		if formula != "" {
			formulas = append(formulas, FormulaInfo{
				Start:   absIdx,
				End:     absEndIdx + 1,
				Formula: formula,
			})
		}
		
		start = absEndIdx + 1
	}
	
	return formulas
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
			// 对于普通文本，直接添加（公式已在段落级别处理）
			run := p.AddRun(text)
			run.Bold = bold
			run.Italic = italic
			run.Strike = strike
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
		if para, ok := p.(*docx.Paragraph); ok {
			rID := c.doc.AddHyperlink(url)
			link := para.AddHyperlink(rID)
			for child := node.FirstChild(); child != nil; child = child.NextSibling() {
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

	// 使用智能尺寸计算
	displayW, displayH := c.calculateOptimalImageSize(width, height, false)

	rID := c.doc.AddImage(data, contentType, width, height)
	// Word使用EMU单位: 1 pixel 约等于 9525 EMUs
	p.AddImageRun(rID, int64(displayW)*9525, int64(displayH)*9525)
}

// calculateOptimalImageSize 计算图片的最佳显示尺寸
// 目标：适配Word页面宽度，保持高清，控制合理尺寸
func (c *Converter) calculateOptimalImageSize(originalWidth, originalHeight int, isFlowchart bool) (displayWidth, displayHeight int) {
	// Word A4页面可用宽度约为605像素（96DPI下）
	// 但我们使用稍微保守的值以确保在不同页面设置下都能正常显示
	maxPageWidth := 580 // 像素
	
	// 对于流程图，允许更大的宽度以保持清晰度
	if isFlowchart {
		maxPageWidth = 650
	}
	
	// 设置最小和最大尺寸限制
	minWidth := 100
	maxHeight := 800 // 避免图片过高
	
	displayWidth = originalWidth
	displayHeight = originalHeight
	
	// 对于高分辨率图片（如2倍缩放的流程图），先缩小到合理尺寸
	if isFlowchart && originalWidth > 1000 {
		// 流程图通常是高分辨率渲染的，需要缩小到合适的显示尺寸
		scale := 0.6 // 缩小到60%
		displayWidth = int(float64(originalWidth) * scale)
		displayHeight = int(float64(originalHeight) * scale)
	}
	
	// 如果图片太小，适当放大（但不超过原始尺寸的2倍）
	if displayWidth < minWidth && displayWidth > 0 {
		scale := float64(minWidth) / float64(displayWidth)
		if scale <= 2.0 { // 最多放大2倍
			displayWidth = int(float64(displayWidth) * scale)
			displayHeight = int(float64(displayHeight) * scale)
		}
	}
	
	// 如果宽度超过页面宽度，按比例缩小
	if displayWidth > maxPageWidth {
		ratio := float64(maxPageWidth) / float64(displayWidth)
		displayWidth = maxPageWidth
		displayHeight = int(float64(displayHeight) * ratio)
	}
	
	// 如果高度过高，按比例缩小
	if displayHeight > maxHeight {
		ratio := float64(maxHeight) / float64(displayHeight)
		displayHeight = maxHeight
		displayWidth = int(float64(displayWidth) * ratio)
	}
	
	// 确保尺寸不为0
	if displayWidth <= 0 {
		displayWidth = minWidth
	}
	if displayHeight <= 0 {
		displayHeight = int(float64(displayWidth) * 0.75) // 默认4:3比例
	}
	
	return displayWidth, displayHeight
}

// calculateFormulaSize 计算数学公式的最佳显示尺寸
func (c *Converter) calculateFormulaSize(originalWidth, originalHeight int, isInline bool) (displayWidth, displayHeight int) {
	displayWidth = originalWidth
	displayHeight = originalHeight
	
	if isInline {
		// 行内公式的限制
		maxInlineWidth := 350   // 行内公式最大宽度（像素）- 适配正常文本行
		maxInlineHeight := 18   // 行内公式最大高度（像素）- 匹配文本行高
		
		// 首先检查高度，确保不超过行高
		if displayHeight > maxInlineHeight {
			ratio := float64(maxInlineHeight) / float64(displayHeight)
			displayWidth = int(float64(displayWidth) * ratio)
			displayHeight = maxInlineHeight
		}
		
		// 然后检查宽度，确保不超过合理范围
		if displayWidth > maxInlineWidth {
			ratio := float64(maxInlineWidth) / float64(displayWidth)
			displayWidth = maxInlineWidth
			displayHeight = int(float64(displayHeight) * ratio)
			
			// 如果缩小后高度太小，适当调整
			if displayHeight < 12 {
				displayHeight = 12
				displayWidth = int(float64(originalWidth) * float64(displayHeight) / float64(originalHeight))
				if displayWidth > maxInlineWidth {
					displayWidth = maxInlineWidth
				}
			}
		}
	} else {
		// 块级公式的限制
		maxBlockWidth := 380    // 块级公式最大宽度（像素）- 适中尺寸，确保两侧有充足留白
		maxBlockHeight := 220   // 块级公式最大高度（像素）- 适中高度，避免过高
		
		// 对于原始尺寸较小的公式，不要过度放大
		if originalWidth < 200 && originalHeight < 100 {
			// 小公式保持相对原始尺寸，但限制最大值
			maxSmallWidth := int(float64(maxBlockWidth) * 0.7) // 小公式最多占70%宽度
			if displayWidth > maxSmallWidth {
				ratio := float64(maxSmallWidth) / float64(displayWidth)
				displayWidth = maxSmallWidth
				displayHeight = int(float64(displayHeight) * ratio)
			}
		} else {
			// 大公式按正常规则缩放
			if displayWidth > maxBlockWidth {
				ratio := float64(maxBlockWidth) / float64(displayWidth)
				displayWidth = maxBlockWidth
				displayHeight = int(float64(displayHeight) * ratio)
			}
		}
		
		// 检查高度
		if displayHeight > maxBlockHeight {
			ratio := float64(maxBlockHeight) / float64(displayHeight)
			displayHeight = maxBlockHeight
			displayWidth = int(float64(displayWidth) * ratio)
		}
		
		// 确保块级公式有合理的最小尺寸
		if displayWidth < 60 {
			displayWidth = 60
		}
		if displayHeight < 20 {
			displayHeight = 20
		}
	}
	
	// 确保尺寸不为0
	if displayWidth <= 0 {
		if isInline {
			displayWidth = 30
		} else {
			displayWidth = 50
		}
	}
	if displayHeight <= 0 {
		if isInline {
			displayHeight = 15
		} else {
			displayHeight = 20
		}
	}
	
	return displayWidth, displayHeight
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

	imgData, err := RenderMermaidWithContext(ctx, mermaidCode, c.config.Mermaid.Theme, c.config.Mermaid.Width, c.config.Mermaid.Height, c.config.Mermaid.Scale)
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
	
	// 使用智能尺寸计算，流程图标记为true
	displayW, displayH := c.calculateOptimalImageSize(width, height, true)

	rID := c.doc.AddImage(imgData, "image/png", width, height)
	p := docx.NewParagraph("")
	p.Align = "center"
	p.AddImageRun(rID, int64(displayW)*9525, int64(displayH)*9525)
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
	
	// 计算适合的显示尺寸
	displayW, displayH := c.calculateFormulaSize(width, height, false) // false表示块级公式
	
	rID := c.doc.AddImage(imgData, "image/png", width, height)
	p := docx.NewParagraph("")
	if display {
		p.Align = "center"
	}
	p.AddImageRun(rID, int64(displayW)*9525, int64(displayH)*9525)
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

	// 使用首行缩进(不是悬挂缩进):
	// - 内容顶格(左缩进=0)
	// - 序号向右移动360 twips
	// 结果:序号行在360位置(有缩进),内容换行后在0位置(顶格)
	p.Indent = 0
	p.FirstLineIndent = 360 // 序号缩进

	p.LineHeight = c.config.Styles.Body.LineHeight
	if isOrdered {
		p.AddRun(fmt.Sprintf("%d. ", index)).Bold = true
	} else {
		p.AddRun("• ").Bold = true
	}

	// 收集嵌套列表,稍后处理
	var nestedLists []*ast.List

	// 先处理内联内容
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if list, ok := child.(*ast.List); ok {
			// 保存嵌套列表,稍后处理
			nestedLists = append(nestedLists, list)
			continue
		}
		c.processInlineNodes(child, p)
	}

	// 先添加当前段落
	c.doc.AddParagraph(p)

	// 再处理嵌套列表(在当前段落之后)
	for _, list := range nestedLists {
		c.processList(list, level+1)
	}
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
