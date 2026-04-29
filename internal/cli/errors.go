package cli

// CodedError lets a command request a specific process exit code.
// Exit codes in use:
//
//	1  generic error
//	2  captcha/anti-bot intercepted
//	3  no results / empty search
//	4  invalid argument
type CodedError interface {
	error
	ExitCode() int
}

type codedError struct {
	err  error
	code int
}

func (c *codedError) Error() string { return c.err.Error() }
func (c *codedError) Unwrap() error { return c.err }
func (c *codedError) ExitCode() int { return c.code }

func withCode(err error, code int) error {
	if err == nil {
		return nil
	}
	return &codedError{err: err, code: code}
}
