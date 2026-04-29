# CNKI Reference CLI HTTP Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `cnki` a focused CLI for finding CNKI papers for reference lists, while preserving direct GB/T 7714-style citation export.

**Architecture:** Keep `internal/cnki.Client` as the HTTP data source and `internal/render` as the output contract. Split the current large HTTP implementation into request construction, grid parsing, detail/reference parsing, and shared HTML utilities. Treat citation output as a first-class regression-tested feature, not a secondary renderer.

**Tech Stack:** Go 1.26, `net/http`, `httptest`, `cobra`, existing `internal/model` and `internal/render`; no browser automation and no new dependencies unless a task provides failing evidence.

---

## Product Positioning

- This is a reference-finding CLI, not a full-text downloader.
- Core workflow: `cnki search "topic" --size=N --format=citation`.
- Supporting workflows: `cnki search ... --format=json|table|markdown`, `cnki detail <url> --format=citation`, `cnki detail <url> --with-refs`, `cnki refs <url>`.
- The direct citation export path must not regress: `Search(..., "citation")`, `Detail(..., "citation")`, and CLI `--format=citation` stay supported.
- Out of scope for this phase: account login, Chrome/profile automation, NetEase OAuth, full-text download, CAJ/PDF saving, reader jump URL management, or bypassing authorization.

## Evidence From `/cnki分析`

- `cnki_search.py` supports the search direction: the core HTTP endpoint is `POST https://kns.cnki.net/kns8s/brief/grid`, with `QueryJson`, `hidTurnPage`, and `SearchFrom=1` for first page / `SearchFrom=4` for following pages.
- `cnki_search.py` uses `curl_cffi.Session(impersonate="chrome136")`; keep Go stdlib HTTP for now, but add diagnostics before adding TLS fingerprint dependencies.
- `cnki_reader.py` confirms abstract/detail URLs from search results are the right source for metadata and references. Do not reconstruct detail URLs by hand.
- `cnki_batch_download.py` and `netease_login.py` are useful background, but they are not part of the reference CLI core. They show separate authorization/download complexity and should not drive this refactor unless the product scope changes.

## Target File Structure

- `internal/cnki/client.go`: HTTP client construction, request helpers, response handling, captcha detection.
- `internal/cnki/search.go`: public `Client.Search` orchestration only.
- `internal/cnki/query.go`: field/type/sort maps and `buildQueryJSON`.
- `internal/cnki/grid_parser.go`: `parseGridHTML`, `parseGridRow`, total/turnpage/max-page parsing.
- `internal/cnki/detail.go`: public `Client.Detail` and detail metadata parser.
- `internal/cnki/refs.go`: public `Client.References` and reference parser.
- `internal/cnki/htmlutil.go`: shared HTML/text helpers, URL attrs, integer/year parsing.
- `internal/cnki/http_client_test.go`: integration-style HTTP client tests.
- `internal/cnki/query_test.go`: QueryJson/form unit tests.
- `internal/cnki/grid_parser_test.go`: grid parser unit tests.
- `internal/cnki/detail_parser_test.go`: detail/reference parser unit tests.
- `internal/render/citation_test.go`: citation export contract tests.
- `README.md`, `INSTALL.md`, `skills/cnki-search/SKILL.md`: describe the reference-finding positioning and citation export as core.

---

### Task 1: Lock Citation Export Contract

**Files:**
- Create or keep: `internal/render/citation_test.go`
- Modify only if test fails: `internal/render/citation.go`

- [ ] **Step 1: Add the citation regression test**

Create `internal/render/citation_test.go`:

```go
package render

import (
	"testing"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

func TestSearchCitationFormatIsPreserved(t *testing.T) {
	t.Parallel()

	got, err := Search(&model.SearchResult{
		Results: []model.Paper{
			{
				Seq:     1,
				Title:   "基于知识图谱的参考文献检索研究",
				Authors: []string{"张三", "李四", "王五", "赵六"},
				Source:  "情报学报",
				Year:    2024,
				Issue:   "2024年第3期",
			},
		},
	}, "citation")
	if err != nil {
		t.Fatal(err)
	}

	want := "[1] 张三, 李四, 王五, 等. 基于知识图谱的参考文献检索研究[J]. 情报学报, 2024, 2024年第3期."
	if got != want {
		t.Fatalf("citation output changed:\nwant: %s\n got: %s", want, got)
	}
}

func TestDetailCitationFormatIsPreserved(t *testing.T) {
	t.Parallel()

	got, err := Detail(&model.Detail{
		Title:   "大语言模型文献综述",
		Authors: []string{"作者甲"},
		Source:  "计算机科学",
		Year:    2025,
	}, "citation")
	if err != nil {
		t.Fatal(err)
	}

	want := "[1] 作者甲. 大语言模型文献综述[J]. 计算机科学, 2025."
	if got != want {
		t.Fatalf("detail citation output changed:\nwant: %s\n got: %s", want, got)
	}
}
```

