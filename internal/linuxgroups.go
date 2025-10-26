package internal

import (
	"os/exec"
)

func CreateLinuxGroup(group string) (bool, error) {
	cmd := exec.Command("groupadd", group)
	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok {
		switch exitErr.ExitCode() {
		case 0:
			return false, nil
		case 9:
			return true, nil
		default:
			return false, err
		}
	} else if err != nil {
		return false, err
	}
	return false, nil
}
