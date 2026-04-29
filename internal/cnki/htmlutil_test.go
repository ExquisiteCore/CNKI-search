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
