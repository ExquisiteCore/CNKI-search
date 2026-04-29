package cli

import (
	"context"
	"fmt"

	"github.com/ExquisiteCore/cnki-search/internal/cnki"
	"github.com/ExquisiteCore/cnki-search/internal/render"
	"github.com/spf13/cobra"
)

func newRefsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refs <url>",
		Short: "抽取论文的参考文献列表",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]
			ctx, cancel := context.WithTimeout(cmd.Context(), globals.Timeout)
			defer cancel()

			client := cnki.NewClient(cnki.ClientOptions{UserAgent: globals.UserAgent})
			refs, err := client.References(ctx, url)
			if err != nil {
				return withCode(err, cnki.ExitCodeFor(err))
			}
			out, err := render.References(refs, globals.Format)
			if err != nil {
				return withCode(err, 4)
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
	return cmd
}
