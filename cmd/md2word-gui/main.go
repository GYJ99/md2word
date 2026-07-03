package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/layout"

	"md2word/internal/config"
	"md2word/internal/converter"
)

type App struct {
	fyneApp     fyne.App
	window      fyne.Window
	inputPath   *widget.Entry
	outputPath  *widget.Entry
	configPath  *widget.Entry
	logText     *widget.Entry
	convertBtn  *widget.Button
	progressBar *widget.ProgressBar
	statusLabel *widget.Label
	fileInfo    *widget.Card
}

func main() {
	app := &App{}
	app.setupUI()
	app.window.ShowAndRun()
}

func (a *App) setupUI() {
	a.fyneApp = app.New()
	a.fyneApp.SetIcon(theme.DocumentIcon())
	
	// 设置应用主题
	a.fyneApp.Settings().SetTheme(&customTheme{})
	
	a.window = a.fyneApp.NewWindow("md2word - Markdown 转 Word 工具")
	a.window.Resize(fyne.NewSize(900, 700))
	a.window.CenterOnScreen()
	a.window.SetFixedSize(false)

	// 创建UI组件
	a.createComponents()
	
	// 设置拖拽支持
	a.setupDragAndDrop()
	
	// 创建主布局
	content := a.createMainLayout()
	a.window.SetContent(content)
	
	// 设置快捷键
	a.setupShortcuts()
}

func (a *App) createComponents() {
	// 输入路径
	a.inputPath = widget.NewEntry()
	a.inputPath.SetPlaceHolder("选择 Markdown 文件或拖拽到此处...")
	a.inputPath.Validator = func(s string) error {
		if s != "" && !strings.HasSuffix(strings.ToLower(s), ".md") && !strings.HasSuffix(strings.ToLower(s), ".markdown") {
			return fmt.Errorf("请选择 Markdown 文件 (.md 或 .markdown)")
		}
		return nil
	}
	
	// 输出路径
	a.outputPath = widget.NewEntry()
	a.outputPath.SetPlaceHolder("输出 Word 文件路径（可选，默认同目录）...")
	
	// 配置路径
	a.configPath = widget.NewEntry()
	a.configPath.SetPlaceHolder("配置文件路径（可选，使用默认配置）...")
	
	// 富文本日志
	a.logText = widget.NewMultiLineEntry()
	a.logText.Wrapping = fyne.TextWrapWord
	a.logText.SetPlaceHolder("转换日志将显示在这里...")
	
	// 转换按钮
	a.convertBtn = widget.NewButton("🚀 开始转换", a.handleConvert)
	a.convertBtn.Importance = widget.HighImportance
	a.convertBtn.Resize(fyne.NewSize(200, 50))
	
	// 进度条
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()
	
	// 状态标签
	a.statusLabel = widget.NewLabel("就绪")
	a.statusLabel.Alignment = fyne.TextAlignCenter
	
	// 文件信息卡片
	a.fileInfo = widget.NewCard("", "", widget.NewLabel("选择文件后显示详细信息"))
	a.fileInfo.Hide()
}

func (a *App) createMainLayout() *fyne.Container {
	// 头部区域
	header := a.createHeader()
	
	// 文件选择区域
	fileSection := a.createFileSection()
	
	// 配置区域
	configSection := a.createConfigSection()
	
	// 操作区域
	actionSection := a.createActionSection()
	
	// 日志区域
	logSection := a.createLogSection()
	
	// 状态栏
	statusBar := a.createStatusBar()
	
	// 主内容区域
	mainContent := container.NewVBox(
		header,
		widget.NewSeparator(),
		fileSection,
		widget.NewSeparator(),
		configSection,
		widget.NewSeparator(),
		actionSection,
		widget.NewSeparator(),
		logSection,
	)
	
	// 使用 Border 布局，底部放状态栏
	return container.NewBorder(
		nil, statusBar, nil, nil,
		container.NewScroll(mainContent),
	)
}

