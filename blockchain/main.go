package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Frederic-Zhou/MetaNet-alpha/blockchain/node"
	// localRpcClient "github.com/tendermint/tendermint/rpc/client/local"
)

func main_for_tendermint() {
	err := node.InitConfig("validator")
	if err != nil {
		os.Exit(0)
	}

	err = node.InitNode()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	go func() {
		if err := node.Start(); err != nil {
			fmt.Println("over with error:", err)
		}
	}()

	client, err := node.GetClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	for {
		client.BroadcastTxCommit(context.Background(), []byte("abc="+string(time.Now().Second())))
		time.Sleep(10 * time.Second)
	}

}
