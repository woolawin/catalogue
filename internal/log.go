package internal

import (
	"fmt"
	"strings"
	"time"

	stdout "github.com/fatih/color"
)

type Logger interface {
	Info(stmt *LogStatement)
	Error(stmt *LogStatement)
}

type LogStatement struct {
	logger Logger
	stage  string

	level     int
	timestamp time.Time
	message   string
	args      map[string]any
}

type Log struct {
	stage  string
	logger Logger
}

func NewLog(logger Logger) *Log {
	return &Log{logger: logger}
}

func (log *Log) Stage(stage string) {
	log.stage = strings.TrimSpace(stage)
}

func (log *Log) Msg(level int, msg string) *LogStatement {
	stmt := LogStatement{logger: log.logger, stage: log.stage, level: level, message: msg, timestamp: time.Now().UTC()}
	return &stmt
}

func (stmt *LogStatement) With(key string, value any) *LogStatement {
	if stmt.args == nil {
		stmt.args = make(map[string]any)
	}
	stmt.args[key] = value
	return stmt
}

func (stmt *LogStatement) Info() {
	stmt.logger.Info(stmt)
}

func (stmt *LogStatement) Error() {
	stmt.logger.Error(stmt)
}

type StdoutLogger struct {
	level int
}

func NewStdoutLogger(level int) *StdoutLogger {
	return &StdoutLogger{level: level}
}

func (log *StdoutLogger) Info(stmt *LogStatement) {
	if stmt.level < log.level {
		return
	}

	stdout.New(stdout.Bold).Printf("[INFO] [%s]\t\t", stmt.stage)
	fmt.Println(stmt.message)
	for key, value := range stmt.args {
		fmt.Printf("\t\t\t\t%s: %v\n", key, value)
	}
}

func (log *StdoutLogger) Error(stmt *LogStatement) {
	if stmt.level < log.level {
		return
	}

	stdout.New(stdout.FgRed, stdout.Bold).Printf("[ERROR]")
	stdout.New(stdout.Bold).Printf("[%s]\t\t", stmt.stage)
	fmt.Println(stmt.message)
	for key, value := range stmt.args {
		fmt.Printf("\t\t\t\t%s: %v\n", key, value)
	}
}
