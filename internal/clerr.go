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
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return &CLErr{Message: msg}
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
