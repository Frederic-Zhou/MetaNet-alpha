package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func main() {
	startArgs := []string{
		"agent",
		"-config-file=./config.json",
	}
	output := bytes.NewBuffer([]byte{})

	go func() {
		if err := Run("serf", startArgs, output); err != nil {
			fmt.Println(err)
			return
		}
	}()

	input := bufio.NewScanner(os.Stdin)
	fmt.Printf("Please type in something:\n")
INPUT:
	for input.Scan() {
		line := input.Text()

		switch {
		case line == "bye":
			break INPUT
		default:
			cmd_output := bytes.NewBuffer([]byte{})
			if err := Run("serf", strings.Split(line, " "), cmd_output); err != nil {
				fmt.Println("*****:", cmd_output.String(), err)
				continue
			}
			fmt.Println(">>>>>:", cmd_output.String())
		}

	}
	fmt.Println("Bye!!")
}

func Run(name string, args []string, output *bytes.Buffer) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: false}
	err = cmd.Run()
	return
}
