package internal

import (
	"fmt"
	"testing"
)

func TestCLerr(t *testing.T) {
	var err error
	err = ErrOf(Err("invalid target name"), "can not parse index file")
	fmt.Println(err.Error())
}
