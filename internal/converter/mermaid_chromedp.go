package converter

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

//go:embed mermaid.min.js
var mermaidJS string

func RenderMermaidChromedp(code string, theme string) ([]byte, error) {
	execPath, err := FindChromePath()
	if err != nil {
		return nil, err
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.ExecPath(execPath), chromedp.Headless, chromedp.DisableGPU)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	return RenderMermaidWithContext(ctx, code, theme, 1200, 800, 2)
}

func RenderMermaidWithContext(ctx context.Context, code string, theme string, width, height, scale int) ([]byte, error) {
	if theme == "" {
		theme = "default"
	}
	if width <= 0 {
		width = 1200
	}
	if height <= 0 {
		height = 900
	}
	if scale <= 0 {
		scale = 2
	}

	homeDir, _ := os.UserHomeDir()
	tmpDir := filepath.Join(homeDir, ".md2word-mermaid-tmp")
	os.MkdirAll(tmpDir, 0755)

	jsPath := filepath.Join(tmpDir, "mermaid.min.js")
	os.WriteFile(jsPath, []byte(mermaidJS), 0644)

	htmlPath := filepath.Join(tmpDir, "render.html")
	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <script src="mermaid.min.js"></script>
    <style>
        body { 
            margin: 0; 
            padding: 20px; 
            background: white; 
            font-family: Arial, sans-serif;
        }
        #diagram { 
            display: inline-block;
            background: white;
        }
        .mermaid svg {
            max-width: none !important;
            height: auto !important;
        }
    </style>
</head>
<body>
    <div id="diagram" class="mermaid">%s</div>
    <script>
        mermaid.initialize({ 
            startOnLoad: true, 
            theme: '%s',
            securityLevel: 'loose',
            flowchart: {
                useMaxWidth: false,
                htmlLabels: true
            },
            sequence: {
                useMaxWidth: false
            },
            gantt: {
                useMaxWidth: false
            }
        });
        
        // 等待渲染完成
        setTimeout(() => {
            const diagram = document.getElementById('diagram');
            const svg = diagram.querySelector('svg');
            if (svg) {
                // 设置高质量渲染
                svg.style.background = 'white';
                svg.setAttribute('width', svg.getBBox().width * %d);
                svg.setAttribute('height', svg.getBBox().height * %d);
            }
        }, 2000);
    </script>
</body>
</html>`, code, theme, scale, scale)

	os.WriteFile(htmlPath, []byte(htmlContent), 0644)

	var buf []byte
	// 增加超时时间，确保高质量渲染完成
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	absHtmlPath, _ := filepath.Abs(htmlPath)
	
	// 改进的渲染流程，确保高质量输出
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate("file://"+absHtmlPath),
		chromedp.Sleep(3*time.Second), // 等待页面加载
		chromedp.WaitVisible(`#diagram svg`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // 等待渲染完成
		chromedp.Screenshot(`#diagram`, &buf, chromedp.NodeVisible, chromedp.ByID),
	)

	return buf, err
}
