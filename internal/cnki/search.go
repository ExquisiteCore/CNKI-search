package cnki

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/model"
)

// Search searches CNKI through the kns8s brief/grid HTTP endpoint.
func (c *Client) Search(ctx context.Context, q model.Query) (*model.SearchResult, error) {
	if strings.TrimSpace(q.Q) == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if q.Size <= 0 {
		q.Size = defaultPageSize
	}
	if len(q.Sources) > 0 {
		return nil, fmt.Errorf("source filters are not supported by HTTP mode yet")
	}

	field := fieldCodeMap[q.Field]
	if field == "" {
		field = "SU"
	}
	sort := httpSortCodeMap[q.Sort]

	crossIDs := crossIDsForTypes(q.Types)
	if crossIDs == "" {
		crossIDs = defaultCrossIDs
	}

	if err := c.warmSearch(ctx, q.Q, field, crossIDs); err != nil {
		return nil, fmt.Errorf("warm search session: %w", err)
	}

	collected := make([]model.Paper, 0, q.Size)
	seen := make(map[string]bool)
	turnpage := ""
	total := 0
	maxPages := q.Size/defaultPageSize + 3

	for page := 1; page <= maxPages && len(collected) < q.Size; page++ {
		gp, err := c.fetchGridPage(ctx, q, field, crossIDs, sort, page, turnpage)
		if err != nil {
			return nil, err
		}
		if total == 0 && gp.total > 0 {
			total = gp.total
		}
		if len(gp.papers) == 0 {
			break
		}
		for _, paper := range gp.papers {
			key := paper.URL
			if key == "" {
				key = paper.Title
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			paper.Seq = len(collected) + 1
			collected = append(collected, paper)
			if len(collected) >= q.Size {
				break
			}
		}
		turnpage = gp.turnpage
		if turnpage == "" {
			break
		}
		if gp.maxPage > 0 && page >= gp.maxPage {
			break
		}
	}

	if len(collected) == 0 {
		return nil, ErrEmpty
	}
	return &model.SearchResult{
		Query:     q,
		TotalHits: total,
		Fetched:   len(collected),
		Results:   collected,
	}, nil
}

func (c *Client) warmSearch(ctx context.Context, keyword, field, crossIDs string) error {
	home := c.resolve("/")
	if c.baseURL.Host == "kns.cnki.net" {
		home = URLHome
	}
	req, err := c.newRequest(ctx, http.MethodGet, home, nil)
	if err != nil {
		return err
	}
	if _, err := c.doText(req); err != nil {
		return err
	}
	if err := c.ensureClientID(ctx); err != nil {
		return err
	}

	values := url.Values{}
	values.Set("rc", starterResources)
	values.Set("kw", keyword)
	values.Set("rt", "crossdb")
	values.Set("fd", field)
	req, err = c.newRequest(ctx, http.MethodGet, c.resolve("/starter?"+values.Encode()), nil)
	if err != nil {
		return err
	}
	if _, err := c.doText(req); err != nil {
		return err
	}

	req, err = c.newRequest(ctx, http.MethodGet, c.searchIndexURL(keyword, field, crossIDs), nil)
	if err != nil {
		return err
	}
	_, err = c.doText(req)
	return err
}

func (c *Client) fetchGridPage(ctx context.Context, q model.Query, field, crossIDs string, sort sortSpec, page int, turnpage string) (gridPage, error) {
	isFirst := page == 1 && turnpage == ""
	searchFrom := 1
	if !isFirst {
		searchFrom = 4
	}

	queryJSON, err := buildQueryJSON(q, field, crossIDs, searchFrom)
	if err != nil {
		return gridPage{}, err
	}

	form := url.Values{}
	form.Set("boolSearch", strconv.FormatBool(isFirst))
	form.Set("QueryJson", queryJSON)
	form.Set("queryJson", queryJSON)
	form.Set("pageNum", strconv.Itoa(page))
	form.Set("pageSize", strconv.Itoa(defaultPageSize))
	form.Set("CurPage", strconv.Itoa(page))
	form.Set("RecordsCntPerPage", strconv.Itoa(defaultPageSize))
	form.Set("sortField", valueIf(!isFirst, sort.field))
	form.Set("sortType", valueIf(!isFirst, sort.order))
	form.Set("dstyle", "listmode")
	form.Set("productStr", defaultProductStr)
	form.Set("aside", valueIf(isFirst, fieldTitleMap[field]+"："+q.Q))
	form.Set("searchFrom", "资源范围：总库")
	form.Set("language", "")
	form.Set("uniplatform", "")
	if turnpage != "" {
		form.Set("turnpage", turnpage)
	}

	req, err := c.newRequest(ctx, http.MethodPost, c.resolve(URLBriefGridPath), strings.NewReader(form.Encode()))
	if err != nil {
		return gridPage{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Origin", c.baseURL.Scheme+"://"+c.baseURL.Host)
	req.Header.Set("Referer", c.searchIndexURL(q.Q, field, crossIDs))
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	body, err := c.doText(req)
	if err != nil {
		return gridPage{}, err
	}
	if strings.Contains(body, `class="no-content"`) {
		return gridPage{}, ErrEmpty
	}
	gp := parseGridHTML(body, c)
	gp.pageIndex = page
	return gp, nil
}

func (c *Client) searchIndexURL(keyword, field, crossIDs string) string {
	values := url.Values{}
	values.Set("crossids", crossIDs)
	values.Set("korder", field)
	values.Set("kw", keyword)
	return c.resolve("/kns8s/defaultresult/index?" + values.Encode())
}