func (a *App) createHeader() *fyne.Container {
	// 应用图标和标题
	icon := widget.NewIcon(theme.DocumentIcon())
	title := widget.NewLabelWithStyle("md2word", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	
	subtitle := widget.NewLabelWithStyle("专业的 Markdown 转 Word 文档工具", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	
	headerContent := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), icon, title, layout.NewSpacer()),
		subtitle,
	)
	
	return container.NewPadded(headerContent)
}

func (a *App) createFileSection() *fyne.Container {
	// 输入文件区域
	inputLabel := widget.NewLabelWithStyle("📄 输入文件", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	inputBtn := widget.NewButtonWithIcon("浏览", theme.FolderOpenIcon(), func() {
		a.selectInputFile()
	})
	
	clearInputBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		a.inputPath.SetText("")
		a.fileInfo.Hide()
		a.updateStatus("就绪")
	})
	clearInputBtn.Importance = widget.LowImportance
	
	inputContainer := container.NewBorder(
		nil, nil, 
		inputLabel,
		container.NewHBox(clearInputBtn, inputBtn),
		a.inputPath,
	)
	
	// 输出文件区域
	outputLabel := widget.NewLabelWithStyle("💾 输出文件", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	outputBtn := widget.NewButtonWithIcon("浏览", theme.DocumentSaveIcon(), func() {
		a.selectOutputFile()
	})
	
	autoBtn := widget.NewButtonWithIcon("自动", theme.SettingsIcon(), func() {
		if a.inputPath.Text != "" {
			outputFile := strings.TrimSuffix(a.inputPath.Text, filepath.Ext(a.inputPath.Text)) + ".docx"
			a.outputPath.SetText(outputFile)
		}
	})
	autoBtn.Importance = widget.LowImportance
	
	outputContainer := container.NewBorder(
		nil, nil,
		outputLabel,
		container.NewHBox(autoBtn, outputBtn),
		a.outputPath,
	)
	
	// 文件信息显示
	fileSection := container.NewVBox(
		inputContainer,
		outputContainer,
		a.fileInfo,
	)
	
	return container.NewPadded(fileSection)
}

