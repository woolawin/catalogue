package daemon

import (
	"errors"

	"github.com/woolawin/catalogue/internal"
)

var ErrNotStringArg = errors.New("not a string argument")
var ErrNotIntArg = errors.New("not a int argument")

type Message struct {
	Cmd *Cmd
	Log *Log
	End *End
}

type Command int

const (
	Add          Command = 1
	HasPackage   Command = 2
	ListPackages Command = 3
	Update       Command = 4
)

type Cmd struct {
	Command Command
	Args    map[string]any
}

func (cmd *Cmd) StringArg(name string) (string, bool, error) {
	value, ok := cmd.Args[name]
	if !ok {
		return "", false, nil
	}
	if value == nil {
		return "", true, nil
	}

	val, ok := value.(string)
	if ok {
		return val, true, nil
	}
	return "", false, ErrNotStringArg
}

func (cmd *Cmd) IntArg(name string) (int, bool, any, error) {
	value, ok := cmd.Args[name]
	if !ok {
		return 0, false, value, nil
	}
	if value == nil {
		return 0, false, value, nil
	}
	val8, ok := value.(int8)
	if ok {
		return int(val8), true, value, nil
	}

	val16, ok := value.(int16)
	if ok {
		return int(val16), true, value, nil
	}

	val32, ok := value.(int32)
	if ok {
		return int(val32), true, value, nil
	}

	val64, ok := value.(int64)
	if ok {
		return int(val64), true, value, nil
	}

	return 0, false, value, ErrNotIntArg
}

type Log struct {
	Statement *internal.LogStatement
}

type End struct {
	Ok    bool
	Value any
}
