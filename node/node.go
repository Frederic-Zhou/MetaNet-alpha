package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	abciclient "github.com/tendermint/tendermint/abci/client"
	"github.com/tendermint/tendermint/libs/service"

	"github.com/tendermint/tendermint/libs/log"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmtime "github.com/tendermint/tendermint/libs/time"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"

	mnapp "github.com/Frederic-Zhou/MetaNet-alpha/app"
	cfg "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmnode "github.com/tendermint/tendermint/node"
	rpclocal "github.com/tendermint/tendermint/rpc/client/local"
)

var (
	keyType     = types.ABCIPubKeyTypeEd25519 //目前没有作用
	config      *cfg.Config
	logger      log.Logger
	ctxTimeout  = 4 * time.Second
	genesisHash []byte //目前没有作用
	node        service.Service
	client      *rpclocal.Local
)

func InitConfig(nodeType string) (err error) {

	config, err = parseConfig()
	if err != nil {
		return err
	}
	logger = log.MustNewDefaultLogger(config.LogFormat, config.LogLevel, false).With("module", "main")
	config.Mode = nodeType   //"指定节点类型: one of [validator|full|seed]"
	config.RPC.Unsafe = true //这样设置可以使用接口Dial_peers加入节点
	config.Consensus.CreateEmptyBlocks = false

	return initFilesWithConfig(config)
}

func InitNode() (err error) {
	if err = checkGenesisHash(config); err != nil {
		return
	}

	app := mnapp.NewPersistentKVStoreApplication("./db")
	cf := abciclient.NewLocalCreator(app)

	node, err = tmnode.New(config, logger, cf, nil)
	if err != nil {
		err = fmt.Errorf("failed to create node: %w", err)
		return
	}

	return
}

func GetClient() (*rpclocal.Local, error) {

	var err error
	if client == nil {
		ns, ok := node.(rpclocal.NodeService)
		if !ok {
			err = fmt.Errorf("failed asset Node to NodeService")
			return client, err
		}
		client, err = rpclocal.New(ns)
	}

	return client, err
}

func Start() error {

	if err := node.Start(); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	logger.Info("started node", "node", node.String())

	// Stop upon receiving SIGTERM or CTRL-C.
	tmos.TrapSignal(logger, func() {
		if node.IsRunning() {
			if err := node.Stop(); err != nil {
				logger.Error("unable to stop the node", "error", err)
			}
		}
	})

	// Run forever.
	select {}

}

func checkGenesisHash(config *cfg.Config) error {
	if len(genesisHash) == 0 || config.Genesis == "" {
		return nil
	}

	// Calculate SHA-256 hash of the genesis file.
	f, err := os.Open(config.GenesisFile())
	if err != nil {
		return fmt.Errorf("can't open genesis file: %w", err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("error when hashing genesis file: %w", err)
	}
	actualHash := h.Sum(nil)

	// Compare with the flag.
	if !bytes.Equal(genesisHash, actualHash) {
		return fmt.Errorf(
			"--genesis-hash=%X does not match %s hash: %X",
			genesisHash, config.GenesisFile(), actualHash)
	}

	return nil
}

func initFilesWithConfig(config *cfg.Config) error {
	var (
		pv  *privval.FilePV
		err error
	)

	if config.Mode == cfg.ModeValidator {
		// private validator
		privValKeyFile := config.PrivValidator.KeyFile()
		privValStateFile := config.PrivValidator.StateFile()
		if tmos.FileExists(privValKeyFile) {
			pv, err = privval.LoadFilePV(privValKeyFile, privValStateFile)
			if err != nil {
				return err
			}

			logger.Info("Found private validator", "keyFile", privValKeyFile,
				"stateFile", privValStateFile)
		} else {
			pv, err = privval.GenFilePV(privValKeyFile, privValStateFile, keyType)
			if err != nil {
				return err
			}
			pv.Save()
			logger.Info("Generated private validator", "keyFile", privValKeyFile,
				"stateFile", privValStateFile)
		}
	}

	nodeKeyFile := config.NodeKeyFile()
	if tmos.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := types.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if tmos.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {

		genDoc := types.GenesisDoc{
			ChainID:         fmt.Sprintf("metanet-chain-%v", tmrand.Str(6)),
			GenesisTime:     tmtime.Now(),
			ConsensusParams: types.DefaultConsensusParams(),
		}

		if keyType == types.ABCIPubKeyTypeSecp256k1 {
			genDoc.ConsensusParams.Validator = types.ValidatorParams{
				PubKeyTypes: []string{types.ABCIPubKeyTypeSecp256k1},
			}
		}

		ctx, cancel := context.WithTimeout(context.TODO(), ctxTimeout)
		defer cancel()

		// if this is a validator we add it to genesis
		if pv != nil {
			pubKey, err := pv.GetPubKey(ctx)
			if err != nil {
				return fmt.Errorf("can't get pubkey: %w", err)
			}
			genDoc.Validators = []types.GenesisValidator{{
				Address: pubKey.Address(),
				PubKey:  pubKey,
				Power:   10,
			}}
		}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	// write config file
	if err := cfg.WriteConfigFile(config.RootDir, config); err != nil {
		return err
	}
	logger.Info("Generated config", "mode", config.Mode)

	return nil
}

func parseConfig() (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	conf.SetRoot(os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	cfg.EnsureRoot(conf.RootDir)
	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}

	return conf, nil
}