- [ ] **Step 2: Run the focused test**

Run: `go test ./internal/render -run Citation -count=1`

Expected: PASS. If it fails, update `citation.go` to preserve the expected direct citation output.

---

### Task 2: Split Search Request Construction

**Files:**
- Create: `internal/cnki/query.go`
- Create: `internal/cnki/query_test.go`
- Modify: `internal/cnki/search.go`
- Modify: `internal/cnki/fields.go`

- [ ] **Step 1: Add QueryJson characterization test**

Create `internal/cnki/query_test.go`:

```go
package cnki

import (
	"encoding/json"
	"testing"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

func TestBuildQueryJSONIncludesKeywordYearTypeAndSearchFrom(t *testing.T) {
	t.Parallel()

	raw, err := buildQueryJSON(model.Query{
		Q:     "大语言模型",
		From:  2020,
		To:    2025,
		Types: []string{"journal", "conference"},
	}, "SU", "YSTT4HG0,JUP3MUPD", 4)
	if err != nil {
		t.Fatal(err)
	}

	var got struct {
		Resource   string `json:"Resource"`
		Classid    string `json:"Classid"`
		KuaKuCodes string `json:"KuaKuCodes"`
		SearchFrom int    `json:"SearchFrom"`
		Query      struct {
			QGroup []struct {
				Groups []struct {
					Items []struct {
						Field  string `json:"Field"`
						Value  string `json:"Value"`
						Value2 string `json:"Value2"`
					} `json:"Items"`
				} `json:"Groups"`
			} `json:"QGroup"`
		} `json:"Query"`
	}
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}

	if got.Resource != "CROSSDB" || got.Classid != "WD0FTY92" {
		t.Fatalf("scope = %q/%q", got.Resource, got.Classid)
	}
	if got.KuaKuCodes != "YSTT4HG0,JUP3MUPD" || got.SearchFrom != 4 {
		t.Fatalf("crossids/searchFrom = %q/%d", got.KuaKuCodes, got.SearchFrom)
	}
	item := got.Query.QGroup[0].Groups[0].Items[0]
	if item.Field != "SU" || item.Value != "大语言模型" {
		t.Fatalf("keyword item = %#v", item)
	}
	year := got.Query.QGroup[1].Groups[0].Items[0]
	if year.Field != "PT" || year.Value != "2020-01-01" || year.Value2 != "2025-12-31" {
		t.Fatalf("year item = %#v", year)
	}
}
```

- [ ] **Step 2: Run characterization test**

Run: `go test ./internal/cnki -run TestBuildQueryJSONIncludesKeywordYearTypeAndSearchFrom -count=1`

Expected: PASS before moving code.

- [ ] **Step 3: Move request construction**

Move from `internal/cnki/search.go` to `internal/cnki/query.go`:

```go
const (
	defaultClassID    = "WD0FTY92"
	defaultCrossIDs   = "YSTT4HG0,LSTPFY1C,JUP3MUPD,MPMFIG1A,EMRPGLPA,WQ0UVIAA,BLZOG7CK,PWFIRAGL,NN3FJMUV,NLBO1Z6R"
	defaultProductStr = defaultCrossIDs
	defaultPageSize   = 20
)
```

Also move `fieldTitleMap`, `typeCrossIDMap`, `sortSpec`, `httpSortCodeMap`, `buildQueryJSON`, `crossIDsForTypes`, and `valueIf`.

- [ ] **Step 4: Run package tests**

Run: `go test ./internal/cnki -count=1`

Expected: PASS.

---

### Task 3: Split Grid HTML Parsing

