# md2word

Go 语言实现的 Markdown 转 DOCX 工具，支持流程图、图片、表格、公式等丰富内容，专为中文排版优化。

## ✨ 特性

- 🎯 **完整 Markdown 支持**：标题、列表、表格、代码块、引用等
- 🔢 **智能自动编号**：将 Markdown 编号标题转换为 Word 可编辑的自动编号
- 📊 **Mermaid 流程图**：使用 chromedp 离线渲染，无需外部工具
- 🧮 **数学公式**：MathJax 渲染为高清图片，智能尺寸适配
- 🖼️ **智能图片处理**：支持本地/网络/Base64 图片，自动格式检测
- 🔗 **原生超链接**：生成可点击的 Word 超链接
- 🎨 **灵活样式配置**：通过 YAML 文件自定义字体、字号、行距、缩进
- 📝 **中文排版优化**：默认宋体正文、黑体标题，支持首行缩进
- 🖥️ **图形界面版本**：提供现代化的 GUI 界面，支持 Windows 和 macOS
- 🚀 **开箱即用**：配置内嵌于二进制，单文件即可运行

## 📦 安装

### 方式一：下载预编译版本

从 [Releases](https://github.com/GYJ99/md2word/releases) 页面下载对应平台的版本：

- **Windows**: `md2word-windows.exe` (GUI) 或 `md2word-cli.exe` (命令行)
- **macOS**: `md2word-macos.app` (GUI) 或 `md2word-cli` (命令行)
- **Linux**: `md2word-linux.tar.xz` (GUI) 或 `md2word-cli` (命令行)

### 方式二：从源码构建

```bash
# 克隆项目
git clone https://github.com/GYJ99/md2word.git
cd md2word

# 构建命令行版本
go build -o md2word ./cmd/md2word

# 构建GUI版本
go build -o md2word-gui ./cmd/md2word-gui

# 安装到 $GOPATH/bin
go install ./cmd/md2word
go install ./cmd/md2word-gui
```

## 🚀 使用方法

### 图形界面版本（推荐）

1. **启动应用**：双击 `md2word-gui` 或 `md2word.app`
2. **选择文件**：点击"选择文件"按钮选择 Markdown 文件
3. **设置输出**：选择输出位置（可选，默认与输入文件同目录）
4. **选择配置**：选择配置文件或使用默认配置
5. **开始转换**：点击"开始转换"按钮
6. **查看结果**：转换完成后会显示成功提示

### 命令行版本

#### 基本用法

```bash
md2word -i input.md -o output.docx
```

#### 使用自定义配置

```bash
md2word -i input.md -o output.docx -c config.yaml
```

#### 快速测试

```bash
# 运行综合测试
./md2word -i test/comprehensive_test.md -o test/comprehensive_test.docx -c config.yaml

# 运行完整功能测试
./md2word -i test/sample.md -o test/sample.docx -c config.yaml
```

### 命令行参数

| 参数 | 说明 |
|------|------|
| `-i, --input` | 输入 Markdown 文件路径（必需） |
| `-o, --output` | 输出 DOCX 文件路径（可选，默认与输入同名） |
| `-c, --config` | 配置文件路径（可选） |

### 配置加载优先级

程序按以下顺序查找配置文件：
1. **命令行参数** `-c config.yaml`
2. **当前目录** `./config.yaml`
3. **安装目录** `$EXE_DIR/config.yaml`
4. **内置默认** 编译时嵌入的配置

## ⚙️ 配置文件

配置文件使用 YAML 格式，支持丰富的样式定制：

```yaml
# 样式配置
styles:
  body:
    font: "宋体"
    size: 10.5           # 五号字
    lineHeight: 360      # 1.5倍行距 (twips, 240=单倍)
    firstLineIndent: 420 # 首行缩进 (twips, 约2字符)

  heading1:
    font: "黑体"
    size: 16             # 三号字
    bold: true

  code:
    font: "Consolas"
    size: 8
    background: "#f5f5f5"

# 表格配置
table:
  font: "宋体"
  size: 10.5
  borders: true
  headerBold: true

# Mermaid 流程图
mermaid:
  enabled: true
  theme: "default"

# 数学公式
math:
  enabled: true
  render: "image"

# 图片配置
images:
  maxWidth: 600
  downloadTimeout: 30
```

### 单位说明

- **字号 (size)**：pt（磅）
- **间距/行高/缩进**：twips（1/20 磅）
  - `240` twips = 12pt = 单倍行距
  - `360` twips = 18pt = 1.5倍行距
  - `420` twips ≈ 2字符首行缩进（基于五号字）

## 📋 默认样式

| 元素 | 字体 | 字号 | 说明 |
|------|------|------|------|
| 正文 | 宋体 | 10.5pt (五号) | 1.5倍行距，首行缩进2字符 |
| 一级标题 | 黑体 | 16pt (三号) | 加粗 |
| 二级标题 | 黑体 | 15pt (小三) | 加粗 |
| 三级标题 | 黑体 | 15pt (小三) | 加粗 |
| 四-九级标题 | 黑体 | 14pt (四号) | 加粗 |
| 代码块 | Consolas | 8pt | 浅灰背景 |

## 🛠️ 环境要求

### 必需

- **Go 1.18+**：用于编译构建（仅开发时需要）
- **Chrome/Chromium 浏览器**：用于 Mermaid 流程图渲染

### GUI 版本额外要求

- **Windows**: Windows 10 或更高版本
- **macOS**: macOS 10.14 或更高版本  
- **Linux**: 支持 GTK3 的现代 Linux 发行版

### 安装 Chrome (macOS)

```bash
brew install --cask google-chrome
```

## 📁 项目结构

```
md2word/
├── cmd/
│   ├── md2word/           # CLI 入口
│   │   └── main.go
│   └── md2word-gui/       # GUI 入口
│       └── main.go
├── internal/
│   ├── config/            # 配置管理
│   │   ├── config.go
│   │   └── default.yaml   # 内嵌默认配置
│   ├── converter/         # 核心转换器
│   │   ├── converter.go   # 主转换逻辑
│   │   ├── code_native.go # 代码高亮
│   │   └── ...
│   ├── docx/              # DOCX 生成
│   │   ├── document.go    # 文档结构
│   │   ├── paragraph.go   # 段落/超链接
│   │   ├── table.go       # 表格元素
│   │   └── styles.go      # 样式定义
│   └── parser/            # Markdown 解析
├── scripts/               # 构建脚本
│   ├── build-gui.sh       # GUI 构建脚本
│   └── package-gui.sh     # 打包脚本
├── assets/                # 应用资源
│   └── icon.png           # 应用图标
├── config.yaml            # 示例配置
├── test/                  # 测试文件
│   └── comprehensive_test.md
├── FyneApp.toml           # GUI 应用配置
├── GUI_BUILD.md           # GUI 构建指南
└── README.md
```

## 🔧 开发

### 构建

```bash
go build -o md2word ./cmd/md2word
```

### 测试

```bash
go test ./...
```

### 运行示例

```bash
# 综合测试（推荐）
./md2word -i test/comprehensive_test.md -o test/comprehensive_test.docx

# 完整功能测试
./md2word -i test/sample.md -o test/output.docx

# GUI 版本
./md2word-gui
```

## 🖥️ GUI 版本

GUI 版本提供了友好的图形界面，特别适合不熟悉命令行的用户。

### 主要特性

- 🎯 **直观操作**：拖拽或点击选择文件
- 📊 **实时进度**：可视化转换进度
- 📝 **详细日志**：显示转换过程和错误信息
- ⚙️ **配置管理**：支持自定义配置或默认配置
- 🎨 **现代界面**：Material Design 风格

### 构建 GUI 版本

详细的 GUI 构建指南请参考 [GUI_BUILD.md](GUI_BUILD.md)

```bash
# 构建 GUI 版本
go build -o md2word-gui ./cmd/md2word-gui

# 使用构建脚本
./scripts/build-gui.sh

# 打包安装程序
./scripts/package-gui.sh
```

## 🔢 自动编号功能

md2word 支持将 Markdown 中的编号标题转换为 Word 的自动编号标题，实现真正可编辑的编号系统。

### 支持的编号格式

```markdown
## **2.1 主要章节**
### **2.1.1 子章节**
#### **2.1.1.1 详细内容**
```

### 编号规则

- **智能分组**：相同顶级编号的标题使用同一编号序列
  - 例如：2.1、2.1.1、2.2 属于"2"组
  - 而 3.1、3.1.1 属于"3"组
- **跳跃编号**：支持任意起始编号（如直接从 5.1 开始）
- **自动编号**：在 Word 中显示为可编辑的自动编号，支持插入和重新排序
- **文本分离**：自动移除原始编号，只保留标题文本

### 示例效果

| Markdown 输入 | Word 输出 |
|---------------|-----------|
| `## **2.1 系统设计**` | 2.1 系统设计（自动编号） |
| `### **2.1.1 架构说明**` | 2.1.1 架构说明（自动编号） |
| `## **3.1 测试计划**` | 3.1 测试计划（新编号序列） |

## 📝 支持的 Markdown 语法

- [✅] 标题 (1-9级)
- [✅] **自动编号标题** (如 `## **2.1 标题**`)
- [✅] 段落和文本格式 (加粗、斜体、删除线)
- [✅] 有序/无序列表
- [✅] 表格 (GFM 格式)
- [✅] 代码块 (语法高亮)
- [✅] 行内代码
- [✅] 超链接
- [✅] 图片 (本地/网络/Base64)
- [✅] 引用块
- [✅] 分隔线
- [✅] Mermaid 流程图
- [✅] 数学公式 ($...$, $$...$$)

## 📄 License

MIT License
