package docx

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// HeadingNumber 标题编号信息
type HeadingNumber struct {
	Level  int    // 编号级别 (1-9)
	Values []int  // 各级编号值 [2, 1, 1] 表示 2.1.1
	Text   string // 移除编号后的标题文本
}

// NumberingState 编号状态跟踪
type NumberingState struct {
	currentValues []int                    // 当前各级编号值
	lastLevel     int                      // 上一个标题的级别
	numInstances  map[string]*NumInstance // 编号实例映射
	nextNumId     int                      // 下一个可用的编号ID
}

// NumInstance 编号实例
type NumInstance struct {
	NumId       int     // 编号ID
	StartValues []int   // 各级起始值
	AbstractId  int     // 抽象编号ID
}

// NewNumberingState 创建编号状态
func NewNumberingState() *NumberingState {
	return &NumberingState{
		currentValues: make([]int, 9), // 支持9级标题
		lastLevel:     0,
		numInstances:  make(map[string]*NumInstance),
		nextNumId:     1,
	}
}

// 编号解析正则: 匹配 "2.1 标题" 或 "2.1.1 标题" 格式
var headingNumberPattern = regexp.MustCompile(`^\s*(\d+(?:\.\d+)*)\s+(.*)$`)

// ParseHeadingNumber 解析标题中的编号
// 输入: "2.1 CME智能体子系统" 或 "**2.1.1 系统需求分析**"
// 返回: HeadingNumber 结构，如果没有编号则返回 nil
func ParseHeadingNumber(text string) *HeadingNumber {
	// 先移除 markdown 加粗标记
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "**")
	text = strings.TrimSuffix(text, "**")
	text = strings.TrimSpace(text)

	matches := headingNumberPattern.FindStringSubmatch(text)
	if matches == nil {
		return nil // 没有编号
	}

	numberStr := matches[1] // "2.1" 或 "2.1.1"
	titleText := matches[2] // "CME智能体子系统"

	// 解析编号值
	parts := strings.Split(numberStr, ".")
	values := make([]int, len(parts))
	for i, part := range parts {
		val, err := strconv.Atoi(part)
		if err != nil {
			return nil // 解析失败
		}
		values[i] = val
	}

	// 再次清理标题文本，因为正则分组可能捕获 trailing "**"
	titleText = strings.TrimSuffix(strings.TrimSpace(titleText), "**")

	return &HeadingNumber{
		Level:  len(values),
		Values: values,
		Text:   titleText,
	}
}

// GenerateNumberingXML 生成 numbering.xml 内容
// 支持多级列表编号，每级可以有不同的起始值
func GenerateNumberingXML(numInstances map[string]*NumInstance) string {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)

	// 如果没有编号实例，创建一个默认的
	if len(numInstances) == 0 {
		buf.WriteString(`
    <!-- 默认抽象编号定义 -->
    <w:abstractNum w:abstractNumId="0">
        <w:multiLevelType w:val="multilevel"/>`)

		// 生成9级编号定义
		for level := 0; level < 9; level++ {
			buf.WriteString(fmt.Sprintf(`
        <w:lvl w:ilvl="%d">
            <w:start w:val="1"/>
            <w:numFmt w:val="decimal"/>
            <w:lvlText w:val="%s"/>
            <w:lvlJc w:val="left"/>
            <w:pPr>
                <w:ind w:left="0" w:hanging="0"/>
            </w:pPr>
        </w:lvl>`, level, generateLevelText(level+1)))
		}

		buf.WriteString(`
    </w:abstractNum>
    
    <!-- 默认编号实例 -->
    <w:num w:numId="1">
        <w:abstractNumId w:val="0"/>
    </w:num>`)
	} else {
		// 生成抽象编号定义
		abstractIds := make(map[int]bool)
		for _, instance := range numInstances {
			if !abstractIds[instance.AbstractId] {
				abstractIds[instance.AbstractId] = true
				buf.WriteString(fmt.Sprintf(`
    <!-- 抽象编号定义 %d -->
    <w:abstractNum w:abstractNumId="%d">
        <w:multiLevelType w:val="multilevel"/>`, instance.AbstractId, instance.AbstractId))

				// 生成9级编号定义
				for level := 0; level < 9; level++ {
					startVal := 1
					if level < len(instance.StartValues) && instance.StartValues[level] > 0 {
						startVal = instance.StartValues[level]
					}
					
					buf.WriteString(fmt.Sprintf(`
        <w:lvl w:ilvl="%d">
            <w:start w:val="%d"/>
            <w:numFmt w:val="decimal"/>
            <w:lvlText w:val="%s"/>
            <w:lvlJc w:val="left"/>
            <w:pPr>
                <w:ind w:left="0" w:hanging="0"/>
            </w:pPr>
        </w:lvl>`, level, startVal, generateLevelText(level+1)))
				}

				buf.WriteString(`
    </w:abstractNum>`)
			}
		}

		// 生成编号实例
		for _, instance := range numInstances {
			buf.WriteString(fmt.Sprintf(`
    
    <!-- 编号实例 %d -->
    <w:num w:numId="%d">
        <w:abstractNumId w:val="%d"/>`, instance.NumId, instance.NumId, instance.AbstractId))

			// 为每一级设置起始值覆盖
			for level := 0; level < len(instance.StartValues) && level < 9; level++ {
				if instance.StartValues[level] > 1 {
					buf.WriteString(fmt.Sprintf(`
        <w:lvlOverride w:ilvl="%d">
            <w:startOverride w:val="%d"/>
        </w:lvlOverride>`, level, instance.StartValues[level]))
				}
			}

			buf.WriteString(`
    </w:num>`)
		}
	}

	buf.WriteString(`
</w:numbering>`)

	return buf.String()
}

