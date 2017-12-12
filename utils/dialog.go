package utils

import (
	"os/exec"
)

// Dial creates UI dialog with two buttons "Yes" and "No".
// Selecting "Yes" provides returning nil value, otherwise
// 1 is returned.
func Dial(message string) error {

	command := exec.Command("zenity", "--question", "--text=\""+message+"\"")
	return command.Run()
}
