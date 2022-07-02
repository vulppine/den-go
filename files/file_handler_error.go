package files

import "fmt"

type fileHandlerErrorCode int

const (
	accessError fileHandlerErrorCode = iota
	notAllowed
	invalidMethod
)

type fileHandlerError struct {
	file string
	code fileHandlerErrorCode
	err  error
}

func (e fileHandlerError) Error() string {
	switch e.code {
	case accessError:
		return fmt.Sprintf("error accessing file %s: %s", e.file, e.err)
	case notAllowed:
		return fmt.Sprintf("access denied for %s: %s", e.file, e.err)
	case invalidMethod:
		return fmt.Sprintf("invalid method")
	}

	return "no error detected: bug?"
}

func (e fileHandlerError) Unwrap() error {
	return e.err
}

func newFileHandlerError(code fileHandlerErrorCode, path string, err error) fileHandlerError {
	return fileHandlerError{path, code, err}
}
