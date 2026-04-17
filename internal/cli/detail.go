package cli

import (
	"context"
	"fmt"

	"github.com/ExquisiteCore/cnki-search/internal/browser"
	"github.com/ExquisiteCore/cnki-search/internal/cnki"
	"github.com/ExquisiteCore/cnki-search/internal/render"
	"github.com/spf13/cobra"
)

func newDetailCmd() *cobra.Command {
	var withRefs bool
	cmd := &cobra.Command{
		Use:   "detail <url>",
		Short: "抽取论文详情页的完整元数据",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			ctx, cancel := context.WithTimeout(cmd.Context(), globals.Timeout)
			defer cancel()

			br, closeBr, err := browser.New(ctx, browserOptsFromGlobals())
			if err != nil {
				return withCode(err, 1)
			}
			defer closeBr()

			detail, err := cnki.Detail(br, url, withRefs)
			if err != nil {
				return withCode(err, cnki.ExitCodeFor(err))
			}
			out, err := render.Detail(detail, globals.Format)
			if err != nil {
				return withCode(err, 4)
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
	cmd.Flags().BoolVar(&withRefs, "with-refs", false, "同时抽取参考文献列表")
	return cmd
}
