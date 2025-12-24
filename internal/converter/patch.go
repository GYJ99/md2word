package converter

// PatchMermaidError adds better error handling for chromedp Mermaid rendering
func PatchMermaidError() string {
	return `
		p.SpacingB = 200
		p.SpacingA = 200
		p.Shading = "FFF3CD"
		p.Border = true

		notice := p.AddRun("[流程图渲染失败]\n")
		notice.Bold = true
		notice.Color = "856404"
		
		errMsg := p.AddRun(fmt.Sprintf("错误: %v\n\n", err))
		errMsg.FontSize = 9
		errMsg.Color = "856404"
		
		helpMsg := p.AddRun("请确保:\n1. 已安装Chrome或Chromium浏览器\n2. Mermaid语法正确\n\n原始代码:\n")
		helpMsg.FontSize = 9

		codeRun := p.AddRun(mermaidCode)
		codeRun.FontName = "Consolas"
		codeRun.FontSize = 9
	`
}
