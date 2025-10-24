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
	Logger Logger
	Stage  string

	Level     int
	Timestamp time.Time
	Message   string
	Cause     *CLErr
	Args      map[string]any

	IsErr bool
}

func NewLogStatement(stage string, level int, timestamp time.Time, message string, args map[string]any, isErr bool) LogStatement {
	return LogStatement{Stage: stage, Level: level, Timestamp: timestamp, Message: message, Args: args, IsErr: isErr}
}

type Log struct {
	stage  string
	logger Logger
}

func NewLog(logger Logger) *Log {
	return &Log{logger: logger}
}

func (log *Log) Stage(stage string) func() {
	prev := log.stage
	log.stage = strings.TrimSpace(stage)
	return func() {
		log.stage = prev
	}
}

func (log *Log) Msg(level int, msg string) *LogStatement {
	stmt := LogStatement{Logger: log.logger, Stage: log.stage, Level: level, Message: msg, Timestamp: time.Now().UTC()}
	return &stmt
}

func (log *Log) Err(cause error, format string, args ...any) *LogStatement {
	error := ErrOf(cause, fmt.Sprintf(format, args...))
	stmt := LogStatement{
		Logger:    log.logger,
		Stage:     log.stage,
		Level:     10,
		Message:   error.Error(),
		Cause:     error.Parent,
		Timestamp: time.Now().UTC(),
		IsErr:     true,
	}
	return &stmt
}

func (log *Log) Info(level int, format string, args ...any) *LogStatement {
	stmt := LogStatement{
		Logger:    log.logger,
		Stage:     log.stage,
		Level:     10,
		Message:   fmt.Sprintf(format, args...),
		Cause:     nil,
		Timestamp: time.Now().UTC(),
		IsErr:     false,
	}
	return &stmt
}

func (stmt *LogStatement) With(key string, value any) *LogStatement {
	if stmt.Args == nil {
		stmt.Args = make(map[string]any)
	}
	stmt.Args[key] = value
	return stmt
}

func (stmt *LogStatement) Done() {
	stmt.Logger.Log(stmt)
}

func (stmt *LogStatement) Info() {
	stmt.Logger.Log(stmt)
}

func (stmt *LogStatement) Error() {
	stmt.Logger.Log(stmt)
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
	if stmt.IsErr {
		log.Error(stmt)
	} else {
		log.Info(stmt)
	}
}

func (log *StdoutLogger) Info(stmt *LogStatement) {
	if stmt.Level < log.level {
		return
	}

	stdout.New(stdout.Bold).Printf("[INFO] [%s] ", stmt.Stage)
	fmt.Println(stmt.Message)
	for key, value := range stmt.Args {
		fmt.Printf("\t\t%s: %v\n", key, value)
	}
}

func (log *StdoutLogger) Error(stmt *LogStatement) {
	if stmt.Level < log.level {
		return
	}

	stdout.New(stdout.FgRed, stdout.Bold).Printf("[ERROR]")
	stdout.New(stdout.Bold).Printf("[%s]", stmt.Stage)
	fmt.Println(stmt.Message)
	for key, value := range stmt.Args {
		fmt.Printf("\t\t%s: %v\n", key, value)
	}
}
