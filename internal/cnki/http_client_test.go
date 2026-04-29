package cnki

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

func TestHTTPClientSearchPostsBriefGridAndParsesRows(t *testing.T) {
	t.Parallel()

	var gridCalls int
	var sawWarm bool
	var sawStarter bool
	var sawNextTurnpage bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			sawWarm = true
			w.WriteHeader(http.StatusOK)
		case "/starter":
			sawStarter = true
			if r.URL.Query().Get("kw") != "深度学习" || r.URL.Query().Get("fd") != "SU" {
				t.Fatalf("starter query = %s", r.URL.RawQuery)
			}
			w.WriteHeader(http.StatusOK)
		case "/kns8s/defaultresult/index":
			if r.URL.Query().Get("kw") != "深度学习" {
				t.Fatalf("warm query kw = %q", r.URL.Query().Get("kw"))
			}
			w.WriteHeader(http.StatusOK)
		case "/kns8s/brief/grid":
			gridCalls++
			if r.Method != http.MethodPost {
				t.Fatalf("brief/grid method = %s", r.Method)
			}
			if got := r.Header.Get("X-Requested-With"); got != "XMLHttpRequest" {
				t.Fatalf("X-Requested-With = %q", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			assertFormHas(t, r.Form, "QueryJson")
			assertFormHas(t, r.Form, "pageSize")
			assertFormHas(t, r.Form, "CurPage")

			var query struct {
				Resource   string `json:"Resource"`
				Classid    string `json:"Classid"`
				KuaKuCode  string `json:"KuaKuCode"`
				SearchType int    `json:"SearchType"`
				SearchFrom int    `json:"SearchFrom"`
				QNode      struct {
					QGroup []struct {
						Items []struct {
							Field string `json:"Field"`
							Value string `json:"Value"`
						} `json:"Items"`
					} `json:"QGroup"`
				} `json:"QNode"`
			}
			if err := json.Unmarshal([]byte(r.Form.Get("QueryJson")), &query); err != nil {
				t.Fatalf("QueryJson is invalid JSON: %v\n%s", err, r.Form.Get("QueryJson"))
			}
			if query.Resource != "CROSSDB" || query.Classid != "WD0FTY92" {
				t.Fatalf("unexpected query scope: resource=%q classid=%q", query.Resource, query.Classid)
			}
			if query.KuaKuCode == "" || query.SearchType != 2 {
				t.Fatalf("unexpected query mode: kuaku=%q searchType=%d", query.KuaKuCode, query.SearchType)
			}
			if len(query.QNode.QGroup) == 0 || len(query.QNode.QGroup[0].Items) == 0 {
				t.Fatalf("query JSON did not contain the keyword condition: %#v", query.QNode)
			}
			item := query.QNode.QGroup[0].Items[0]
			if item.Field != "SU" || item.Value != "深度学习" {
				t.Fatalf("query item = field %q value %q", item.Field, item.Value)
			}

			switch gridCalls {
			case 1:
				if query.SearchFrom != 1 {
					t.Fatalf("first SearchFrom = %d", query.SearchFrom)
				}
				writeHTML(w, searchGridHTML("turn-token", "2",
					rowHTML(1, "题名一", "/kcms2/article/abstract?v=one", "张三", "软件学报", "2024-03", "15", "120", "期刊"),
				))
			case 2:
				if query.SearchFrom != 4 {
					t.Fatalf("next SearchFrom = %d", query.SearchFrom)
				}
				if r.Form.Get("turnpage") != "turn-token" {
					t.Fatalf("turnpage = %q", r.Form.Get("turnpage"))
				}
				sawNextTurnpage = true
				writeHTML(w, searchGridHTML("", "2",
					rowHTML(2, "题名二", "https://kns.cnki.net/kcms2/article/abstract?v=two", "李四; 王五", "计算机学报", "2023", "6", "80", "期刊"),
				))
			default:
				t.Fatalf("unexpected grid call %d", gridCalls)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := NewClient(ClientOptions{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	result, err := client.Search(context.Background(), model.Query{
		Q:     "深度学习",
		Field: "topic",
		Sort:  "date",
		Size:  2,
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if !sawWarm || !sawStarter {
		t.Fatalf("client did not warm the session: home=%v starter=%v", sawWarm, sawStarter)
	}
	if !sawNextTurnpage {
		t.Fatal("client did not forward turnpage token to the next request")
	}
	if result.TotalHits != 1234 || result.Fetched != 2 {
		t.Fatalf("result counts = total %d fetched %d", result.TotalHits, result.Fetched)
	}
	if got := result.Results[0]; got.Seq != 1 || got.Title != "题名一" || got.URL != srv.URL+"/kcms2/article/abstract?v=one" || got.Year != 2024 || got.Cited != 15 || got.Downloads != 120 {
		t.Fatalf("first paper parsed incorrectly: %#v", got)
	}
	if got := strings.Join(result.Results[1].Authors, ","); got != "李四,王五" {
		t.Fatalf("authors = %q", got)
	}
}

func TestHTTPClientDetectsSecurityVerificationPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeHTML(w, `<html><title>安全验证</title><body>拖动下方拼图完成验证</body></html>`)
	}))
	defer srv.Close()

	client := NewClient(ClientOptions{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	req, err := client.newRequest(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.doText(req)
	if !errors.Is(err, ErrCaptcha) {
		t.Fatalf("err = %v, want ErrCaptcha", err)
	}
}

func TestHTTPClientDetailAndReferencesParseHTML(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/kcms2/article/abstract" {
			http.NotFound(w, r)
			return
		}
		writeHTML(w, `<!doctype html>
<html><body>
<h1>基于深度学习的图像识别研究</h1>
<div class="author"><a>张三</a><a>李四</a></div>
<div class="orgn"><a>清华大学</a></div>
<div id="ChDivSummary">摘要内容。</div>
<p class="keywords"><a>深度学习；</a><a>图像识别</a></p>
<div class="top-tip">软件学报 2024年第3期 DOI：10.123/test 分类号：TP391</div>
<p class="funds">国家自然科学基金</p>
<span id="annotationcount">被引 12</span><span id="downloadcount">下载 34</span>
<div id="CataLogContent">
  <ul>
    <li>[1] 参考文献一.</li>
    <li>[2] 参考文献二.</li>
  </ul>
</div>
</body></html>`)
	}))
	defer srv.Close()

	client := NewClient(ClientOptions{
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
	})
	detail, err := client.Detail(context.Background(), srv.URL+"/kcms2/article/abstract?v=abc", true)
	if err != nil {
		t.Fatalf("Detail returned error: %v", err)
	}
	if detail.Title != "基于深度学习的图像识别研究" || detail.Abstract != "摘要内容。" {
		t.Fatalf("detail parsed incorrectly: %#v", detail)
	}
	if strings.Join(detail.Authors, ",") != "张三,李四" {
		t.Fatalf("authors = %#v", detail.Authors)
	}
	if detail.DOI != "10.123/test" || detail.CLC != "TP391" || detail.Year != 2024 || detail.Cited != 12 || detail.Downloads != 34 {
		t.Fatalf("metadata parsed incorrectly: %#v", detail)
	}
	if len(detail.References) != 2 || detail.References[1].Text != "[2] 参考文献二." {
		t.Fatalf("references parsed incorrectly: %#v", detail.References)
	}

	refs, err := client.References(context.Background(), srv.URL+"/kcms2/article/abstract?v=abc")
	if err != nil {
		t.Fatalf("References returned error: %v", err)
	}
	if len(refs) != 2 || refs[0].Seq != 1 {
		t.Fatalf("refs parsed incorrectly: %#v", refs)
	}
}

func assertFormHas(t *testing.T, form url.Values, key string) {
	t.Helper()
	if form.Get(key) == "" {
		t.Fatalf("form key %q is empty; form=%v", key, form)
	}
}

func writeHTML(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

func searchGridHTML(turnpage, maxPage string, rows ...string) string {
	var b strings.Builder
	b.WriteString(`<span>共找到<em>1,234</em>条</span>`)
	if turnpage != "" {
		b.WriteString(`<input id="hidTurnPage" value="` + turnpage + `">`)
	}
	if maxPage != "" {
		b.WriteString(`<span class="countPageMark" data-pagenum="` + maxPage + `"></span>`)
	}
	b.WriteString(`<table><tbody>`)
	for _, row := range rows {
		b.WriteString(row)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func rowHTML(seq int, title, href, authors, source, date, cited, downloads, database string) string {
	var authorHTML strings.Builder
	for _, author := range strings.FieldsFunc(authors, func(r rune) bool {
		return r == ';' || r == '；' || r == ',' || r == '，'
	}) {
		author = strings.TrimSpace(author)
		if author == "" {
			continue
		}
		authorHTML.WriteString(`<a class="KnowledgeNetLink">` + author + `</a>`)
	}
	return `<tr>` +
		`<td class="seq">` + strconv.Itoa(seq) + `</td>` +
		`<td class="name"><a class="fz14" href="` + href + `">` + title + `</a></td>` +
		`<td class="author">` + authorHTML.String() + `</td>` +
		`<td class="source"><a>` + source + `</a></td>` +
		`<td class="date">` + date + `</td>` +
		`<td class="quote">` + cited + `</td>` +
		`<td class="download">` + downloads + `</td>` +
		`<td class="data">` + database + `</td>` +
		`</tr>`
}
