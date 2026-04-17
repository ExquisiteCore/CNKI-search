package render

import (
	"fmt"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// searchAsCitation formats each paper per GB/T 7714-2015.
// Format for a journal article: [seq] 作者. 题名[J]. 刊名, 年, 卷(期): 页码.
// We cannot always recover volume/pages from the search row, so we include
// whatever is available.
func searchAsCitation(r *model.SearchResult) string {
	var sb strings.Builder
	for _, p := range r.Results {
		sb.WriteString(formatCitation(p))
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formatCitation(p model.Paper) string {
	authors := joinAuthors(p.Authors)
	title := p.Title
	source := p.Source
	year := ""
	if p.Year > 0 {
		year = fmt.Sprintf("%d", p.Year)
	}

	var parts []string
	if authors != "" {
		parts = append(parts, authors+".")
	}
	if title != "" {
		parts = append(parts, title+"[J].")
	}
	tail := []string{}
	if source != "" {
		tail = append(tail, source)
	}
	if year != "" {
		tail = append(tail, year)
	}
	if p.Issue != "" && p.Issue != year {
		tail = append(tail, p.Issue)
	}
	if len(tail) > 0 {
		parts = append(parts, strings.Join(tail, ", ")+".")
	}
	return fmt.Sprintf("[%d] %s", p.Seq, strings.Join(parts, " "))
}

// joinAuthors renders up to 3 authors joined by ", "; if more, appends "等".
func joinAuthors(as []string) string {
	switch {
	case len(as) == 0:
		return ""
	case len(as) <= 3:
		return strings.Join(as, ", ")
	default:
		return strings.Join(as[:3], ", ") + ", 等"
	}
}

// detailAsCitation formats a single Detail the same way as a row.
func detailAsCitation(d *model.Detail) string {
	p := model.Paper{
		Seq:     1,
		Title:   d.Title,
		Authors: d.Authors,
		Source:  d.Source,
		Year:    d.Year,
		Issue:   d.Issue,
	}
	return formatCitation(p)
}
