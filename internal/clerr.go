package internal

import (
	"fmt"
	"strings"
)

type CLErr struct {
	Message string
	Parent  *CLErr
}

func ErrOf(err error, format string, args ...any) *CLErr {
	return &CLErr{Message: fmt.Sprintf(format, args...), Parent: ErrCause(err)}
}

func ErrCause(err error) *CLErr {
	return &CLErr{Message: err.Error()}
}

func ErrFileBlocked(path string, action string) *CLErr {
	return &CLErr{Message: fmt.Sprintf("file '%s' is not permitted to be %s", path, action)}
}

func Err(format string, args ...any) *CLErr {
	return &CLErr{Message: fmt.Sprintf(format, args...)}
}

func (err *CLErr) Error() string {
	builder := strings.Builder{}
	builder.WriteString(err.Message)
	var parent *CLErr = err.Parent
	for parent != nil {
		builder.WriteString("\n")
		builder.WriteString("↳ ")
		builder.WriteString(parent.Message)
		parent = parent.Parent
	}
	return builder.String()
}
