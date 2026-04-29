package cnki

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

type gridPage struct {
	papers    []model.Paper
	total     int
	turnpage  string
	maxPage   int
	rawHTML   string
	pageIndex int
}

func parseGridHTML(body string, client *Client) gridPage {
	gp := gridPage{
		total:    parseTotalHits(body),
		turnpage: attrValue(firstMatch(body, `(?is)<input\b[^>]*\bid\s*=\s*["']hidTurnPage["'][^>]*>`), "value"),
		maxPage:  intOrZero(attrValue(firstMatch(body, `(?is)<span\b[^>]*\bcountPageMark\b[^>]*>`), "data-pagenum")),
		rawHTML:  body,
	}

	for _, tr := range allMatches(body, `(?is)<tr\b[^>]*>(.*?)</tr>`) {
		paper := parseGridRow(tr, client)
		if paper.Title == "" {
			continue
		}
		gp.papers = append(gp.papers, paper)
	}
	return gp
}

func parseGridRow(row string, client *Client) model.Paper {
	nameCell := cellHTML(row, "name")
	href, title := anchorHrefTextByClass(nameCell, "fz14")
	if title == "" {
		href, title = firstAnchorHrefText(nameCell)
	}

	authorCell := cellHTML(row, "author")
	authors := anchorTexts(authorCell)
	if len(authors) == 0 {
		authors = splitPeople(textOnly(authorCell))
	}

	issue := textOnly(cellHTML(row, "date"))
	paperURL := ""
	if href != "" {
		paperURL = client.resolve(href)
	}
	return model.Paper{
		Title:     title,
		URL:       paperURL,
		Authors:   authors,
		Source:    textOnly(cellHTML(row, "source")),
		Year:      firstYear(issue),
		Issue:     issue,
		Cited:     intOrZero(textOnly(cellHTML(row, "quote"))),
		Downloads: intOrZero(textOnly(cellHTML(row, "download"))),
	}
}

func parseTotalHits(body string) int {
	for _, em := range allMatches(body, `(?is)<em\b[^>]*>(.*?)</em>`) {
		if n := intOrZero(textOnly(em)); n > 0 {
			return n
		}
	}
	return 0
}

func cellHTML(row, className string) string {
	pattern := fmt.Sprintf(`(?is)<td\b[^>]*class\s*=\s*["'][^"']*\b%s\b[^"']*["'][^>]*>(.*?)</td>`, regexp.QuoteMeta(className))
	return firstMatch(row, pattern)
}

func anchorHrefTextByClass(fragment, className string) (string, string) {
	pattern := fmt.Sprintf(`(?is)<a\b([^>]*class\s*=\s*["'][^"']*\b%s\b[^"']*["'][^>]*)>(.*?)</a>`, regexp.QuoteMeta(className))
	re := regexp.MustCompile(pattern)
	m := re.FindStringSubmatch(fragment)
	if len(m) == 0 {
		return "", ""
	}
	return attrValue("<a "+m[1]+">", "href"), textOnly(m[2])
}

func firstAnchorHrefText(fragment string) (string, string) {
	re := regexp.MustCompile(`(?is)<a\b([^>]*)>(.*?)</a>`)
	m := re.FindStringSubmatch(fragment)
	if len(m) == 0 {
		return "", ""
	}
	return attrValue("<a "+m[1]+">", "href"), textOnly(m[2])
}

func anchorTexts(fragment string) []string {
	re := regexp.MustCompile(`(?is)<a\b[^>]*>(.*?)</a>`)
	matches := re.FindAllStringSubmatch(fragment, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		t := textOnly(m[1])
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func splitPeople(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ';' || r == '；' || r == ',' || r == '，' || r == '、'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
