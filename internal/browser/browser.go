package browser

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

// Options controls how a Browser instance is created.
type Options struct {
	Headless   bool
	ChromePath string
	ProfileDir string
	Verbose    bool
}

// Browser wraps a chromedp execution context plus its allocator cancel.
type Browser struct {
	Ctx        context.Context
	allocCanc  context.CancelFunc
	ctxCancel  context.CancelFunc
	opts       Options
	logWriter  io.Writer
}

// New builds the chromedp context with the given options and returns a Browser
// together with a close func that must be deferred by the caller.
func New(parent context.Context, opts Options) (*Browser, func(), error) {
	if opts.ProfileDir == "" {
		return nil, nil, fmt.Errorf("ProfileDir is required")
	}
	if err := os.MkdirAll(opts.ProfileDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create profile dir: %w", err)
	}

	execOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.UserAgent(defaultUA),
		chromedp.UserDataDir(opts.ProfileDir),
		chromedp.WindowSize(1440, 900),
	)
	if opts.ChromePath != "" {
		execOpts = append(execOpts, chromedp.ExecPath(opts.ChromePath))
	} else if guessed := guessChromePath(); guessed != "" {
		execOpts = append(execOpts, chromedp.ExecPath(guessed))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(parent, execOpts...)

	var logWriter io.Writer = io.Discard
	if opts.Verbose {
		logWriter = os.Stderr
	}
	logger := log.New(logWriter, "[chromedp] ", log.LstdFlags)

	ctx, ctxCancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(logger.Printf),
	)

	// Force Chrome to actually boot so errors surface early.
	if err := chromedp.Run(ctx); err != nil {
		ctxCancel()
		allocCancel()
		return nil, nil, fmt.Errorf("launch chrome: %w", err)
	}

	br := &Browser{
		Ctx:       ctx,
		allocCanc: allocCancel,
		ctxCancel: ctxCancel,
		opts:      opts,
		logWriter: logWriter,
	}
	closeFn := func() {
		ctxCancel()
		allocCancel()
	}
	return br, closeFn, nil
}

// Options returns the options this Browser was created with.
func (b *Browser) Options() Options { return b.opts }

const defaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"

// guessChromePath tries a few well-known install locations on the current OS.
// Empty string means "let chromedp search the PATH itself".
func guessChromePath() string {
	candidates := osChromeCandidates()
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func expandEnv(p string) string {
	return filepath.Clean(os.ExpandEnv(p))
}