func (a *App) createConfigSection() *fyne.Container {
	configLabel := widget.NewLabelWithStyle("⚙️ 配置选项", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	configBtn := widget.NewButtonWithIcon("选择配置", theme.SettingsIcon(), func() {
		a.selectConfigFile()
	})
	
	defaultBtn := widget.NewButtonWithIcon("默认配置", theme.HomeIcon(), func() {
		a.configPath.SetText("")
		a.appendLog("✅ 将使用内置默认配置")
	})
	defaultBtn.Importance = widget.LowImportance
	
	configContainer := container.NewBorder(
		nil, nil,
		configLabel,
		container.NewHBox(defaultBtn, configBtn),
		a.configPath,
	)
	
	return container.NewPadded(configContainer)
}

func (a *App) createActionSection() *fyne.Container {
	// 进度信息
	progressContainer := container.NewVBox(
		a.progressBar,
		container.NewCenter(a.statusLabel),
	)
	
	// 按钮区域
	buttonContainer := container.NewCenter(a.convertBtn)
	
	return container.NewPadded(
		container.NewVBox(
			progressContainer,
			buttonContainer,
		),
	)
}

func (a *App) createLogSection() *fyne.Container {
	logLabel := widget.NewLabelWithStyle("📋 转换日志", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	clearBtn := widget.NewButtonWithIcon("清空", theme.DeleteIcon(), func() {
		a.logText.SetText("")
		a.updateStatus("日志已清空")
	})
	clearBtn.Importance = widget.LowImportance
	
	exportBtn := widget.NewButtonWithIcon("导出", theme.DocumentSaveIcon(), func() {
		a.exportLog()
	})
	exportBtn.Importance = widget.LowImportance
	
	logHeader := container.NewBorder(
		nil, nil,
		logLabel,
		container.NewHBox(clearBtn, exportBtn),
	)
	
	logScroll := container.NewScroll(a.logText)
	logScroll.SetMinSize(fyne.NewSize(0, 200))
	
	return container.NewPadded(
		container.NewBorder(
			logHeader, nil, nil, nil,
			logScroll,
		),
	)
}

func (a *App) createStatusBar() *fyne.Container {
	timeLabel := widget.NewLabel(time.Now().Format("2006-01-02 15:04:05"))
	
	// 定时更新时间
	go func() {
		for {
			time.Sleep(time.Second)
			timeLabel.SetText(time.Now().Format("2006-01-02 15:04:05"))
		}
	}()
	
	versionLabel := widget.NewLabel("v1.0.0")
	
	return container.NewBorder(
		nil, nil,
		container.NewHBox(widget.NewLabel("状态:"), a.statusLabel),
		container.NewHBox(versionLabel, widget.NewSeparator(), timeLabel),
	)
}

func (a *App) setupDragAndDrop() {
	// 这里可以添加拖拽支持的代码
	// Fyne 的拖拽支持可能需要额外的实现
}

func (a *App) setupShortcuts() {
	// Ctrl+O 打开文件
	a.window.Canvas().AddShortcut(&fyne.ShortcutCut{}, func(shortcut fyne.Shortcut) {
		a.selectInputFile()
	})
	
	// Ctrl+S 开始转换
	a.window.Canvas().AddShortcut(&fyne.ShortcutPaste{}, func(shortcut fyne.Shortcut) {
		if a.convertBtn.Visible() {
			a.handleConvert()
		}
	})
}

func (a *App) selectInputFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			a.showError("选择文件失败: " + err.Error())
			return
		}
		if reader != nil {
			path := reader.URI().Path()
			a.inputPath.SetText(path)
			reader.Close()
			
			// 显示文件信息
			a.showFileInfo(path)
			
			// 自动设置输出路径
			if a.outputPath.Text == "" {
				outputFile := strings.TrimSuffix(path, filepath.Ext(path)) + ".docx"
				a.outputPath.SetText(outputFile)
			}
			
			a.updateStatus("文件已选择")
		}
	}, a.window)
}

func (a *App) selectOutputFile() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			a.showError("选择保存位置失败: " + err.Error())
			return
		}
		if writer != nil {
			path := writer.URI().Path()
			if !strings.HasSuffix(strings.ToLower(path), ".docx") {
				path += ".docx"
			}
			a.outputPath.SetText(path)
			writer.Close()
			a.updateStatus("输出路径已设置")
		}
	}, a.window)
}

func (a *App) selectConfigFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			a.showError("选择配置文件失败: " + err.Error())
			return
		}
		if reader != nil {
			a.configPath.SetText(reader.URI().Path())
			reader.Close()
			a.updateStatus("配置文件已选择")
		}
	}, a.window)
}

func (a *App) showFileInfo(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	
	size := info.Size()
	sizeStr := formatFileSize(size)
	modTime := info.ModTime().Format("2006-01-02 15:04:05")
	
	infoText := fmt.Sprintf("文件名: %s\n大小: %s\n修改时间: %s", 
		filepath.Base(path), sizeStr, modTime)
	
	a.fileInfo.SetTitle("📁 文件信息")
	a.fileInfo.SetContent(widget.NewLabel(infoText))
	a.fileInfo.Show()
}

func (a *App) handleConvert() {
	// 验证输入
	if a.inputPath.Text == "" {
		a.showError("请选择输入的 Markdown 文件")
		return
	}
	
	if a.outputPath.Text == "" {
		a.showError("请设置输出文件路径")
		return
	}
	
	// 检查输入文件是否存在
	if _, err := os.Stat(a.inputPath.Text); os.IsNotExist(err) {
		a.showError("输入文件不存在: " + a.inputPath.Text)
		return
	}
	
	// 开始转换
	a.startConversion()
}

