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
