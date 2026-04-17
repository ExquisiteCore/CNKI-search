package render

import (
	"fmt"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// searchAsMarkdown renders results as a Markdown table plus a summary line.
func searchAsMarkdown(r *model.SearchResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## 知网检索结果：`%s`\n\n", r.Query.Q)
	fmt.Fprintf(&sb, "命中 %d 条，展示前 %d 条（排序：%s）\n\n", r.TotalHits, r.Fetched, r.Query.Sort)
	sb.WriteString("| # | 标题 | 作者 | 来源 | 年份 | 被引 | 下载 |\n")
	sb.WriteString("|---|------|------|------|------|------|------|\n")
	for _, p := range r.Results {
		fmt.Fprintf(&sb, "| %d | [%s](%s) | %s | %s | %s | %s | %s |\n",
			p.Seq,
			mdEscape(p.Title),
			p.URL,
			mdEscape(strings.Join(p.Authors, ", ")),
			mdEscape(p.Source),
			intToStr(p.Year),
			intToStr(p.Cited),
			intToStr(p.Downloads),
		)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func detailAsMarkdown(d *model.Detail) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "### 《%s》\n\n", d.Title)
	if len(d.Authors) > 0 {
		fmt.Fprintf(&sb, "- **作者**：%s\n", strings.Join(d.Authors, ", "))
	}
	if len(d.Institutions) > 0 {
		fmt.Fprintf(&sb, "- **单位**：%s\n", strings.Join(d.Institutions, "; "))
	}
	if d.Source != "" {
		line := d.Source
		if d.Issue != "" {
			line += "  " + d.Issue
		}
		fmt.Fprintf(&sb, "- **来源**：%s\n", line)
	}
	if d.DOI != "" {
		fmt.Fprintf(&sb, "- **DOI**：%s\n", d.DOI)
	}
	if d.CLC != "" {
		fmt.Fprintf(&sb, "- **分类号**：%s\n", d.CLC)
	}
	if d.Fund != "" {
		fmt.Fprintf(&sb, "- **基金**：%s\n", d.Fund)
	}
	if d.Cited != 0 || d.Downloads != 0 {
		fmt.Fprintf(&sb, "- **被引**：%d 次 | **下载**：%d 次\n", d.Cited, d.Downloads)
	}
	if len(d.Keywords) > 0 {
		fmt.Fprintf(&sb, "- **关键词**：%s\n", strings.Join(d.Keywords, "; "))
	}
	if d.Abstract != "" {
		fmt.Fprintf(&sb, "\n**摘要**：%s\n", d.Abstract)
	}
	if len(d.References) > 0 {
		sb.WriteString("\n**参考文献**：\n\n")
		for _, r := range d.References {
			fmt.Fprintf(&sb, "[%d] %s\n", r.Seq, r.Text)
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

func mdEscape(s string) string {
	r := strings.NewReplacer("|", "\\|", "\n", " ")
	return r.Replace(s)
}
