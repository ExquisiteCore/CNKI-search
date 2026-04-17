package browser

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// ErrCaptcha is returned when the page shows a slider / captcha / anti-bot wall
// and the command cannot proceed automatically.
var ErrCaptcha = errors.New("captcha or anti-bot challenge detected")

// CaptchaJS is evaluated in the page; it returns true if a captcha element is
// currently visible. Selectors are generous — better to false-positive and
// surface an actionable error than to hang.
const captchaJS = `(() => {
  const sels = [
    ".slide-verify",
    ".nc_wrapper",
    ".nc-container",
    "#nc_1_wrapper",
    "iframe[src*='captcha']",
    "iframe[src*='verify']",
    ".captcha",
    ".verifyWrap",
  ];
  for (const s of sels) {
    const el = document.querySelector(s);
    if (el) {
      const r = el.getBoundingClientRect();
      if (r.width > 0 && r.height > 0) return true;
    }
  }
  const txt = document.body && document.body.innerText || "";
  if (txt.includes("请完成安全验证") || txt.includes("滑动验证")) return true;
  return false;
})()`

// DetectCaptcha evaluates the captcha probe on the current page.
func (b *Browser) DetectCaptcha() (bool, error) {
	var present bool
	if err := chromedp.Run(b.Ctx, chromedp.Evaluate(captchaJS, &present)); err != nil {
		return false, err
	}
	return present, nil
}

// HandleCaptcha is invoked by callers at risk points. In headed mode it waits
// for the user to solve it (polling every 3s, up to 2 minutes). In headless
// mode it dumps a screenshot and returns ErrCaptcha.
func (b *Browser) HandleCaptcha() error {
	present, err := b.DetectCaptcha()
	if err != nil {
		return err
	}
	if !present {
		return nil
	}
	if b.opts.Headless {
		shot := filepath.Join(os.TempDir(), fmt.Sprintf("cnki-captcha-%d.png", time.Now().Unix()))
		_ = b.Screenshot(shot)
		fmt.Fprintf(os.Stderr, "检测到验证码。截图已保存：%s\n", shot)
		fmt.Fprintln(os.Stderr, "请改用 --headed 重试，或先运行 `cnki login` 在有头模式过验证。")
		return ErrCaptcha
	}
	fmt.Fprintln(os.Stderr, "检测到验证码。请在弹出的 Chrome 窗口手动完成验证，等待中...")
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)
		p, derr := b.DetectCaptcha()
		if derr != nil {
			return derr
		}
		if !p {
			fmt.Fprintln(os.Stderr, "验证通过，继续执行。")
			return nil
		}
	}
	return ErrCaptcha
}

// Screenshot writes a full-page screenshot to path.
func (b *Browser) Screenshot(path string) error {
	var buf []byte
	if err := chromedp.Run(b.Ctx, chromedp.FullScreenshot(&buf, 80)); err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0o644)
}
