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
