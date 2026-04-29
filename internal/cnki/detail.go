package cnki

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// Detail fetches a paper abstract page over HTTP and extracts metadata.
func (c *Client) Detail(ctx context.Context, rawURL string, withRefs bool) (*model.Detail, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("url is empty")
	}
	if err := c.ensureClientID(ctx); err != nil {
		return nil, fmt.Errorf("prepare client id: %w", err)
	}
	body, err := c.fetchHTML(ctx, rawURL, URLKNSBase)
	if err != nil {
		return nil, fmt.Errorf("open detail: %w", err)
	}

	d := parseDetailHTML(body)
	d.URL = rawURL
	if withRefs {
		d.References = parseReferencesHTML(body)
	}
	return &d, nil
}

func (c *Client) fetchHTML(ctx context.Context, rawURL, referer string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", referer)
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	return c.doText(req)
}

func parseDetailHTML(body string) model.Detail {
	topTip := textOnly(firstMatch(body, `(?is)<(?:div|p|span)\b[^>]*class\s*=\s*["'][^"']*(?:top-tip|wxTitle|top-space|sourinfo)[^"']*["'][^>]*>(.*?)</(?:div|p|span)>`))
	bodyText := textOnly(body)
	source := firstNonEmptyText(body, []string{
		`(?is)<[^>]*class\s*=\s*["'][^"']*(?:top-tip|sourinfo|top-space)[^"']*["'][^>]*>.*?<a\b[^>]*>(.*?)</a>`,
		`(?is)<[^>]*class\s*=\s*["'][^"']*journal-name[^"']*["'][^>]*>(.*?)</[^>]+>`,
	})
	if source == "" {
		source = sourceFromTopTip(topTip)
	}
	issue := firstNonEmptyText(body, []string{
		`(?is)<[^>]*class\s*=\s*["'][^"']*(?:top-tip|sourinfo|top-space)[^"']*["'][^>]*>.*?<[^>]*class\s*=\s*["'][^"']*year[^"']*["'][^>]*>(.*?)</[^>]+>`,
		`(?is)<[^>]*class\s*=\s*["'][^"']*\bissue\b[^"']*["'][^>]*>(.*?)</[^>]+>`,
	})
	if issue == "" {
		issue = issueFromTopTip(topTip)
	}

	return model.Detail{
		Title:        firstNonEmptyText(body, []string{`(?is)<h1\b[^>]*>(.*?)</h1>`, `(?is)<title\b[^>]*>(.*?)</title>`}),
		Authors:      anchorsFromClass(body, "author"),
		Institutions: anchorsFromClass(body, "orgn"),
		Abstract:     firstNonEmptyText(body, []string{`(?is)<[^>]*\bid\s*=\s*["']ChDivSummary["'][^>]*>(.*?)</[^>]+>`, `(?is)<[^>]*class\s*=\s*["'][^"']*abstract-text[^"']*["'][^>]*>(.*?)</[^>]+>`}),
		Keywords:     cleanKeywords(anchorsFromClass(body, "keywords")),
		DOI:          matchAfter(bodyText, "DOI"),
		CLC:          matchAfter(bodyText, "分类号"),
		Source:       source,
		Issue:        issue,
		Year:         firstYear(issue),
		Fund:         firstNonEmptyText(body, []string{`(?is)<[^>]*class\s*=\s*["'][^"']*\bfunds?\b[^"']*["'][^>]*>(.*?)</[^>]+>`}),
		Cited:        intOrZero(firstNonEmptyText(body, []string{`(?is)<[^>]*\bid\s*=\s*["']annotationcount["'][^>]*>(.*?)</[^>]+>`, `(?is)<[^>]*class\s*=\s*["'][^"']*\bcited\b[^"']*["'][^>]*>(.*?)</[^>]+>`})),
		Downloads:    intOrZero(firstNonEmptyText(body, []string{`(?is)<[^>]*\bid\s*=\s*["']downloadcount["'][^>]*>(.*?)</[^>]+>`, `(?is)<[^>]*class\s*=\s*["'][^"']*\bdownload\b[^"']*["'][^>]*>(.*?)</[^>]+>`})),
	}
}

func firstNonEmptyText(body string, patterns []string) string {
	for _, pattern := range patterns {
		if t := textOnly(firstMatch(body, pattern)); t != "" {
			t = strings.TrimSuffix(t, " - 中国知网")
			if t != "" {
				return t
			}
		}
	}
	return ""
}

func anchorsFromClass(body, className string) []string {
	block := blockByClass(body, className)
	if block == "" {
		return nil
	}
	return anchorTexts(block)
}

func blockByClass(body, className string) string {
	for _, tag := range []string{"div", "p", "h3", "section", "span"} {
		pattern := fmt.Sprintf(`(?is)<%s\b[^>]*class\s*=\s*["'][^"']*\b%s\b[^"']*["'][^>]*>(.*?)</%s>`, tag, regexp.QuoteMeta(className), tag)
		if block := firstMatch(body, pattern); block != "" {
			return block
		}
	}
	return ""
}

func cleanKeywords(in []string) []string {
	out := make([]string, 0, len(in))
	for _, keyword := range in {
		keyword = strings.Trim(keyword, " ;；,，、")
		if keyword != "" {
			out = append(out, keyword)
		}
	}
	return out
}

func matchAfter(text, label string) string {
	pattern := fmt.Sprintf(`(?is)%s\s*[：:]\s*([^\n\r]+)`, regexp.QuoteMeta(label))
	m := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if len(m) < 2 {
		return ""
	}
	value := strings.TrimSpace(m[1])
	for _, next := range []string{" DOI", " 分类号", " 基金", " 来源", " 被引", " 下载"} {
		if idx := strings.Index(value, next); idx >= 0 {
			value = strings.TrimSpace(value[:idx])
		}
	}
	if label == "分类号" || label == "DOI" {
		if fields := strings.Fields(value); len(fields) > 0 {
			value = fields[0]
		}
	}
	return strings.Trim(value, " ;；,，")
}

func sourceFromTopTip(topTip string) string {
	if topTip == "" {
		return ""
	}
	yearRE := regexp.MustCompile(`\b(19|20|21)\d{2}\b`)
	loc := yearRE.FindStringIndex(topTip)
	if loc == nil {
		return topTip
	}
	return strings.TrimSpace(topTip[:loc[0]])
}

func issueFromTopTip(topTip string) string {
	if topTip == "" {
		return ""
	}
	yearRE := regexp.MustCompile(`\b(19|20|21)\d{2}[^\s]*`)
	if m := yearRE.FindString(topTip); m != "" {
		return strings.TrimSpace(m)
	}
	return ""
}
