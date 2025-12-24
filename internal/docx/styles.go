package docx

import (
	"bytes"
	"fmt"

	"md2word/internal/config"
)

// GenerateStyles 生成样式XML
func GenerateStyles(cfg *config.Config) string {
	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:docDefaults>
        <w:rPrDefault>
            <w:rPr>
                <w:rFonts w:ascii="` + cfg.Styles.Body.Font + `" w:eastAsia="` + cfg.Styles.Body.Font + `" w:hAnsi="` + cfg.Styles.Body.Font + `"/>
                <w:sz w:val="` + fmt.Sprintf("%d", int(cfg.Styles.Body.Size*2)) + `"/>
                <w:szCs w:val="` + fmt.Sprintf("%d", int(cfg.Styles.Body.Size*2)) + `"/>
            </w:rPr>
        </w:rPrDefault>
        <w:pPrDefault>
            <w:pPr>
                <w:spacing w:after="0" w:line="276" w:lineRule="auto"/>
            </w:pPr>
        </w:pPrDefault>
    </w:docDefaults>`)

	// Normal样式
	buf.WriteString(`
    <w:style w:type="paragraph" w:default="1" w:styleId="Normal">
        <w:name w:val="Normal"/>
        <w:rPr>
            <w:rFonts w:ascii="` + cfg.Styles.Body.Font + `" w:eastAsia="` + cfg.Styles.Body.Font + `" w:hAnsi="` + cfg.Styles.Body.Font + `"/>
            <w:sz w:val="` + fmt.Sprintf("%d", int(cfg.Styles.Body.Size*2)) + `"/>
            <w:szCs w:val="` + fmt.Sprintf("%d", int(cfg.Styles.Body.Size*2)) + `"/>
        </w:rPr>
    </w:style>`)

	// 各级标题样式
	for level := 1; level <= 9; level++ {
		style := cfg.GetHeadingStyle(level)
		styleID := fmt.Sprintf("Heading%d", level)
		outlineLvl := level - 1

		buf.WriteString(`
    <w:style w:type="paragraph" w:styleId="` + styleID + `">
        <w:name w:val="heading ` + fmt.Sprintf("%d", level) + `"/>
        <w:basedOn w:val="Normal"/>
        <w:next w:val="Normal"/>
        <w:pPr>
            <w:keepNext/>
            <w:keepLines/>
            <w:spacing w:before="240" w:after="120"/>
            <w:outlineLvl w:val="` + fmt.Sprintf("%d", outlineLvl) + `"/>
        </w:pPr>
        <w:rPr>
            <w:rFonts w:ascii="` + style.Font + `" w:eastAsia="` + style.Font + `" w:hAnsi="` + style.Font + `"/>
            <w:sz w:val="` + fmt.Sprintf("%d", int(style.Size*2)) + `"/>
            <w:szCs w:val="` + fmt.Sprintf("%d", int(style.Size*2)) + `"/>`)

		if style.Bold {
			buf.WriteString(`
            <w:b/>
            <w:bCs/>`)
		}

		buf.WriteString(`
        </w:rPr>
    </w:style>`)
	}

	// 代码样式
	buf.WriteString(`
    <w:style w:type="paragraph" w:styleId="Code">
        <w:name w:val="Code"/>
        <w:basedOn w:val="Normal"/>
        <w:pPr>
            <w:shd w:val="clear" w:color="auto" w:fill="F5F5F5"/>
            <w:spacing w:before="120" w:after="120"/>
        </w:pPr>
        <w:rPr>
            <w:rFonts w:ascii="Consolas" w:hAnsi="Consolas" w:cs="Consolas"/>
            <w:sz w:val="` + fmt.Sprintf("%d", int(cfg.Styles.Code.Size*2)) + `"/>
            <w:szCs w:val="` + fmt.Sprintf("%d", int(cfg.Styles.Code.Size*2)) + `"/>
        </w:rPr>
    </w:style>`)

	// 表格样式
	buf.WriteString(`
    <w:style w:type="table" w:styleId="TableGrid">
        <w:name w:val="Table Grid"/>
        <w:basedOn w:val="TableNormal"/>
        <w:tblPr>
            <w:tblBorders>
                <w:top w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                <w:left w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                <w:bottom w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                <w:right w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                <w:insideH w:val="single" w:sz="4" w:space="0" w:color="auto"/>
                <w:insideV w:val="single" w:sz="4" w:space="0" w:color="auto"/>
            </w:tblBorders>
        </w:tblPr>
    </w:style>`)

	buf.WriteString(`
</w:styles>`)

	return buf.String()
}

// FontSizeToTwips 将磅值转换为Twips (1pt = 2 half-points)
func FontSizeToTwips(pt float64) int {
	return int(pt * 2)
}

// ChineseFontSizeMap 中文字号到磅值的映射
var ChineseFontSizeMap = map[string]float64{
	"初号": 42,
	"小初": 36,
	"一号": 26,
	"小一": 24,
	"二号": 22,
	"小二": 18,
	"三号": 16,
	"小三": 15,
	"四号": 14,
	"小四": 12,
	"五号": 10.5,
	"小五": 9,
	"六号": 7.5,
	"小六": 6.5,
	"七号": 5.5,
	"八号": 5,
}

// ParseChineseFontSize 解析中文字号
func ParseChineseFontSize(size string) (float64, bool) {
	pt, ok := ChineseFontSizeMap[size]
	return pt, ok
}
