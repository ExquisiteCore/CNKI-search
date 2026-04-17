package cli

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// GlobalFlags are the flags shared by every subcommand.
type GlobalFlags struct {
	Headed     bool
	Timeout    time.Duration
	ChromePath string
	ProfileDir string
	Format     string
	Verbose    bool
}

var globals GlobalFlags

func defaultProfileDir() string {
	cache, err := os.UserCacheDir()
	if err != nil || cache == "" {
		if home, herr := os.UserHomeDir(); herr == nil {
			return filepath.Join(home, ".cache", "cnki-search", "chrome")
		}
		return filepath.Join(".", ".cnki-search", "chrome")
	}
	return filepath.Join(cache, "cnki-search", "chrome")
}

// NewRoot builds the top-level cobra command with all subcommands registered.
func NewRoot(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cnki",
		Short:   "CNKI (知网) 学术论文命令行检索工具",
		Long:    "cnki 通过 chromedp 驱动 Chrome 访问中国知网，支持关键词检索、论文详情抽取与参考文献导出。",
		Version: version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := cmd.PersistentFlags()
	pf.BoolVar(&globals.Headed, "headed", false, "以有头模式启动 Chrome（调试 / 登录 / 过验证码时使用）")
	pf.DurationVar(&globals.Timeout, "timeout", 90*time.Second, "单次命令整体超时")
	pf.StringVar(&globals.ChromePath, "chrome", "", "Chrome/Edge 可执行文件路径（默认自动探测）")
	pf.StringVar(&globals.ProfileDir, "profile-dir", defaultProfileDir(), "浏览器用户数据目录（保留登录态）")
	pf.StringVar(&globals.Format, "format", "json", "输出格式：json|table|citation|markdown")
	pf.BoolVarP(&globals.Verbose, "verbose", "v", false, "打印 chromedp 调试日志到 stderr")

	cmd.AddCommand(
		newSearchCmd(),
		newDetailCmd(),
		newRefsCmd(),
		newLoginCmd(),
	)
	return cmd
}
