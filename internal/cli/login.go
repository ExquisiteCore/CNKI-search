package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ExquisiteCore/cnki-search/internal/browser"
	"github.com/ExquisiteCore/cnki-search/internal/cnki"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "有头模式打开知网，用户手动登录后保存 cookie 到本地 profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			// login 永远走有头模式
			opts := browserOptsFromGlobals()
			opts.Headless = false

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
			defer cancel()

			br, closeBr, err := browser.New(ctx, opts)
			if err != nil {
				return withCode(err, 1)
			}
			defer closeBr()

			if err := cnki.OpenHome(br); err != nil {
				return withCode(err, 1)
			}

			fmt.Fprintln(os.Stderr, "已打开知网。请在弹出的 Chrome 窗口完成登录。")
			fmt.Fprintln(os.Stderr, "登录完成后，回到此终端，按 Enter 退出并保存 cookie...")
			reader := bufio.NewReader(os.Stdin)
			_, _ = reader.ReadString('\n')
			fmt.Fprintln(os.Stderr, "cookie 已随 profile-dir 保存，后续无头命令可复用登录态。")
			return nil
		},
	}
	return cmd
}
