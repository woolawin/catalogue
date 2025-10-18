package internal

import (
	"fmt"
	"strings"
)

type CLErr struct {
	message string
	parent  *CLErr
}

func ErrOf(err error, format string, args ...any) *CLErr {
	return &CLErr{message: fmt.Sprintf(format, args...), parent: ErrCause(err)}
}

func ErrCause(err error) *CLErr {
	return &CLErr{message: err.Error()}
}

func ErrFileBlocked(path string, action string) *CLErr {
	return &CLErr{message: fmt.Sprintf("file '%s' is not permitted to be %s", path, action)}
}

func Err(format string, args ...any) *CLErr {
	return &CLErr{message: fmt.Sprintf(format, args...)}
}

func (err *CLErr) Error() string {
	builder := strings.Builder{}
	builder.WriteString(err.message)
	var parent *CLErr = err.parent
	for parent != nil {
		builder.WriteString("\n")
		builder.WriteString("↳ ")
		builder.WriteString(parent.message)
		parent = parent.parent
	}
	return builder.String()
}
