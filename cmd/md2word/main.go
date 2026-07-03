package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"md2word/internal/config"
	"md2word/internal/converter"
)

// 由 -ldflags "-X main.Version=... -X main.BuildTime=..." 注入
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	var (
		inputFile  string
		outputFile string
		configFile string
	)

	flag.StringVar(&inputFile, "i", "", "输入Markdown文件路径")
	flag.StringVar(&inputFile, "input", "", "输入Markdown文件路径")
	flag.StringVar(&outputFile, "o", "", "输出DOCX文件路径")
	flag.StringVar(&outputFile, "output", "", "输出DOCX文件路径")
	flag.StringVar(&configFile, "c", "", "配置文件路径")
	flag.StringVar(&configFile, "config", "", "配置文件路径")
	flag.Parse()

	if inputFile == "" {
		fmt.Fprintln(os.Stderr, "错误: 请指定输入文件 (-i)")
		flag.Usage()
		os.Exit(1)
	}

	if outputFile == "" {
		// 默认输出文件名
		ext := filepath.Ext(inputFile)
		outputFile = inputFile[:len(inputFile)-len(ext)] + ".docx"
	}

	// 加载配置：-c > ./config.yaml > $EXE_DIR/config.yaml > 内置默认
	cfg, source, err := config.LoadConfigWithFallback(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("加载配置: %s\n", source)

	// 读取Markdown文件
	mdContent, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取文件失败: %v\n", err)
		os.Exit(1)
	}

	// 转换
	conv := converter.NewConverter(cfg)
	if err := conv.Convert(mdContent, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "转换失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("转换成功: %s -> %s\n", inputFile, outputFile)
}
