package converter

import (
"context"
"encoding/base64"
_ "embed"
"fmt"
"strings"
"time"

"github.com/chromedp/chromedp"
)

//go:embed highlight.min.js
var highlightJS string

//go:embed atom-one-dark.min.css
var atomOneDarkCSS string

// RenderCodeBlock 使用chromedp渲染带语法高亮的代码块为图片
func RenderCodeBlock(code string, language string) ([]byte, error) {
	execPath, err := FindChromePath()
	if err != nil {
		return nil, err
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>%s</style>
    <script>%s</script>
    <style>
        body {
            margin: 0;
            padding: 0;
            background: transparent;
        }
        #code-container {
            display: inline-block;
            padding: 24px;
            background: #282c34;
            border-radius: 8px;
            box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
            min-width: 400px;
        }
        pre {
            margin: 0;
            padding: 0;
        }
        code {
            font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
            font-size: 14px;
            line-height: 1.5;
            display: block;
            white-space: pre;
        }
    </style>
</head>
<body>
    <div id="code-container">
        <pre><code class="language-%s">%s</code></pre>
    </div>
    <script>
        try {
            hljs.highlightAll();
        } catch (e) {
            document.body.innerHTML += e.message;
        }
    </script>
</body>
</html>`, atomOneDarkCSS, highlightJS, language, escapeHTML(code))

	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(htmlContent))

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
chromedp.ExecPath(execPath),
chromedp.NoFirstRun,
chromedp.NoDefaultBrowserCheck,
chromedp.Headless,
chromedp.DisableGPU,
chromedp.WindowSize(2000, 2000),
chromedp.Flag("no-sandbox", true),
chromedp.Flag("disable-web-security", true),
)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err = chromedp.Run(ctx,
chromedp.Navigate(dataURL),
chromedp.Sleep(2*time.Second),
chromedp.Screenshot(`#code-container`, &buf, chromedp.NodeVisible),
	)

	if err != nil || len(buf) == 0 {
		return nil, fmt.Errorf("chromedp渲染代码块失败: %v", err)
	}

	return buf, nil
}

// escapeHTML 转义HTML特殊字符
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
