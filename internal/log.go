package internal

import (
	"fmt"
	"strings"

	stdout "github.com/fatih/color"
)

type Log interface {
	Stage(value string)
	File(value string)

	Info(level int, format string, args ...any)
	Error(level int, format string, args ...any)
}

type StdoutLogger struct {
	level int
	stage string
	file  string
}

func NewStdoutLogger(level int) StdoutLogger {
	return StdoutLogger{level: level}
}

func (log *StdoutLogger) Stage(value string) {
	log.stage = strings.TrimSpace(value)
}

func (log *StdoutLogger) File(value string) {
	log.file = strings.TrimSpace(value)
}

func (log *StdoutLogger) Info(level int, format string, args ...any) {
	if level < log.level {
		return
	}

	stdout.New(stdout.Bold).Printf("[INFO] [%s]\t\t", log.stage)
	fmt.Printf("%s\n", fmt.Sprintf(format, args...))
}

func (log *StdoutLogger) Error(level int, format string, args ...any) {
	if level < log.level {
		return
	}
	stdout.New(stdout.FgRed, stdout.Bold).Printf("[ERROR]")
	stdout.New(stdout.Bold).Printf(" [%s]\t\t", log.stage)
	fmt.Printf("%s\n", fmt.Sprintf(format, args...))
}
