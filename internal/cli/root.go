package cli

import (
	"time"

	"github.com/spf13/cobra"
)

// GlobalFlags are the flags shared by every subcommand.
type GlobalFlags struct {
	Timeout   time.Duration
	Format    string
	UserAgent string
}

var globals GlobalFlags

// NewRoot builds the top-level cobra command with all subcommands registered.
func NewRoot(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "cnki",
		Short:         "CNKI (知网) 参考文献检索与引用导出 CLI",
		Long:          "cnki 通过 HTTP 访问中国知网 kns8s 接口，用于查找参考文献，支持关键词检索、论文详情抽取、参考文献列表与 GB/T 7714 风格引用导出。",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := cmd.PersistentFlags()
	pf.DurationVar(&globals.Timeout, "timeout", 90*time.Second, "单次命令整体超时")
	pf.StringVar(&globals.Format, "format", "json", "输出格式：json|table|citation|markdown")
	pf.StringVar(&globals.UserAgent, "user-agent", "", "HTTP User-Agent（默认模拟 Chrome）")

	cmd.AddCommand(
		newSearchCmd(),
		newDetailCmd(),
		newRefsCmd(),
	)
	return cmd
}