**Files:**
- Create: `internal/cnki/grid_parser.go`
- Create: `internal/cnki/grid_parser_test.go`
- Modify: `internal/cnki/search.go`

- [ ] **Step 1: Add grid parser characterization test**

Create `internal/cnki/grid_parser_test.go`:

```go
package cnki

import "testing"

func TestParseGridHTMLReadsRowsAndPaginationTokens(t *testing.T) {
	t.Parallel()

	client := NewClient(ClientOptions{BaseURL: "https://kns.cnki.net"})
	html := `<span>找到<em>2,345</em>条</span>
<input id="hidTurnPage" value="tp-abc">
<span class="countPageMark" data-pagenum="118"></span>
<table><tbody>
<tr>
  <td class="seq">7</td>
  <td class="name"><a class="fz14" href="/kcms2/article/abstract?v=abc">测试题名</a></td>
  <td class="author"><a class="KnowledgeNetLink">张三</a><a class="KnowledgeNetLink">李四</a></td>
  <td class="source"><a>软件学报</a></td>
  <td class="date">2024-03</td>
  <td class="quote">12</td>
  <td class="download">34</td>
</tr>
</tbody></table>`

	page := parseGridHTML(html, client)

	if page.total != 2345 || page.turnpage != "tp-abc" || page.maxPage != 118 {
		t.Fatalf("page metadata = %#v", page)
	}
	if len(page.papers) != 1 {
		t.Fatalf("papers len = %d", len(page.papers))
	}
	p := page.papers[0]
	if p.Title != "测试题名" || p.URL != "https://kns.cnki.net/kcms2/article/abstract?v=abc" {
		t.Fatalf("paper title/url = %#v", p)
	}
	if p.Year != 2024 || p.Cited != 12 || p.Downloads != 34 {
		t.Fatalf("paper metrics = %#v", p)
	}
}
```

- [ ] **Step 2: Run characterization test**

Run: `go test ./internal/cnki -run TestParseGridHTMLReadsRowsAndPaginationTokens -count=1`

Expected: PASS before moving code.

- [ ] **Step 3: Move grid parser code**

Move from `search.go` to `grid_parser.go`: `gridPage`, `parseGridHTML`, `parseGridRow`, `parseTotalHits`, `cellHTML`, `anchorHrefTextByClass`, `firstAnchorHrefText`, `anchorTexts`, `splitPeople`.

- [ ] **Step 4: Run focused tests**

Run: `go test ./internal/cnki -run 'Test(ParseGridHTML|HTTPClientSearch)' -count=1`

Expected: PASS.

---

### Task 4: Centralize HTML/Text Helpers

**Files:**
- Create: `internal/cnki/htmlutil.go`
- Create: `internal/cnki/htmlutil_test.go`
- Modify: `internal/cnki/search.go`
- Modify: `internal/cnki/detail.go`
- Modify: `internal/cnki/refs.go`

- [ ] **Step 1: Add helper characterization test**

Create `internal/cnki/htmlutil_test.go`:

```go
package cnki

import "testing"

func TestHTMLTextHelpers(t *testing.T) {
	t.Parallel()

	if got := textOnly(`<p> A&nbsp;<b>中文</b>  B </p>`); got != "A 中文 B" {
		t.Fatalf("textOnly = %q", got)
	}
	if got := attrValue(`<a href="/x?a=1&amp;b=2">`, "href"); got != "/x?a=1&b=2" {
		t.Fatalf("attrValue = %q", got)
	}
	if got := firstYear("网络出版时间：2024-03-01"); got != 2024 {
		t.Fatalf("firstYear = %d", got)
	}
	if got := intOrZero("下载 1,234 次"); got != 1234 {
		t.Fatalf("intOrZero = %d", got)
	}
}
```

- [ ] **Step 2: Run helper test**

Run: `go test ./internal/cnki -run TestHTMLTextHelpers -count=1`

Expected: PASS before moving code.

- [ ] **Step 3: Move helpers**

Move from `search.go` to `htmlutil.go`: `textOnly`, `collapseSpace`, `attrValue`, `firstMatch`, `allMatches`, `firstYear`, `intOrZero`.

- [ ] **Step 4: Run full package tests**

Run: `go test ./internal/cnki -count=1`

Expected: PASS.

---

### Task 5: Harden Detail, References, And Citation Path

