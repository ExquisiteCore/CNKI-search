package cnki

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ExquisiteCore/cnki-search/internal/browser"
	"github.com/ExquisiteCore/cnki-search/internal/model"
	"github.com/chromedp/chromedp"
)

// Search runs the full AdvSearch flow: open page, fill fields, submit,
// paginate, collect rows until q.Size is satisfied.
func Search(br *browser.Browser, q model.Query) (*model.SearchResult, error) {
	ctx := br.Ctx

	if err := chromedp.Run(ctx,
		chromedp.Navigate(URLAdvSearch),
		chromedp.Sleep(1500*time.Millisecond),
	); err != nil {
		return nil, fmt.Errorf("open AdvSearch: %w", err)
	}
	if err := br.HandleCaptcha(); err != nil {
		return nil, ErrCaptcha
	}

	if err := fillAndSubmit(br, q); err != nil {
		return nil, err
	}
	if err := waitResults(br); err != nil {
		return nil, err
	}
	if err := br.HandleCaptcha(); err != nil {
		return nil, ErrCaptcha
	}

	// Apply sort / filters after the results first render, since CNKI's
	// 来源/类型 filters only appear once there is a result set.
	if err := applySort(br, q.Sort); err != nil {
		return nil, err
	}
	if len(q.Sources) > 0 || len(q.Types) > 0 {
		if err := applyFilters(br, q.Types, q.Sources); err != nil {
			return nil, err
		}
	}

	total, _ := readTotalHits(br)
	papers, err := collectPapers(br, q.Size)
	if err != nil {
		return nil, err
	}
	if len(papers) == 0 {
		return nil, ErrEmpty
	}

	return &model.SearchResult{
		Query:     q,
		TotalHits: total,
		Fetched:   len(papers),
		Results:   papers,
	}, nil
}

// fillAndSubmit plants the query into the search box, picks the field,
// sets year range, then clicks search.
func fillAndSubmit(br *browser.Browser, q model.Query) error {
	code := fieldCodeMap[q.Field]
	if code == "" {
		code = "SU"
	}

	script := fmt.Sprintf(`(() => {
  const pick = (list) => { for (const s of list) { const el = document.querySelector(s); if (el) return el; } return null; };
  const sInput = %s;
  const sField = %s;
  const sFrom  = %s;
  const sTo    = %s;

  const input = pick(sInput);
  if (!input) return "NO_INPUT";
  input.value = %s;
  input.dispatchEvent(new Event("input", {bubbles:true}));
  input.dispatchEvent(new Event("change", {bubbles:true}));

  const sel = pick(sField);
  if (sel && sel.tagName === "SELECT") {
    sel.value = %s;
    sel.dispatchEvent(new Event("change", {bubbles:true}));
  }

  const from = pick(sFrom);
  if (from) { from.value = %s; from.dispatchEvent(new Event("input", {bubbles:true})); }
  const to = pick(sTo);
  if (to) { to.value = %s; to.dispatchEvent(new Event("input", {bubbles:true})); }

  return "OK";
})()`,
		jsList(selSearchInput),
		jsList(selFieldDropdown),
		jsList(selFromYear),
		jsList(selToYear),
		jsString(q.Q),
		jsString(code),
		jsYear(q.From),
		jsYear(q.To),
	)

	var out string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &out)); err != nil {
		return fmt.Errorf("fill search form: %w", err)
	}
	if out == "NO_INPUT" {
		return fmt.Errorf("could not locate search input (selectors may have changed)")
	}

	// Small natural pause before clicking.
	if err := chromedp.Run(br.Ctx, chromedp.Sleep(900*time.Millisecond)); err != nil {
		return err
	}

	clickScript := fmt.Sprintf(`(() => {
  const list = %s;
  for (const s of list) {
    const el = document.querySelector(s);
    if (el) { el.click(); return "CLICKED"; }
  }
  return "NO_BTN";
})()`, jsList(selSubmitBtn))

	var clicked string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(clickScript, &clicked)); err != nil {
		return fmt.Errorf("click submit: %w", err)
	}
	if clicked == "NO_BTN" {
		return fmt.Errorf("could not locate submit button")
	}
	return nil
}

