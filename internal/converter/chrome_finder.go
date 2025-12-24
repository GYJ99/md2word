package converter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// FindChromePath 查找Chrome或Chromium可执行文件路径
func FindChromePath() (string, error) {
	var paths []string

	switch runtime.GOOS {
	case "darwin": // macOS
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/usr/local/bin/chromium",
			"/opt/homebrew/bin/chromium",
			"/opt/homebrew/bin/google-chrome",
		}
	case "linux":
		paths = []string{
			"/usr/bin/google-chrome",
			"/usr/bin/chromium",
			"/usr/bin/chromium-browser",
			"/snap/bin/chromium",
			"/usr/bin/microsoft-edge",
		}
	case "windows":
		paths = []string{
			"C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
			"C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
		}
	}

	// 1. 检查已知路径
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			// 如果是软链接，尝试解析它（特别是对于 homebrew binstubs）
			if realPath, err := filepath.EvalSymlinks(path); err == nil {
				// 检查解析后的路径是否以 .sh 结尾（homebrew wrapper）
				if filepath.Ext(realPath) == ".sh" {
					// 如果是 wrapper，可能需要查找它指向的实际路径，或者直接使用它（依赖系统环境）
					// 这里我们继续尝试其他路径，除非没有更好的选择
				} else {
					return realPath, nil
				}
			}
			return path, nil
		}
	}

	// 2. 尝试使用 exec.LookPath 查找环境变量中的路径
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "msedge", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("未找到Chrome或Chromium浏览器，请确保已安装 Google Chrome 或 Microsoft Edge")
}
