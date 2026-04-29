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
