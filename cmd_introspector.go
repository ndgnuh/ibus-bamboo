package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func cmdGetFocusedWmClass() string {
	cmdString := os.Getenv("BAMBOO_INTROSPECTOR_CMD")
	cmd := exec.Command(cmdString)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get wmclass from command (%s)", cmdString)
		return ""
	}
	return strings.Trim(string(out), "\r\n ")
}
