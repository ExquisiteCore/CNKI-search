package cnki

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ExquisiteCore/cnki-search/internal/browser"
	"github.com/ExquisiteCore/cnki-search/internal/model"
	"github.com/chromedp/chromedp"
)

// Detail opens a paper's abstract page and extracts full metadata.
// If withRefs is true, also harvests the references list.
func Detail(br *browser.Browser, url string, withRefs bool) (*model.Detail, error) {
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

	var raw string
	if err := chromedp.Run(br.Ctx, chromedp.Evaluate(detailJS, &raw)); err != nil {
		return nil, fmt.Errorf("extract detail: %w", err)
	}
	var d model.Detail
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return nil, fmt.Errorf("parse detail: %w", err)
	}
	d.URL = url

	if withRefs {
		refs, err := extractReferences(br)
		if err != nil {
			return &d, err
		}
		d.References = refs
	}
	return &d, nil
}

// detailJS collects every field we care about on the paper detail page.
// CNKI has several layouts (老版/新版/个人版/机构版); selectors are generous.
const detailJS = `(() => {
  const pickTxt = (sels) => {
    for (const s of sels) { const el = document.querySelector(s); if (el) { const t = (el.innerText || el.textContent || "").trim(); if (t) return t; } }
    return "";
  };
  const pickAll = (sels) => {
    for (const s of sels) { const list = [...document.querySelectorAll(s)]; if (list.length) return list.map(e => (e.textContent || "").trim()).filter(Boolean); }
    return [];
  };
  const intOr0 = (s) => { const n = parseInt((s||"").replace(/[^0-9]/g, ""), 10); return isNaN(n) ? 0 : n; };

  const title = pickTxt(["h1", ".wx-tit h1", ".title h1"]);
  const authors = pickAll([".author a", "h3.author a", ".wx-tit h3:first-of-type a"]);
  const institutions = pickAll([".orgn a", "h3.orgn a", "h3:nth-of-type(2) a"]);
  const abs = pickTxt(["#ChDivSummary", ".abstract-text", ".brief .abs"]);
  const keywords = pickAll([".keywords a", "p.keywords a"]).map(s => s.replace(/[;；,\s]+$/, ""));

  const topTipText = (() => {
    const el = document.querySelector(".top-tip, .wxTitle, .top-space");
    return el ? (el.innerText || "").trim() : "";
  })();

  const matchAfter = (label) => {
    const re = new RegExp(label + "[：:]\\s*([^\\n\\r]+)");
    const m = (topTipText + "\\n" + document.body.innerText).match(re);
    return m ? m[1].trim() : "";
  };
  const doi = matchAfter("DOI");
  const clc = matchAfter("分类号");
  const fund = pickTxt([".funds", ".fund", "p.funds"]);

  const source = pickTxt([".top-tip a:first-child", ".sourinfo a", ".top-space a", ".journal-name"]);
  const issueRaw = pickTxt([".top-tip .year", ".sourinfo .year", ".top-space .year", ".issue"]);
  const yearMatch = issueRaw.match(/(\d{4})/);
  const year = yearMatch ? parseInt(yearMatch[1], 10) : 0;

  const cited = intOr0(pickTxt(["#annotationcount", ".cited", ".num.cited"]));
  const downloads = intOr0(pickTxt(["#downloadcount", ".download", ".num.download"]));

  return JSON.stringify({
    title, authors, institutions, abstract: abs, keywords,
    doi, clc, fund, source, issue: issueRaw, year, cited, downloads
  });
})()`
