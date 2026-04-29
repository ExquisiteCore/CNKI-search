package cnki

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// References fetches a paper abstract page and extracts its reference list.
func (c *Client) References(ctx context.Context, rawURL string) ([]model.Reference, error) {
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
	return parseReferencesHTML(body), nil
}

func parseReferencesHTML(body string) []model.Reference {
	block := firstNonEmptyRaw(body, []string{
		`(?is)<[^>]*\bid\s*=\s*["']CataLogContent["'][^>]*>(.*?)(?:</section>|<footer\b|</body>)`,
		`(?is)<[^>]*class\s*=\s*["'][^"']*ref-list[^"']*["'][^>]*>(.*?)</[^>]+>`,
		`(?is)<[^>]*\bid\s*=\s*["']references["'][^>]*>(.*?)</[^>]+>`,
		`(?is)<[^>]*\bid\s*=\s*["']div_Summary["'][^>]*>(.*?)</[^>]+>`,
	})
	if block == "" {
		return nil
	}

	items := allMatches(block, `(?is)<li\b[^>]*>(.*?)</li>`)
	refs := make([]model.Reference, 0, len(items))
	for _, item := range items {
		text := textOnly(item)
		if text == "" {
			continue
		}
		refs = append(refs, model.Reference{
			Seq:  len(refs) + 1,
			Text: text,
		})
	}
	return refs
}

func firstNonEmptyRaw(body string, patterns []string) string {
	for _, pattern := range patterns {
		m := regexp.MustCompile(pattern).FindStringSubmatch(body)
		if len(m) >= 2 && m[1] != "" {
			return m[1]
		}
	}
	return ""
}
