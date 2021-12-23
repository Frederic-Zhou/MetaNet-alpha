package main

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Frederic-Zhou/MetaNet-alpha/node"

	mnapp "github.com/Frederic-Zhou/MetaNet-alpha/app"
	abciclient "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/proto/tendermint/crypto"
)

var ctx = context.Background()
var client abciclient.Client
var err error

func main() {
	node.InitFiles("validator")

	cf := abciclient.NewLocalCreator(mnapp.NewPersistentKVStoreApplication("./db"))
	go func() {
		node.Start(&cf)
	}()

	client, err = cf()
	var logger = log.MustNewDefaultLogger(log.LogFormatPlain, log.LogLevelInfo, false)
	client.SetLogger(logger.With("module", "abci-client"))

	for {
		fmt.Printf("> ")
		bufReader := bufio.NewReader(os.Stdin)
		line, more, err := bufReader.ReadLine()
		if more {
			fmt.Println("input is too long")
		} else if err != nil {
			fmt.Println(err)
		}

		pArgs := persistentArgs(line)
		// fmt.Println(pArgs)
		if err := muxOnCommands(pArgs); err != nil {
			fmt.Println(err)
		}
	}

}

func persistentArgs(line []byte) []string {

	// generate the arguments to run from original os.Args
	// to maintain flag arguments
	args := os.Args
	args = args[:len(args)-1] // remove the previous command argument

	if len(line) > 0 { // prevents introduction of extra space leading to argument parse errors
		args = append(args, strings.Split(string(line), " ")...)
	}
	return args
}

func muxOnCommands(pArgs []string) error {
	if len(pArgs) < 1 {
		return errors.New("expecting persistent args of the form: abci-cli [command] <...>")
	}

	// TODO: this parsing is fragile
	args := []string{}
	for i := 0; i < len(pArgs); i++ {
		arg := pArgs[i]

		// check for flags
		if strings.HasPrefix(arg, "-") {
			// if it has an equal, we can just skip
			if strings.Contains(arg, "=") {
				continue
			}
			// if its a boolean, we can just skip
			// _, err := cmd.Flags().GetBool(strings.TrimLeft(arg, "-"))
			// if err == nil {
			// 	continue
			// }

			// otherwise, we need to skip the next one too
			i++
			continue
		}

		// append the actual arg
		args = append(args, arg)
	}
	var subCommand string
	var actualArgs []string
	if len(args) > 0 {
		subCommand = args[0]
	}
	if len(args) > 1 {
		actualArgs = args[1:]
	}
	// cmd.Use = subCommand // for later print statements ...

	switch strings.ToLower(subCommand) {
	case "check_tx":
		return cmdCheckTx(actualArgs)
	case "commit":
		return cmdCommit(actualArgs)
	case "deliver_tx":
		return cmdDeliverTx(actualArgs)
	case "echo":
		return cmdEcho(actualArgs)
	case "info":
		return cmdInfo(actualArgs)
	case "query":
		return cmdQuery(actualArgs)
	default:
		return cmdUnimplemented(pArgs)
	}
}

// Have the application echo a message
func cmdEcho(args []string) error {
	msg := ""
	if len(args) > 0 {
		msg = args[0]
	}
	res, err := client.EchoSync(ctx, msg)
	if err != nil {
		return err
	}
	printResponse(args, response{
		Data: []byte(res.Message),
	})
	return nil
}

// Get some info from the application
func cmdInfo(args []string) error {
	var version string
	if len(args) == 1 {
		version = args[0]
	}
	res, err := client.InfoSync(ctx, types.RequestInfo{Version: version})
	if err != nil {
		return err
	}
	printResponse(args, response{
		Data: []byte(res.Data),
	})
	return nil
}

const codeBad uint32 = 10

// Append a new tx to application
func cmdDeliverTx(args []string) error {
	if len(args) == 0 {
		printResponse(args, response{
			Code: codeBad,
			Log:  "want the tx",
		})
		return nil
	}
	txBytes, err := stringOrHexToBytes(args[0])
	if err != nil {
		return err
	}
	res, err := client.DeliverTxSync(ctx, types.RequestDeliverTx{Tx: txBytes})
	if err != nil {
		return err
	}
	printResponse(args, response{
		Code: res.Code,
		Data: res.Data,
		Info: res.Info,
		Log:  res.Log,
	})
	return nil
}

