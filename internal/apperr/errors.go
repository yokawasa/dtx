package apperr

import "errors"

type Error struct {
	Message string
	Code    int
	Silent  bool
}

func (e *Error) Error() string {
	return e.Message
}

func New(message string) *Error {
	return &Error{Message: message, Code: 1}
}

func Silent(code int) *Error {
	return &Error{Message: "command failed", Code: code, Silent: true}
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var appErr *Error
	if errors.As(err, &appErr) && appErr.Code != 0 {
		return appErr.Code
	}
	return 1
}

func IsSilent(err error) bool {
	var appErr *Error
	return errors.As(err, &appErr) && appErr.Silent
}
