package daemon

import "errors"

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

func (cmd *Cmd) IntArg(name string) (int, bool, error) {
	value, ok := cmd.Args[name]
	if !ok {
		return 0, false, nil
	}
	if value == nil {
		return 0, false, nil
	}

	val, ok := value.(int)
	if ok {
		return val, true, nil
	}
	return 0, false, ErrNotIntArg
}

type Log struct {
	Value string
}

type End struct {
	Ok    bool
	Value any
}
