package cnki

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ExquisiteCore/cnki-search/internal/browser"
	"github.com/ExquisiteCore/cnki-search/internal/model"
	"github.com/chromedp/chromedp"
)

// References navigates to the paper detail page (if not already there) and
// returns the parsed references list.
func References(br *browser.Browser, url string) ([]model.Reference, error) {
	if url == "" {
		return nil, fmt.Errorf("url is empty")
	}
	if err := chromedp.Run(br.Ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
	); err != nil {
		return nil, fmt.Errorf("open detail: %w", err)
	}
	if err := br.HandleCaptcha(); err != nil {
		return nil, ErrCaptcha
	}
	return extractReferences(br)
}

// extractReferences is shared by `cnki refs` and `cnki detail --with-refs`.
// It first tries to expand the 参考文献 tab, then scrapes the list.
func extractReferences(br *browser.Browser) ([]model.Reference, error) {
	const expandJS = `(() => {
  const anchors = [...document.querySelectorAll("a, .tab-title, .tab-item, li")];
  const hit = anchors.find(a => (a.textContent || "").trim().indexOf("参考文献") >= 0);
  if (hit) { hit.click(); return "CLICKED"; }
  return "NONE";
})()`
	var _clicked string
	_ = chromedp.Run(br.Ctx,
		chromedp.Evaluate(expandJS, &_clicked),
		chromedp.Sleep(1500*time.Millisecond),
	)

	const readJS = `(() => {
  const sels = ["#CataLogContent .essayBox li", "#CataLogContent .essayLi", ".ref-list li", "#references li", "#div_Summary li"];
  let nodes = [];
  for (const s of sels) { const n = document.querySelectorAll(s); if (n && n.length) { nodes = [...n]; break; } }
  const out = [];
  nodes.forEach((li, i) => {
    const t = (li.textContent || "").trim().replace(/\s+/g, " ");
    if (t) out.push({ seq: i + 1, text: t });
  });
  return JSON.stringify(out);
})()`

	var raw string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(readJS, &raw)); err != nil {
		return nil, fmt.Errorf("extract refs: %w", err)
	}
	var refs []model.Reference
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return nil, fmt.Errorf("parse refs: %w", err)
	}
	return refs, nil
}
