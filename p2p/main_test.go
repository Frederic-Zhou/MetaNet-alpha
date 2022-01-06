package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	serfClient "github.com/hashicorp/serf/client"
)

func main_for_serf() {

	output := bytes.NewBuffer([]byte{})
	//启动serf
	go func() {
		err := startSerf(output)
		fmt.Println(err)
	}()

	//打印输出
	go func() {
		for {
			out, err := ioutil.ReadAll(output)
			if err != nil {
				return
			}
			if len(out) > 0 {
				fmt.Println(string(out))
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	//接受命令
	time.Sleep(1 * time.Second)
	client, err := newSerfClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	input := bufio.NewScanner(os.Stdin)
	fmt.Printf("Please type in something:\n")
	// 逐行扫描
	for input.Scan() {
		line := input.Text()
		commands := strings.Split(line, " ")

		switch {
		case commands[0] == "bye":
			return
		case commands[0] == "join":
			if len(commands) > 1 {
				client.Join(commands[1:], false)
			}

		case commands[0] == "query":

			ackCh := make(chan string, 128)
			respCh := make(chan serfClient.NodeResponse, 128)

			params := serfClient.QueryParam{
				RequestAck:  false,
				RelayFactor: uint8(0),
				Timeout:     0,
				Name:        commands[1],
				Payload:     []byte(commands[2]),
				AckCh:       ackCh,
				RespCh:      respCh,
			}

			err := client.Query(&params)
			if err != nil {
				fmt.Println(err)
			}
		case commands[0] == "event":

			if len(commands) > 2 {
				err := client.UserEvent(commands[1], []byte(commands[2]), false)
				if err != nil {
					fmt.Println(err)
				}
			}

		case commands[0] == "leave":
			err := client.Leave()
			if err != nil {
				fmt.Println(err)
			}
		case commands[0] == "members":
			members, err := client.Members()
			if err != nil {
				fmt.Println(err)
			}

			for _, m := range members {
				fmt.Println(m)
			}

		default:
			fmt.Println("nothing to do ...")

		}

	}

}

func startSerf(output *bytes.Buffer) (err error) {
	args := []string{
		"agent",
		"-node=node1",
		"-bind=127.0.0.1:5000",
		"-rpc-addr=127.0.0.1:7373",
	}

	cmd := exec.Command("/usr/local/bin/serf", args...)
	cmd.Stdout = output
	cmd.Stderr = output

	err = cmd.Run()

	return
}

func newSerfClient() (client *serfClient.RPCClient, err error) {

	client, err = serfClient.NewRPCClient("127.0.0.1:7373")

	return
}
