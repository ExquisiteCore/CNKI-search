package render

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// searchAsTable renders a fixed-column human-readable table.
// ASCII only on column borders; Chinese content is fine.
func searchAsTable(r *model.SearchResult) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# 共 %d 条命中，已抓取 %d 条（query=%q，field=%s，sort=%s）\n\n",
		r.TotalHits, r.Fetched, r.Query.Q, r.Query.Field, r.Query.Sort)

	headers := []string{"#", "标题", "作者", "来源", "年份", "被引", "下载"}
	rows := make([][]string, 0, len(r.Results))
	for _, p := range r.Results {
		rows = append(rows, []string{
			fmt.Sprintf("%d", p.Seq),
			truncate(p.Title, 40),
			truncate(strings.Join(p.Authors, ", "), 20),
			truncate(p.Source, 18),
			intToStr(p.Year),
			intToStr(p.Cited),
			intToStr(p.Downloads),
		})
	}

	widths := columnWidths(headers, rows)
	renderRow(&sb, headers, widths)
	renderSep(&sb, widths)
	for _, row := range rows {
		renderRow(&sb, row, widths)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func columnWidths(headers []string, rows [][]string) []int {
	ws := make([]int, len(headers))
	for i, h := range headers {
		ws[i] = displayWidth(h)
	}
	for _, r := range rows {
		for i, c := range r {
			if w := displayWidth(c); w > ws[i] {
				ws[i] = w
			}
		}
	}
	return ws
}

func renderRow(sb *strings.Builder, cells []string, widths []int) {
	sb.WriteString("| ")
	for i, c := range cells {
		sb.WriteString(c)
		pad := widths[i] - displayWidth(c)
		if pad > 0 {
			sb.WriteString(strings.Repeat(" ", pad))
		}
		sb.WriteString(" | ")
	}
	sb.WriteString("\n")
}

func renderSep(sb *strings.Builder, widths []int) {
	sb.WriteString("|")
	for _, w := range widths {
		sb.WriteString(strings.Repeat("-", w+2))
		sb.WriteString("|")
	}
	sb.WriteString("\n")
}

// displayWidth approximates a monospaced cell width: CJK chars count as 2.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x1100 && (r <= 0x115f ||
			(r >= 0x2e80 && r <= 0x303e) ||
			(r >= 0x3041 && r <= 0x33ff) ||
			(r >= 0x3400 && r <= 0x4dbf) ||
			(r >= 0x4e00 && r <= 0x9fff) ||
			(r >= 0xa000 && r <= 0xa4cf) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe30 && r <= 0xfe4f) ||
			(r >= 0xff00 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6)) {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func truncate(s string, max int) string {
	if displayWidth(s) <= max {
		return s
	}
	w := 0
	var b strings.Builder
	for _, r := range s {
		rw := 1
		if utf8.RuneLen(r) > 1 {
			rw = 2
		}
		if w+rw > max-1 {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String()
}

func intToStr(n int) string {
	if n == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", n)
}
