package converter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RenderMermaid 使用mermaid-cli渲染Mermaid图
func RenderMermaid(code string, cliCmd string, theme string) ([]byte, error) {
	if cliCmd == "" {
		cliCmd = "mmdc"
	}
	if theme == "" {
		theme = "default"
	}

	// 检查mmdc是否可用
	if _, err := exec.LookPath(cliCmd); err != nil {
		return nil, fmt.Errorf("mermaid-cli not found: %w", err)
	}

	// 创建临时文件
	tmpDir, err := os.MkdirTemp("", "mermaid")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.mmd")
	outputFile := filepath.Join(tmpDir, "output.png")

	// 写入Mermaid代码
	if err := os.WriteFile(inputFile, []byte(code), 0644); err != nil {
		return nil, err
	}

	// 执行mmdc命令
	cmd := exec.Command(cliCmd,
		"-i", inputFile,
		"-o", outputFile,
		"-t", theme,
		"-b", "transparent",
		"-s", "2", // 缩放因子
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("mermaid render failed: %s, %w", string(output), err)
	}

	// 读取输出图片
	return os.ReadFile(outputFile)
}
