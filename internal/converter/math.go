package converter

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RenderMathJax 使用MathJax渲染LaTeX公式为图片
func RenderMathJax(latex string, display bool) ([]byte, error) {
	// 首先尝试使用本地的mathjax-node-cli
	if data, err := renderMathJaxLocal(latex, display); err == nil {
		return data, nil
	}

	// 备用方案：使用在线服务
	return renderMathJaxOnline(latex, display)
}

// renderMathJaxLocal 使用本地mathjax-node渲染
func renderMathJaxLocal(latex string, display bool) ([]byte, error) {
	// 检查tex2svg是否可用
	cmdName := "tex2svg"
	if _, err := exec.LookPath(cmdName); err != nil {
		// 尝试npx
		cmdName = "npx"
	}

	tmpDir, err := os.MkdirTemp("", "mathjax")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	outputFile := filepath.Join(tmpDir, "output.svg")

	var cmd *exec.Cmd
	if cmdName == "npx" {
		args := []string{"mathjax-node-cli", "--output", outputFile}
		if display {
			args = append(args, "--display")
		}
		args = append(args, latex)
		cmd = exec.Command(cmdName, args...)
	} else {
		args := []string{}
		if display {
			args = append(args, "--display")
		}
		args = append(args, latex)
		cmd = exec.Command(cmdName, args...)
		cmd.Stdout = nil
	}

	// 获取SVG输出
	var svgBuf bytes.Buffer
	cmd = exec.Command("tex2svg", latex)
	if display {
		cmd = exec.Command("tex2svg", "--display", latex)
	}
	cmd.Stdout = &svgBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tex2svg failed: %w", err)
	}

	// 将SVG转换为PNG
	return convertSVGtoPNG(svgBuf.Bytes())
}

// renderMathJaxOnline 使用在线服务渲染
func renderMathJaxOnline(latex string, display bool) ([]byte, error) {
	// 使用 latex.codecogs.com 服务
	// URL格式: https://latex.codecogs.com/png.latex?{latex}

	encodedLatex := url.QueryEscape(latex)
	apiURL := fmt.Sprintf("https://latex.codecogs.com/png.latex?\\dpi{150}%s", encodedLatex)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("math render request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("math render failed with status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// convertSVGtoPNG 将SVG转换为PNG
func convertSVGtoPNG(svg []byte) ([]byte, error) {
	// 尝试使用rsvg-convert
	if _, err := exec.LookPath("rsvg-convert"); err == nil {
		cmd := exec.Command("rsvg-convert", "-f", "png", "-d", "150", "-p", "150")
		cmd.Stdin = bytes.NewReader(svg)
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err == nil {
			return out.Bytes(), nil
		}
	}

	// 尝试使用inkscape
	if _, err := exec.LookPath("inkscape"); err == nil {
		tmpDir, err := os.MkdirTemp("", "svg2png")
		if err != nil {
			return nil, err
		}
		defer os.RemoveAll(tmpDir)

		svgFile := filepath.Join(tmpDir, "input.svg")
		pngFile := filepath.Join(tmpDir, "output.png")

		if err := os.WriteFile(svgFile, svg, 0644); err != nil {
			return nil, err
		}

		cmd := exec.Command("inkscape", svgFile, "--export-type=png", "-o", pngFile, "-d", "150")
		if err := cmd.Run(); err != nil {
			return nil, err
		}

		return os.ReadFile(pngFile)
	}

	// 如果没有本地工具，返回错误
	return nil, fmt.Errorf("no SVG to PNG converter available (install rsvg-convert or inkscape)")
}

// RenderMathMermaidAPI 使用mermaid.ink API渲染公式（备用方案）
func RenderMathMermaidAPI(latex string) ([]byte, error) {
	// mermaid.ink 支持数学公式渲染
	// 格式: https://mermaid.ink/img/{base64编码的mermaid图}

	mermaidCode := fmt.Sprintf("graph LR\n  A[\"$$%s$$\"]", latex)

	// base64编码
	encoded := base64.URLEncoding.EncodeToString([]byte(mermaidCode))

	apiURL := fmt.Sprintf("https://mermaid.ink/img/%s", encoded)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// MathJaxConfig MathJax配置
type MathJaxConfig struct {
	Math      string `json:"math"`
	Format    string `json:"format"`
	SVG       bool   `json:"svg"`
	MML       bool   `json:"mml"`
	PNG       bool   `json:"png"`
	SpeechTag string `json:"speakText"`
}

// RenderMathJaxAPI 使用自定义API渲染（如果有部署的话）
func RenderMathJaxAPI(apiURL string, latex string, display bool) ([]byte, error) {
	config := MathJaxConfig{
		Math:   latex,
		Format: "TeX",
		PNG:    true,
	}

	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// ExtractInlineFormulas 提取行内公式
func ExtractInlineFormulas(text string) []struct {
	Start   int
	End     int
	Formula string
} {
	var formulas []struct {
		Start   int
		End     int
		Formula string
	}

	// 匹配 $...$ 格式的行内公式
	inFormula := false
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '$' {
			if !inFormula {
				// 检查是否是 $$（块级公式开始）
				if i+1 < len(text) && text[i+1] == '$' {
					continue
				}
				inFormula = true
				start = i + 1
			} else {
				formulas = append(formulas, struct {
					Start   int
					End     int
					Formula string
				}{
					Start:   start - 1,
					End:     i + 1,
					Formula: text[start:i],
				})
				inFormula = false
			}
		}
	}

	return formulas
}

// IsBlockFormula 检查是否是块级公式
func IsBlockFormula(text string) bool {
	trimmed := strings.TrimSpace(text)
	return strings.HasPrefix(trimmed, "$$") && strings.HasSuffix(trimmed, "$$")
}

// ExtractBlockFormula 提取块级公式内容
func ExtractBlockFormula(text string) string {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "$$") && strings.HasSuffix(trimmed, "$$") {
		return strings.TrimSpace(trimmed[2 : len(trimmed)-2])
	}
	return text
}
