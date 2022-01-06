package main

import (
	"bytes"
	"os/exec"
)

func main() {

	args := []string{
		"agent",
		"-node=node1",
		"-bind=127.0.0.1:5000",
		"-rpc-addr=127.0.0.1:7373",
	}
	output := bytes.NewBuffer([]byte{})
	serfRun(args, output)

}

func serfRun(args []string, output *bytes.Buffer) (err error) {

	cmd := exec.Command("/usr/local/bin/serf", args...)
	cmd.Stdout = output
	cmd.Stderr = output

	err = cmd.Run()

	return
}
