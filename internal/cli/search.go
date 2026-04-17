package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/ExquisiteCore/cnki-search/internal/browser"
	"github.com/ExquisiteCore/cnki-search/internal/cnki"
	"github.com/ExquisiteCore/cnki-search/internal/model"
	"github.com/ExquisiteCore/cnki-search/internal/render"
	"github.com/spf13/cobra"
)

type searchFlags struct {
	field   string
	from    int
	to      int
	types   []string
	sources []string
	sort    string
	size    int
}

func newSearchCmd() *cobra.Command {
	var f searchFlags
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "在知网上检索学术论文",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			q := model.Query{
				Q:       query,
				Field:   f.field,
				From:    f.from,
				To:      f.to,
				Types:   f.types,
				Sources: f.sources,
				Sort:    f.sort,
				Size:    f.size,
			}
			if err := validateSearch(&q); err != nil {
				return withCode(err, 4)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), globals.Timeout)
			defer cancel()

			br, closeBr, err := browser.New(ctx, browserOptsFromGlobals())
			if err != nil {
				return withCode(err, 1)
			}
			defer closeBr()

			result, err := cnki.Search(br, q)
			if err != nil {
				return withCode(err, cnki.ExitCodeFor(err))
			}

			out, err := render.Search(result, globals.Format)
			if err != nil {
				return withCode(err, 4)
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}

	pf := cmd.Flags()
	pf.StringVar(&f.field, "field", "topic", "检索字段：topic|keyword|title|author|abstract|fulltext|doi")
	pf.IntVar(&f.from, "from", 0, "起始年份（含）")
	pf.IntVar(&f.to, "to", 0, "截止年份（含）")
	pf.StringSliceVar(&f.types, "type", nil, "文献类型：journal|master|phd|conference|newspaper|yearbook")
	pf.StringSliceVar(&f.sources, "source", nil, "来源类型：sci|ei|core|cssci|cscd")
	pf.StringVar(&f.sort, "sort", "relevance", "排序方式：relevance|date|cited|downloads")
	pf.IntVar(&f.size, "size", 20, "需要的结果数量（自动翻页满足）")
	return cmd
}

func validateSearch(q *model.Query) error {
	if strings.TrimSpace(q.Q) == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if q.Size <= 0 {
		return fmt.Errorf("--size must be > 0")
	}
	if q.Size > 500 {
		return fmt.Errorf("--size too large (max 500 to avoid rate-limit)")
	}
	if q.From != 0 && q.To != 0 && q.From > q.To {
		return fmt.Errorf("--from (%d) cannot be greater than --to (%d)", q.From, q.To)
	}
	return nil
}

func browserOptsFromGlobals() browser.Options {
	return browser.Options{
		Headless:   !globals.Headed,
		ChromePath: globals.ChromePath,
		ProfileDir: globals.ProfileDir,
		Verbose:    globals.Verbose,
	}
}
