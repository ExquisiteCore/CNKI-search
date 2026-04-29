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
		KuaKuCode  string `json:"KuaKuCode"`
		SearchType int    `json:"SearchType"`
		SearchFrom int    `json:"SearchFrom"`
		QNode      struct {
			QGroup []struct {
				Items []struct {
					Field  string `json:"Field"`
					Value  string `json:"Value"`
					Value2 string `json:"Value2"`
				} `json:"Items"`
			} `json:"QGroup"`
		} `json:"QNode"`
	}
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}

	if got.Resource != "CROSSDB" || got.Classid != "WD0FTY92" {
		t.Fatalf("scope = %q/%q", got.Resource, got.Classid)
	}
	if got.KuaKuCode != "YSTT4HG0,JUP3MUPD" || got.SearchFrom != 4 || got.SearchType != 2 {
		t.Fatalf("crossids/searchFrom/searchType = %q/%d/%d", got.KuaKuCode, got.SearchFrom, got.SearchType)
	}
	item := got.QNode.QGroup[0].Items[0]
	if item.Field != "SU" || item.Value != "大语言模型" {
		t.Fatalf("keyword item = %#v", item)
	}
	year := got.QNode.QGroup[1].Items[0]
	if year.Field != "PT" || year.Value != "2020-01-01" || year.Value2 != "2025-12-31" {
		t.Fatalf("year item = %#v", year)
	}
}
