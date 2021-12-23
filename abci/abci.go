package abci

import (
	"context"
	"fmt"
	"os/signal"

	"syscall"

	abciclient "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/abci/example/kvstore"
	"github.com/tendermint/tendermint/abci/server"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	logger log.Logger        = log.MustNewDefaultLogger(log.LogFormatPlain, log.LogLevelInfo, false)
	app    types.Application = kvstore.NewApplication()
)

func NewClient() (*abciclient.Client, error) {
	var client abciclient.Client
	var err error
	if logger == nil {
		logger = log.MustNewDefaultLogger(log.LogFormatPlain, log.LogLevelInfo, false)
	}

	client, _ = abciclient.NewLocalCreator(app)()

	if err = client.Start(); err != nil {
		return nil, err
	}
	defer client.Stop()
	client.SetLogger(logger.With("module", "abci-client"))
	return &client, nil
}

func RunAsSocketServer() error {
	// Start the listener
	srv, err := server.NewServer("tcp://0.0.0.0:26658", "socket", app)
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	fmt.Println(logger)

	srv.SetLogger(logger.With("module", "abci-server"))

	if err := srv.Start(); err != nil {
		return err
	}
	defer srv.Stop()

	// Run forever.
	<-ctx.Done()
	return nil
}
