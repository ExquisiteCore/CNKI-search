package cnki

import "errors"

// Sentinel errors callers (CLI) can key exit codes off of.
var (
	ErrCaptcha = errors.New("captcha or anti-bot challenge detected")
	ErrEmpty   = errors.New("no results matched the query")
)

// ExitCodeFor maps internal errors to process exit codes.
func ExitCodeFor(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, ErrCaptcha):
		return 2
	case errors.Is(err, ErrEmpty):
		return 3
	default:
		return 1
	}
}
