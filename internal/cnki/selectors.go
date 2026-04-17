package cnki

// CNKI's AdvSearch page uses Vue + dynamically-rendered selectors that drift
// across releases. Everything DOM-facing lives here so future maintenance is
// one-file: update candidate selectors, no need to touch flow code.

// fieldCodeMap translates our external flag values to CNKI's internal
// `txt_1_sel` option codes. The dropdown's option values in AdvSearch have
// historically used short letter codes like SU / KY / TI / AU / AB / FT.
//
// If CNKI changes its codes again, update this map.
var fieldCodeMap = map[string]string{
	"topic":    "SU",  // 主题
	"keyword":  "KY",  // 关键词
	"title":    "TI",  // 篇名
	"author":   "AU",  // 作者
	"abstract": "AB",  // 摘要
	"fulltext": "FT",  // 全文
	"doi":      "DOI", // DOI
}

// sortCodeMap translates our sort flag to CNKI's sort hint. The AdvSearch UI
// exposes: 相关度 / 发表时间 / 被引 / 下载.
var sortCodeMap = map[string]string{
	"relevance": "",          // default
	"date":      "PT,down",   // 发表时间 降序
	"cited":     "CF,down",   // 被引
	"downloads": "DFR,down",  // 下载
}

// sourceFilterMap translates our source flag to the checkbox label CNKI
// uses in the "来源类别" panel.
var sourceFilterMap = map[string]string{
	"sci":   "SCI来源期刊",
	"ei":    "EI来源期刊",
	"core":  "北大核心",
	"cssci": "CSSCI",
	"cscd":  "CSCD",
}

// typeFilterMap translates our type flag to CNKI's document-type label.
var typeFilterMap = map[string]string{
	"journal":    "期刊",
	"master":     "硕士",
	"phd":        "博士",
	"conference": "会议",
	"newspaper":  "报纸",
	"yearbook":   "年鉴",
}

// Candidate selectors. Each list is tried in order; the first match wins.
// Keep these loose — CNKI's kns8s SPA re-renders with different class names
// depending on the logged-in state and A/B buckets.
var (
	selSearchInput = []string{
		`input[name="txt_1_value1"]`,
		`.input-box input.ant-input`,
		`.search-box input[type="text"]`,
	}
	selFieldDropdown = []string{
		`select[name="txt_1_sel"]`,
		`.search-box .ant-select:first-of-type`,
	}
	selFromYear = []string{
		`input[name="txt_1_from"]`,
		`.time-range input:nth-of-type(1)`,
	}
	selToYear = []string{
		`input[name="txt_1_to"]`,
		`.time-range input:nth-of-type(2)`,
	}
	selSubmitBtn = []string{
		`input.btn-search`,
		`.search-btn`,
		`button.search-btn`,
		`button[type="submit"]`,
	}
	selResultRow = []string{
		`.result-table-list tbody tr`,
		`#gridTable tbody tr`,
		`table.result-table-list tbody tr`,
	}
	selNextPage = []string{
		`#PageNext`,
		`a.next`,
		`.pagebar a.next`,
	}
	selTotalHits = []string{
		`span.pagerTitleCell em`,
		`.page-sum em`,
		`.search-result-total em`,
	}
)