// Validate a tx
func cmdCheckTx(args []string) error {
	if len(args) == 0 {
		printResponse(args, response{
			Code: codeBad,
			Info: "want the tx",
		})
		return nil
	}
	txBytes, err := stringOrHexToBytes(args[0])
	if err != nil {
		return err
	}
	res, err := client.CheckTxSync(ctx, types.RequestCheckTx{Tx: txBytes})
	if err != nil {
		return err
	}
	printResponse(args, response{
		Code: res.Code,
		Data: res.Data,
		Info: res.Info,
		Log:  res.Log,
	})
	return nil
}

// Get application Merkle root hash
func cmdCommit(args []string) error {
	res, err := client.CommitSync(ctx)
	if err != nil {
		return err
	}
	printResponse(args, response{
		Data: res.Data,
	})
	return nil
}

// Query application state
func cmdQuery(args []string) error {
	if len(args) == 0 {
		printResponse(args, response{
			Code: codeBad,
			Info: "want the query",
			Log:  "",
		})
		return nil
	}
	queryBytes, err := stringOrHexToBytes(args[0])
	if err != nil {
		return err
	}

	resQuery, err := client.QuerySync(ctx, types.RequestQuery{
		Data:   queryBytes,
		Path:   "flagPath",
		Height: int64(0),
		Prove:  false,
	})
	if err != nil {
		return err
	}
	printResponse(args, response{
		Code: resQuery.Code,
		Info: resQuery.Info,
		Log:  resQuery.Log,
		Query: &queryResponse{
			Key:      resQuery.Key,
			Value:    resQuery.Value,
			Height:   resQuery.Height,
			ProofOps: resQuery.ProofOps,
		},
	})
	return nil
}

func cmdUnimplemented(args []string) error {
	msg := "unimplemented command"

	if len(args) > 0 {
		msg += fmt.Sprintf(" args: [%s]", strings.Join(args, " "))
	}
	printResponse(args, response{
		Code: codeBad,
		Log:  msg,
	})

	fmt.Println("Available commands:")

	fmt.Println("Use \"[command] --help\" for more information about a command.")

	return nil
}

func printResponse(args []string, rsp response) {

	// if flagVerbose {
	// 	fmt.Println(">", cmd.Use, strings.Join(args, " "))
	// }

	// Always print the status code.
	if rsp.Code == types.CodeTypeOK {
		fmt.Printf("-> code: OK\n")
	} else {
		fmt.Printf("-> code: %d\n", rsp.Code)

	}

	if len(rsp.Data) != 0 {
		// Do no print this line when using the commit command
		// because the string comes out as gibberish

		fmt.Printf("-> data: %s\n", rsp.Data)

		fmt.Printf("-> data.hex: 0x%X\n", rsp.Data)
	}
	if rsp.Log != "" {
		fmt.Printf("-> log: %s\n", rsp.Log)
	}

	if rsp.Query != nil {
		fmt.Printf("-> height: %d\n", rsp.Query.Height)
		if rsp.Query.Key != nil {
			fmt.Printf("-> key: %s\n", rsp.Query.Key)
			fmt.Printf("-> key.hex: %X\n", rsp.Query.Key)
		}
		if rsp.Query.Value != nil {
			fmt.Printf("-> value: %s\n", rsp.Query.Value)
			fmt.Printf("-> value.hex: %X\n", rsp.Query.Value)
		}
		if rsp.Query.ProofOps != nil {
			fmt.Printf("-> proof: %#v\n", rsp.Query.ProofOps)
		}
	}
}

// NOTE: s is interpreted as a string unless prefixed with 0x
func stringOrHexToBytes(s string) ([]byte, error) {
	if len(s) > 2 && strings.ToLower(s[:2]) == "0x" {
		b, err := hex.DecodeString(s[2:])
		if err != nil {
			err = fmt.Errorf("error decoding hex argument: %s", err.Error())
			return nil, err
		}
		return b, nil
	}

	if !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
		err := fmt.Errorf("invalid string arg: \"%s\". Must be quoted or a \"0x\"-prefixed hex string", s)
		return nil, err
	}

	return []byte(s[1 : len(s)-1]), nil
}

type response struct {
	// generic abci response
	Data []byte
	Code uint32
	Info string
	Log  string

	Query *queryResponse
}

type queryResponse struct {
	Key      []byte
	Value    []byte
	Height   int64
	ProofOps *crypto.ProofOps
}
