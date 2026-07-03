# md2word

Go 语言实现的 Markdown 转 DOCX 工具，支持流程图、图片、表格、公式等丰富内容，专为中文排版优化。

## ✨ 特性

- 🎯 **完整 Markdown 支持**：标题 (1-9 级)、列表、表格、代码块、引用等
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

### 方式一：从源码构建（一键脚本）

```bash
git clone https://github.com/GYJ99/md2word.git
cd md2word

# 同时编译当前平台的 CLI 和 GUI，输出到 dist-cli/ 和 dist-gui/
./scripts/build.sh
```

产物：
- CLI: `dist-cli/md2word-<os>-<arch>[.exe]`（静态二进制，约 15M）
- GUI: `dist-gui/md2word-gui-<os>-<arch>[.exe]`（约 30M，需 CGO）
- 分发包：同目录下自动生成 `.tar.gz`（Unix）或 `.zip`（Windows）

### 方式二：手动编译

```bash
# CLI（静态编译，支持交叉编译）
CGO_ENABLED=0 go build -o md2word ./cmd/md2word

# GUI（需要 CGO，必须在目标平台编译）
go build -o md2word-gui ./cmd/md2word-gui

# 安装到 $GOPATH/bin
go install ./cmd/md2word
go install ./cmd/md2word-gui
```

## 🚀 使用方法

### 图形界面版本（推荐）

1. **启动应用**：双击 `md2word-gui`（macOS/Linux）或 `md2word-gui.exe`（Windows）
2. **选择文件**：点击"浏览"按钮选择 Markdown 文件
3. **设置输出**：选择输出位置（可选，点击"自动"会填入默认值）
4. **选择配置**：选择配置文件或使用默认配置
5. **开始转换**：点击"🚀 开始转换"按钮（或按 Ctrl+S）
6. **查看结果**：转换完成后会显示成功提示
7. **快捷键**：`Ctrl+O` 打开文件 / `Ctrl+S` 开始转换

### 命令行版本

#### 基本用法

```bash
./dist-cli/md2word-<os>-<arch> -i input.md -o output.docx
```

#### 使用自定义配置

```bash
./dist-cli/md2word-<os>-<arch> -i input.md -o output.docx -c config.yaml
```

#### 命令行参数

| 参数 | 说明 |
|------|------|
| `-i, --input` | 输入 Markdown 文件路径（必需） |
| `-o, --output` | 输出 DOCX 文件路径（可选，默认与输入同名） |
| `-c, --config` | 配置文件路径（可选） |

### 配置加载优先级

CLI 与 GUI 共用同一逻辑，按以下顺序查找配置：

1. **命令行/界面参数** `-c config.yaml`
2. **当前目录** `./config.yaml`
3. **可执行文件所在目录** `$EXE_DIR/config.yaml`
4. **内置默认** 编译时嵌入的配置

## ⚙️ 配置文件

配置文件使用 YAML 格式，支持丰富的样式定制：

```yaml
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

  # heading2 ~ heading9 同结构
  heading7:              # goldmark 默认只支持 1-6 级
    font: "宋体 (标题)"  # 本项目扩展为 1-9 级
    size: 12
    bold: true

  code:
    font: "Consolas"
    size: 8
    background: "#f5f5f5"

table:
  font: "宋体"
  size: 10.5
  borders: true
  headerBold: true

mermaid:
  enabled: true
  theme: "default"

math:
  enabled: true
  render: "image"

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

## 🛠️ 环境要求

- **Go 1.21+**：用于编译构建
- **Chrome/Chromium 浏览器**：用于 Mermaid 流程图渲染

```bash
# macOS 安装 Chrome
brew install --cask google-chrome
```

GUI 版本仅支持在目标平台编译（Fyne 依赖 OpenGL 和平台特定的 C 库）。

## 📁 项目结构

```
md2word/
├── cmd/
│   ├── md2word/         # CLI 入口
│   └── md2word-gui/     # GUI 入口 (Fyne)
├── internal/
│   ├── config/          # 配置管理（含内嵌 default.yaml）
│   │   ├── config.go
│   │   └── default.yaml
│   ├── converter/       # 核心转换器
│   │   ├── converter.go
│   │   ├── code_native.go
│   │   ├── code_renderer.go
│   │   ├── chrome_finder.go
│   │   ├── math.go
│   │   ├── mermaid.go
│   │   └── mermaid_chromedp.go
│   ├── docx/            # DOCX 生成
│   │   ├── document.go
│   │   ├── numbering.go
│   │   ├── paragraph.go
│   │   ├── styles.go
│   │   └── table.go
│   └── parser/          # Markdown 解析
│       ├── markdown.go
│       └── deep_atx_heading.go   # 扩展支持 7-9 级标题
├── scripts/
│   └── build.sh         # 一键打包 CLI + GUI
├── .github/workflows/   # CI/CD
├── assets/              # 应用资源
│   └── icon.png
├── config.yaml          # 默认配置示例
├── config_anke.yaml     # 其它场景示例
├── config_yunnan.yaml
├── FyneApp.toml
└── README.md
```

## 🔧 开发

### 重新打包

```bash
./scripts/build.sh
# 或指定版本号
VERSION=1.0.1 ./scripts/build.sh
```

### 测试

```bash
go test ./...
```

### 跨平台打包（仅 CLI）

GUI 必须在目标平台本地编译，CLI 支持交叉编译：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o md2word-linux ./cmd/md2word
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o md2word.exe ./cmd/md2word
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

- [✅] 标题 (1-9 级)
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
