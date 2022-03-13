package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hpcloud/tail"
	"github.com/pterm/pterm"
)

var serfPath = "serf"

func main() {

	flag.StringVar(&serfPath, "serf-path", "serf", "serf-path: ./serf.exe")
	flag.Parse()

	startArgs := []string{
		"agent",
		"-config-file=./config.json",
	}
	output := bytes.NewBuffer([]byte{})

	//打开节点程序
	go func() {
		if err := Run(serfPath, startArgs, output); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}()

	//读取event日志
	go func() {
		if err := tailFile(); err != nil {
			fmt.Println(err)
			os.Exit(0)
		}

	}()

	input := bufio.NewScanner(os.Stdin)
	fmt.Printf("Please type in something:\n")
INPUT:
	for input.Scan() {
		line := strings.TrimSpace(input.Text())

		switch {
		case line == "bye":
			break INPUT
		case len(line) > 0: //转发命令给serf，如果没有匹配的命令，默认 query say xxx
			var spinnerSuccess *pterm.SpinnerPrinter
			spinnerSuccess, _ = pterm.DefaultSpinner.Start("Sending...")
			cmd_output := bytes.NewBuffer([]byte{})

			if err := Run(serfPath, strings.Split(line, " "), cmd_output); err != nil {

				cmd_output.Truncate(0)
				queryArg := []string{
					"query", "say", line,
				}

				if err := Run(serfPath, queryArg, cmd_output); err != nil {
					spinnerSuccess.Warning(cmd_output.String(), err)
					continue
				}
			}

			spinnerSuccess.Success(cmd_output.String())
		}

	}
	fmt.Println("Bye!!")
	// "while read line;do \n echo '${SERF_EVENT} from ${SERF_SELF_NAME}: ${line}'>>events.log \n done \n echo '收到'"
}

func Run(name string, args []string, output *bytes.Buffer) (err error) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = output
	cmd.Stderr = output
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: false} //linux
	err = cmd.Run()
	return
}

func tailFile() (err error) {

	t, err := tail.TailFile("./events.log", tail.Config{Follow: true})
	if err != nil {
		return
	}

	for line := range t.Lines {

		txt := line.Text
		eventSimbolIndex := strings.Index(txt, ":")
		switch {
		case eventSimbolIndex > 0 && strings.HasPrefix(txt, "query"):
			pterm.DefaultBox.
				WithBoxStyle(pterm.NewStyle(pterm.FgGreen)).
				WithTextStyle(pterm.NewStyle(pterm.FgGreen)).
				WithTitle(txt[:eventSimbolIndex+1]).
				Println(strings.TrimSpace(txt[eventSimbolIndex+1:]))
		default:
			pterm.DefaultBox.
				WithBoxStyle(pterm.NewStyle(pterm.FgDarkGray)).
				WithTextStyle(pterm.NewStyle(pterm.FgDarkGray)).
				WithTitle("event").
				Println(txt)
		}

	}
	return
}
