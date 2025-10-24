package internal

import (
	"fmt"
	"strings"
	"time"

	stdout "github.com/fatih/color"
)

type Logger interface {
	Log(stmt *LogStatement)
}

type LogStatement struct {
	logger Logger
	stage  string

	level     int
	timestamp time.Time
	message   string
	cause     *CLErr
	args      map[string]any

	isErr bool
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

func (log *Log) Error(error *CLErr) *LogStatement {
	stmt := LogStatement{
		logger:    log.logger,
		stage:     log.stage,
		level:     10,
		message:   error.Error(),
		cause:     error.Parent,
		timestamp: time.Now().UTC(),
		isErr:     true,
	}
	return &stmt
}

func (log *Log) Info(level int, format string, args ...any) *LogStatement {
	stmt := LogStatement{
		logger:    log.logger,
		stage:     log.stage,
		level:     10,
		message:   fmt.Sprintf(format, args...),
		cause:     nil,
		timestamp: time.Now().UTC(),
		isErr:     true,
	}
	return &stmt
}

func (stmt *LogStatement) With(key string, value any) *LogStatement {
	if stmt.args == nil {
		stmt.args = make(map[string]any)
	}
	stmt.args[key] = value
	return stmt
}

func (stmt *LogStatement) Done() {
	stmt.logger.Log(stmt)
}

func (stmt *LogStatement) Info() {
	stmt.logger.Log(stmt)
}

func (stmt *LogStatement) Error() {
	stmt.logger.Log(stmt)
}

type MultiLogger struct {
	loggers []Logger
}

func NewMultiLogger(loggers ...Logger) *MultiLogger {
	return &MultiLogger{loggers: loggers}
}

func (log *MultiLogger) Log(stmt *LogStatement) {
	for _, logger := range log.loggers {
		logger.Log(stmt)
	}
}

type StdoutLogger struct {
	level int
}

func NewStdoutLogger(level int) *StdoutLogger {
	return &StdoutLogger{level: level}
}

func (log *StdoutLogger) Log(stmt *LogStatement) {
	if stmt.isErr {
		log.Error(stmt)
	} else {
		log.Info(stmt)
	}
}

func (log *StdoutLogger) Info(stmt *LogStatement) {
	if stmt.level < log.level {
		return
	}

	stdout.New(stdout.Bold).Printf("[INFO] [%s]", stmt.stage)
	fmt.Println(stmt.message)
	for key, value := range stmt.args {
		fmt.Printf("\t\t%s: %v\n", key, value)
	}
}

func (log *StdoutLogger) Error(stmt *LogStatement) {
	if stmt.level < log.level {
		return
	}

	stdout.New(stdout.FgRed, stdout.Bold).Printf("[ERROR]")
	stdout.New(stdout.Bold).Printf("[%s]", stmt.stage)
	fmt.Println(stmt.message)
	for key, value := range stmt.args {
		fmt.Printf("\t\t%s: %v\n", key, value)
	}
}
