package config

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfigData []byte

// StyleConfig 样式配置
type StyleConfig struct {
	Font            string  `yaml:"font"`
	Size            float64 `yaml:"size"`
	Bold            bool    `yaml:"bold"`
	Italic          bool    `yaml:"italic"`
	Color           string  `yaml:"color"`
	Background      string  `yaml:"background"`
	LineSpacing     int     `yaml:"lineSpacing"`     // 行间距 (twips) - 已弃用，使用 SpaceBefore/SpaceAfter
	LineHeight      int     `yaml:"lineHeight"`      // 行高 (twips, 240=1倍, 360=1.5倍)
	SpaceBefore     int     `yaml:"spaceBefore"`     // 段前间距 (twips, 20=1pt)
	SpaceAfter      int     `yaml:"spaceAfter"`      // 段后间距 (twips)
	FirstLineIndent int     `yaml:"firstLineIndent"` // 首行缩进 (twips, 210=10.5pt=1字符(五号))
}

// TableConfig 表格配置
type TableConfig struct {
	Font       string  `yaml:"font"`
	Size       float64 `yaml:"size"`
	Borders    bool    `yaml:"borders"`
	HeaderBold bool    `yaml:"headerBold"`
}

// MermaidConfig Mermaid配置
type MermaidConfig struct {
	Enabled bool   `yaml:"enabled"`
	CLI     string `yaml:"cli"`
	Theme   string `yaml:"theme"`
}

// MathConfig 数学公式配置
type MathConfig struct {
	Enabled bool   `yaml:"enabled"`
	Render  string `yaml:"render"` // "mathjax" or "image"
}

// ImageConfig 图片配置
type ImageConfig struct {
	MaxWidth        int `yaml:"maxWidth"`
	DownloadTimeout int `yaml:"downloadTimeout"`
}

// Config 完整配置
type Config struct {
	Styles struct {
		Body      StyleConfig `yaml:"body"`
		Heading1  StyleConfig `yaml:"heading1"`
		Heading2  StyleConfig `yaml:"heading2"`
		Heading3  StyleConfig `yaml:"heading3"`
		Heading4  StyleConfig `yaml:"heading4"`
		Heading5  StyleConfig `yaml:"heading5"`
		Heading6  StyleConfig `yaml:"heading6"`
		Heading7  StyleConfig `yaml:"heading7"`
		Heading8  StyleConfig `yaml:"heading8"`
		Heading9  StyleConfig `yaml:"heading9"`
		Code      StyleConfig `yaml:"code"`
		CodeBlock StyleConfig `yaml:"codeBlock"`
	} `yaml:"styles"`
	Table   TableConfig   `yaml:"table"`
	Mermaid MermaidConfig `yaml:"mermaid"`
	Math    MathConfig    `yaml:"math"`
	Images  ImageConfig   `yaml:"images"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	var cfg Config
	if err := yaml.Unmarshal(defaultConfigData, &cfg); err != nil {
		// 如果解析嵌入的配置失败，这通常是编译时的错误（default.yaml内容有问题）
		// 在生产环境中，为了健壮性，这里应该 panic 或者打印严重错误
		panic(fmt.Sprintf("internal error: failed to parse embedded default config: %v", err))
	}
	return &cfg
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// GetHeadingStyle 获取标题样式
func (c *Config) GetHeadingStyle(level int) StyleConfig {
	switch level {
	case 1:
		return c.Styles.Heading1
	case 2:
		return c.Styles.Heading2
	case 3:
		return c.Styles.Heading3
	case 4:
		return c.Styles.Heading4
	case 5:
		return c.Styles.Heading5
	case 6:
		return c.Styles.Heading6
	case 7:
		return c.Styles.Heading7
	case 8:
		return c.Styles.Heading8
	case 9:
		return c.Styles.Heading9
	default:
		return c.Styles.Body
	}
}