// waitResults polls for the result table. Returns when at least one row shows
// up, or after 5 attempts spaced ~2s apart.
func waitResults(br *browser.Browser) error {
	script := fmt.Sprintf(`(() => {
  const list = %s;
  for (const s of list) {
    const rows = document.querySelectorAll(s);
    if (rows && rows.length > 0) return rows.length;
  }
  return 0;
})()`, jsList(selResultRow))

	for i := 0; i < 8; i++ {
		var n int
		if err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &n)); err == nil && n > 0 {
			return nil
		}
		if err := chromedp.Run(br.Ctx, chromedp.Sleep(1500*time.Millisecond)); err != nil {
			return err
		}
	}
	return fmt.Errorf("results table did not appear in time")
}

func applySort(br *browser.Browser, sort string) error {
	code, ok := sortCodeMap[sort]
	if !ok || code == "" {
		return nil
	}
	// CNKI exposes sort options as visible tabs; click by label.
	label := map[string]string{
		"PT,down":  "发表时间",
		"CF,down":  "被引",
		"DFR,down": "下载",
	}[code]
	if label == "" {
		return nil
	}
	script := fmt.Sprintf(`(() => {
  const anchors = [...document.querySelectorAll(".sort-list a, .sort-box a, .pagerTitleCell a, .group-item a")];
  const hit = anchors.find(a => a.textContent && a.textContent.trim().indexOf(%s) >= 0);
  if (hit) { hit.click(); return "OK"; }
  return "MISS";
})()`, jsString(label))
	var out string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &out), chromedp.Sleep(1500*time.Millisecond)); err != nil {
		return fmt.Errorf("apply sort: %w", err)
	}
	return nil
}

func applyFilters(br *browser.Browser, types, sources []string) error {
	labels := []string{}
	for _, t := range types {
		if l, ok := typeFilterMap[t]; ok {
			labels = append(labels, l)
		}
	}
	for _, s := range sources {
		if l, ok := sourceFilterMap[s]; ok {
			labels = append(labels, l)
		}
	}
	if len(labels) == 0 {
		return nil
	}
	blob, _ := json.Marshal(labels)
	script := fmt.Sprintf(`(() => {
  const want = new Set(%s);
  const nodes = [...document.querySelectorAll(".group-item .item, .filterGroup li, .filterBox li, aside a")];
  const hits = [];
  for (const n of nodes) {
    const t = (n.textContent || "").trim();
    for (const w of want) {
      if (t.indexOf(w) >= 0) { n.click(); hits.push(w); break; }
    }
  }
  return hits.join(",");
})()`, string(blob))
	var out string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &out), chromedp.Sleep(1500*time.Millisecond)); err != nil {
		return fmt.Errorf("apply filters: %w", err)
	}
	return nil
}

func readTotalHits(br *browser.Browser) (int, error) {
	script := fmt.Sprintf(`(() => {
  const list = %s;
  for (const s of list) {
    const el = document.querySelector(s);
    if (el) {
      const m = (el.textContent || "").replace(/[^0-9]/g, "");
      if (m) return parseInt(m, 10);
    }
  }
  return 0;
})()`, jsList(selTotalHits))
	var n int
	err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &n))
	return n, err
}