**Files:**
- Create: `internal/cnki/detail_parser_test.go`
- Modify: `internal/cnki/detail.go`
- Modify: `internal/cnki/refs.go`
- Keep: `internal/render/citation_test.go`

- [ ] **Step 1: Add detail parser test**

Create `internal/cnki/detail_parser_test.go`:

```go
package cnki

import "testing"

func TestParseDetailHTMLSupportsWxTitleLayout(t *testing.T) {
	t.Parallel()

	html := `<div class="wx-tit">
  <h1>论文标题</h1>
  <h3 class="author"><a>作者甲</a><a>作者乙</a></h3>
  <h3 class="orgn"><a>机构甲</a></h3>
</div>
<div class="top-tip"><a>中国电机工程学报</a><span class="year">2025年第01期</span></div>
<div id="ChDivSummary"> 摘要 A </div>
<p class="keywords"><a>关键词A；</a><a>关键词B</a></p>
<p>DOI：10.1/test 分类号：TM76</p>
<span id="annotationcount">5</span>`

	d := parseDetailHTML(html)

	if d.Title != "论文标题" || d.Source != "中国电机工程学报" || d.Issue != "2025年第01期" {
		t.Fatalf("basic detail = %#v", d)
	}
	if len(d.Authors) != 2 || d.Authors[0] != "作者甲" {
		t.Fatalf("authors = %#v", d.Authors)
	}
	if d.DOI != "10.1/test" || d.CLC != "TM76" || d.Year != 2025 {
		t.Fatalf("metadata = %#v", d)
	}
}
```

- [ ] **Step 2: Run parser and citation tests**

Run: `go test ./internal/cnki ./internal/render -run 'Test(ParseDetailHTMLSupportsWxTitleLayout|.*Citation.*)' -count=1`

Expected: detail parser test may FAIL before parser hardening; citation tests must PASS.

- [ ] **Step 3: Improve source/issue extraction**

In `detail.go`, extract source and issue from anchors/spans before falling back to flattened text:

```go
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
```

Use `source` and `issue` in the returned `model.Detail`.

- [ ] **Step 4: Run package and render tests**

Run: `go test ./internal/cnki ./internal/render -count=1`

Expected: PASS.

---

### Task 6: Documentation And Skill Alignment

**Files:**
- Modify: `README.md`
- Modify: `INSTALL.md`
- Modify: `skills/cnki-search/SKILL.md`
- Modify: `skills/cnki-search/references/cnki.net.md`

- [ ] **Step 1: Update README positioning**

Ensure `README.md` starts with this positioning:

```markdown
`cnki` 是一个用于查找参考文献的 CNKI CLI。核心目标是快速检索论文、查看详情/参考文献，并直接导出 GB/T 7714 风格引用。
```

- [ ] **Step 2: Update examples**

Keep this example visible in README and Skill docs:

```bash
cnki search "知识图谱" --size=20 --format=citation
```

- [ ] **Step 3: Remove core-download language**

Run: `rg -n "全文下载|PDF下载|CAJ下载|下载全文|download feature" README.md INSTALL.md skills`

Expected: no language presents full-text download as a core feature.

- [ ] **Step 4: Run stale-browser grep**

Run: `rg -n "chromedp|--headed|--chrome|cnki login|profile-dir" README.md INSTALL.md skills`

Expected: no stale command guidance remains except historical text explicitly saying those commands are not available.

---

## Execution Notes

- Preserve `--format=citation` before and after every refactor task.
- Do not implement NetEase OAuth in this plan.
- Do not implement full-text download in this plan.
- Do not implement `--source` until a fresh `brief/grid` request with source filters is captured and turned into a failing test.
- Keep each task behavior-preserving except the detail parser hardening task.

## Verification Checklist

- `go test ./... -count=1`
- `go build ./cmd/cnki`
- `go run ./cmd/cnki --help`
- `go run ./cmd/cnki search --help`
- `go run ./cmd/cnki search "测试" --size=1 --format=citation` only when live CNKI network validation is intentionally requested
- `rg -n "全文下载|PDF下载|CAJ下载|下载全文|download feature" README.md INSTALL.md skills`
- `rg -n "chromedp|--headed|--chrome|cnki login|profile-dir" README.md INSTALL.md skills`
