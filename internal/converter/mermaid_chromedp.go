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
	return RenderMermaidWithContext(ctx, code, theme)
}

func RenderMermaidWithContext(ctx context.Context, code string, theme string) ([]byte, error) {
	if theme == "" {
		theme = "default"
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
    <script src="mermaid.min.js"></script>
</head>
<body style="margin:0; background:white;">
    <div id="diagram" class="mermaid" style="display:inline-block; padding:20px;">%s</div>
    <script>
        (async () => {
            try {
                console.log("Start init");
                await mermaid.initialize({ startOnLoad: false, theme: '%s' });
                console.log("Start run");
                await mermaid.run({ nodes: [document.getElementById('diagram')] });
                console.log("Done");
                document.body.classList.add('ready');
            } catch (e) { console.error(e); }
        })();
    </script>
</body>
</html>`, code, theme)

	os.WriteFile(htmlPath, []byte(htmlContent), 0644)

	var buf []byte
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	absHtmlPath, _ := filepath.Abs(htmlPath)
	err := chromedp.Run(timeoutCtx,
chromedp.Navigate("file://"+absHtmlPath),
chromedp.WaitVisible(`body.ready`, chromedp.ByQuery),
chromedp.Screenshot(`#diagram`, &buf, chromedp.NodeVisible),
	)

	return buf, err
}