func (a *App) startConversion() {
	// 禁用转换按钮，显示进度条
	a.convertBtn.Disable()
	a.convertBtn.SetText("转换中...")
	a.progressBar.Show()
	a.progressBar.SetValue(0)
	a.updateStatus("正在转换...")
	
	// 清空日志
	a.logText.SetText("")
	a.appendLog("🚀 开始转换...")
	
	// 在goroutine中执行转换
	go func() {
		defer func() {
			// 恢复UI状态
			a.convertBtn.Enable()
			a.convertBtn.SetText("🚀 开始转换")
			a.progressBar.Hide()
		}()
		
		// 执行转换
		err := a.performConversion()
		
		if err != nil {
			a.appendLog(fmt.Sprintf("❌ 转换失败: %v", err))
			a.updateStatus("转换失败")
			a.showError("转换失败: " + err.Error())
		} else {
			a.appendLog("✅ 转换成功完成!")
			a.updateStatus("转换完成")
			a.showSuccess("转换成功!", "文件已保存到:\n"+a.outputPath.Text)
		}
	}()
}

func (a *App) performConversion() error {
	a.progressBar.SetValue(0.1)
	a.appendLog("📋 加载配置...")
	
	// 加载配置
	var cfg *config.Config
	var err error
	
	if a.configPath.Text != "" {
		cfg, err = config.LoadConfig(a.configPath.Text)
		if err != nil {
			return fmt.Errorf("加载配置文件失败: %w", err)
		}
		a.appendLog("⚙️ 使用自定义配置: " + filepath.Base(a.configPath.Text))
	} else {
		cfg = config.DefaultConfig()
		a.appendLog("⚙️ 使用默认配置")
	}
	
	a.progressBar.SetValue(0.3)
	a.appendLog("📖 读取输入文件...")
	
	// 读取输入文件
	content, err := os.ReadFile(a.inputPath.Text)
	if err != nil {
		return fmt.Errorf("读取输入文件失败: %w", err)
	}
	
	a.progressBar.SetValue(0.5)
	a.appendLog("🔧 初始化转换器...")
	
	// 创建转换器
	conv := converter.NewConverter(cfg)
	defer conv.Close()
	
	a.progressBar.SetValue(0.7)
	a.appendLog("🔄 执行转换...")
	
	// 执行转换
	err = conv.Convert(content, a.outputPath.Text)
	if err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}
	
	a.progressBar.SetValue(1.0)
	return nil
}

func (a *App) appendLog(message string) {
	current := a.logText.Text
	if current != "" {
		current += "\n"
	}
	current += message
	a.logText.SetText(current)
	
	// 滚动到底部
	a.logText.CursorRow = len(strings.Split(current, "\n"))
}

func (a *App) updateStatus(status string) {
	a.statusLabel.SetText(status)
}

func (a *App) exportLog() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			a.showError("导出日志失败: " + err.Error())
			return
		}
		if writer != nil {
			defer writer.Close()
			_, err := writer.Write([]byte(a.logText.Text))
			if err != nil {
				a.showError("写入日志文件失败: " + err.Error())
			} else {
				a.showSuccess("导出成功", "日志已保存到: "+writer.URI().Path())
			}
		}
	}, a.window)
}

func (a *App) showError(message string) {
	dialog.ShowError(fmt.Errorf(message), a.window)
}

func (a *App) showSuccess(title, message string) {
	dialog.ShowInformation(title, message, a.window)
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// 自定义主题
type customTheme struct{}

func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x42, G: 0x85, B: 0xf4, A: 0xff} // Google Blue
	case theme.ColorNameBackground:
		if variant == theme.VariantLight {
			return color.RGBA{R: 0xf8, G: 0xf9, B: 0xfa, A: 0xff} // Light gray
		}
		return color.RGBA{R: 0x1e, G: 0x1e, B: 0x1e, A: 0xff} // Dark gray
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *customTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	}
	return theme.DefaultTheme().Size(name)
}