// generateLevelText 生成编号文本格式
// level 1: "%1"
// level 2: "%1.%2"
// level 3: "%1.%2.%3"
func generateLevelText(level int) string {
	parts := make([]string, level)
	for i := 0; i < level; i++ {
		parts[i] = fmt.Sprintf("%%%d", i+1)
	}
	return strings.Join(parts, ".")
}

// UpdateNumberingState 更新编号状态
// 根据解析的编号值更新当前编号状态
func (ns *NumberingState) UpdateNumberingState(num *HeadingNumber) {
	if num == nil {
		return
	}

	// 更新当前级别的编号值
	for i := 0; i < num.Level && i < len(num.Values); i++ {
		ns.currentValues[i] = num.Values[i]
	}

	// 重置更深层级的编号
	for i := num.Level; i < len(ns.currentValues); i++ {
		ns.currentValues[i] = 0
	}

	ns.lastLevel = num.Level
}

// GetOrCreateNumberingInstance 获取或创建编号实例
func (ns *NumberingState) GetOrCreateNumberingInstance(num *HeadingNumber) *NumInstance {
	if num == nil {
		return nil
	}

	// 对于连续的编号序列，我们需要更智能的实例管理
	// 检查是否是同一个编号序列的延续
	
	// 简化策略：为每个顶级编号（第一级）创建一个实例
	// 例如：2.1, 2.1.1, 2.1.2 都使用同一个实例（以2开头）
	// 而 3.1, 3.1.1 使用另一个实例（以3开头）
	
	topLevelNum := num.Values[0] // 获取顶级编号
	key := fmt.Sprintf("top_%d", topLevelNum)

	// 检查是否已存在
	if instance, exists := ns.numInstances[key]; exists {
		return instance
	}

	// 创建新实例
	instance := &NumInstance{
		NumId:       ns.nextNumId,
		StartValues: make([]int, 9), // 支持9级
		AbstractId:  0, // 暂时都使用同一个抽象编号
	}

	// 设置起始值：只设置第一级的起始值
	instance.StartValues[0] = topLevelNum
	// 其他级别保持默认值1
	for i := 1; i < 9; i++ {
		instance.StartValues[i] = 1
	}

	ns.numInstances[key] = instance
	ns.nextNumId++

	return instance
}

// GetNumberingInstances 获取所有编号实例
func (ns *NumberingState) GetNumberingInstances() map[string]*NumInstance {
	return ns.numInstances
}

// GetNumberingXMLForParagraph 为段落生成编号属性XML
func GetNumberingXMLForParagraph(level int, numId int) string {
	if level < 0 || level > 8 || numId <= 0 {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString(`<w:numPr>`)
	buf.WriteString(fmt.Sprintf(`<w:ilvl w:val="%d"/>`, level))
	buf.WriteString(fmt.Sprintf(`<w:numId w:val="%d"/>`, numId))
	buf.WriteString(`</w:numPr>`)
	return buf.String()
}