// collectPapers extracts rows across pages until `size` is satisfied or pages
// run out.
func collectPapers(br *browser.Browser, size int) ([]model.Paper, error) {
	maxPages := size/20 + 3
	collected := make([]model.Paper, 0, size)
	seen := map[string]bool{}

	for page := 0; page < maxPages && len(collected) < size; page++ {
		rows, err := extractRows(br)
		if err != nil {
			return collected, err
		}
		for _, r := range rows {
			key := r.URL
			if key == "" {
				key = r.Title
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			r.Seq = len(collected) + 1
			collected = append(collected, r)
			if len(collected) >= size {
				break
			}
		}
		if len(collected) >= size {
			break
		}
		more, err := gotoNextPage(br)
		if err != nil {
			return collected, err
		}
		if !more {
			break
		}
	}
	return collected, nil
}

// extractRows runs a JSON-emitting script in the page and unmarshals it.
func extractRows(br *browser.Browser) ([]model.Paper, error) {
	script := fmt.Sprintf(`(() => {
  const pickOne = (el, sels) => { for (const s of sels) { const m = el.querySelector(s); if (m) return m; } return null; };
  const rowSels = %s;
  let rows = [];
  for (const s of rowSels) { const r = document.querySelectorAll(s); if (r && r.length) { rows = [...r]; break; } }
  const out = [];
  for (const row of rows) {
    const titleA = pickOne(row, ["td.name a.fz14", ".name a", "td.name a", "a.fz14"]);
    if (!titleA) continue;
    const authorCell = pickOne(row, ["td.author", ".author"]);
    const sourceCell = pickOne(row, ["td.source a", ".source a", "td.source", ".source"]);
    const dateCell   = pickOne(row, ["td.date", ".date"]);
    const citeCell   = pickOne(row, ["td.quote", ".quote", "td.cited", ".cited"]);
    const dlCell     = pickOne(row, ["td.download", ".download"]);

    const authors = authorCell ? [...authorCell.querySelectorAll("a")].map(a => a.textContent.trim()).filter(Boolean) : [];
    if (authorCell && authors.length === 0) {
      const t = (authorCell.textContent || "").trim();
      if (t) t.split(/[;,，、]\s*/).forEach(x => authors.push(x));
    }
    const dateRaw = dateCell ? (dateCell.textContent || "").trim() : "";
    const year = (dateRaw.match(/(\d{4})/) || [])[1] || "";
    const intOr0 = (s) => { const n = parseInt((s||"").replace(/[^0-9]/g, ""), 10); return isNaN(n) ? 0 : n; };

    out.push({
      title: (titleA.textContent || "").trim(),
      url: titleA.href || "",
      authors: authors,
      source: sourceCell ? (sourceCell.textContent || "").trim() : "",
      year: year ? parseInt(year, 10) : 0,
      issue: dateRaw,
      cited: intOr0(citeCell && citeCell.textContent),
      downloads: intOr0(dlCell && dlCell.textContent),
    });
  }
  return JSON.stringify(out);
})()`, jsList(selResultRow))

	var raw string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &raw)); err != nil {
		return nil, fmt.Errorf("extract rows: %w", err)
	}
	var rows []model.Paper
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, fmt.Errorf("parse rows: %w", err)
	}
	return rows, nil
}

func gotoNextPage(br *browser.Browser) (bool, error) {
	script := fmt.Sprintf(`(() => {
  const list = %s;
  for (const s of list) {
    const el = document.querySelector(s);
    if (el && !el.classList.contains("disabled") && el.offsetParent !== null) { el.click(); return "CLICKED"; }
  }
  return "END";
})()`, jsList(selNextPage))

	var out string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(script, &out)); err != nil {
		return false, err
	}
	if out != "CLICKED" {
		return false, nil
	}
	// AJAX pagination: wait for DOM to settle.
	if err := chromedp.Run(br.Ctx, chromedp.Sleep(2*time.Second)); err != nil {
		return false, err
	}
	return true, nil
}

// OpenHome is used by `cnki login` to park the browser on the landing page.
func OpenHome(br *browser.Browser) error {
	return chromedp.Run(br.Ctx,
		chromedp.Navigate(URLHome),
		chromedp.Sleep(2*time.Second),
	)
}

// -- small JS helpers --

// jsString emits a JSON-encoded string so it is safe to splice into a JS
// literal as a quoted value.
func jsString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func jsYear(y int) string {
	if y <= 0 {
		return `""`
	}
	return fmt.Sprintf(`"%d"`, y)
}

// jsList turns a Go []string into a JS array literal of strings.
func jsList(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(ss))
	for _, s := range ss {
		parts = append(parts, jsString(s))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
