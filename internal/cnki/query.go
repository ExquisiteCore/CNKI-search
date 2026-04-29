package cnki

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

const (
	defaultClassID    = "WD0FTY92"
	defaultCrossIDs   = "YSTT4HG0,LSTPFY1C,JUP3MUPD,MPMFIG1A,WQ0UVIAA,BLZOG7CK,PWFIRAGL,EMRPGLPA,NLBO1Z6R,NN3FJMUV"
	defaultProductStr = "YSTT4HG0,LSTPFY1C,RMJLXHZ3,JQIRZIYA,JUP3MUPD,1UR4K4HZ,BPBAFJ5S,R79MZMCB,MPMFIG1A,WQ0UVIAA,NB3BWEHK,XVLO76FD,HR1YT1Z9,BLZOG7CK,PWFIRAGL,EMRPGLPA,J708GVCE,ML4DRIDX,NLBO1Z6R,NN3FJMUV,"
	defaultPageSize   = 20
	starterResources  = "CJFQ,CDMD,CIPD,CCND,CISD,SNAD,CCJD,BDZK,CCVD,CJFN"
)

var fieldTitleMap = map[string]string{
	"SU":  "主题",
	"TKA": "篇关摘",
	"KY":  "关键词",
	"TI":  "篇名",
	"AU":  "作者",
	"AB":  "摘要",
	"FT":  "全文",
	"DOI": "DOI",
}

var typeCrossIDMap = map[string]string{
	"journal":    "YSTT4HG0",
	"master":     "LSTPFY1C",
	"phd":        "LSTPFY1C",
	"conference": "JUP3MUPD",
	"newspaper":  "MPMFIG1A",
	"yearbook":   "NLBO1Z6R",
}

type sortSpec struct {
	field string
	order string
}

var httpSortCodeMap = map[string]sortSpec{
	"relevance": {"FFD", "desc"},
	"date":      {"PT", "desc"},
	"cited":     {"CF", "desc"},
	"downloads": {"DFR", "desc"},
}

func buildQueryJSON(q model.Query, field, crossIDs string, searchFrom int) (string, error) {
	title := fieldTitleMap[field]
	if title == "" {
		title = field
	}

	qgroups := []map[string]any{
		{
			"Key":   "Subject",
			"Title": "",
			"Logic": 0,
			"Items": []map[string]any{
				{
					"Field":    field,
					"Value":    q.Q,
					"Operator": "TOPRANK",
					"Logic":    0,
					"Title":    title,
				},
			},
			"ChildItems": []any{},
		},
	}

	if q.From > 0 || q.To > 0 {
		from := q.From
		if from == 0 {
			from = 1900
		}
		to := q.To
		if to == 0 {
			to = 2100
		}
		qgroups = append(qgroups, map[string]any{
			"Key":   "ControlGroup",
			"Title": "",
			"Logic": 0,
			"Items": []map[string]any{
				{
					"Key":      "span[value=PT]",
					"Title":    "发表时间",
					"Logic":    0,
					"Field":    "PT",
					"Operator": 7,
					"Value":    fmt.Sprintf("%04d-01-01", from),
					"Value2":   fmt.Sprintf("%04d-12-31", to),
				},
			},
			"ChildItems": []any{},
		})
	}

	payload := map[string]any{
		"Platform":   "",
		"Resource":   "CROSSDB",
		"Classid":    defaultClassID,
		"Products":   "",
		"QNode":      map[string]any{"QGroup": qgroups},
		"ExScope":    1,
		"SearchType": 2,
		"Rlang":      "CHINESE",
		"KuaKuCode":  crossIDs,
		"Expands":    map[string]any{},
		"SearchFrom": searchFrom,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func crossIDsForTypes(types []string) string {
	if len(types) == 0 {
		return ""
	}
	out := make([]string, 0, len(types))
	seen := map[string]bool{}
	for _, t := range types {
		code := typeCrossIDMap[t]
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		out = append(out, code)
	}
	return strings.Join(out, ",")
}

func valueIf(ok bool, value string) string {
	if ok {
		return value
	}
	return ""
}
