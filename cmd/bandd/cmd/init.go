package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/cli"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
)

const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagRecover defines a flag to initialize the private validator key from a specific seed.
	FlagRecover = "recover"

	// FlagTimeoutCommit defines a flag to set timeout commit of node.
	FlagTimeoutCommit = "timeout-commit"
)

type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", "")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(os.Stderr, "%s\n", string(sdk.MustSortJSON(out)))
	return err
}

func genFilePVIfNotExists(keyFilePath, stateFilePath string) error {
	if !tmos.FileExists(keyFilePath) {
		pv := privval.NewFilePV(secp256k1.GenPrivKey(), keyFilePath, stateFilePath)
		// privKey :=

		// pv := &privval.FilePV{
		// 	Key: privval.FilePVKey{
		// 		Address: privKey.PubKey().Address(),
		// 		PubKey:  privKey.PubKey(),
		// 		PrivKey: privKey,

		// 	},
		// 	LastSignState: privval.FilePVLastSignState{
		// 		Step: 0,
		// 	},
		// }
		pv.Save()
		// jsonBytes, err := json.MarshalIndent(pv.Key, "", "  ")
		// if err != nil {
		// 	return err
		// }
		// if err = tempfile.WriteFileAtomic(keyFilePath, jsonBytes, 0600); err != nil {
		// 	return err
		// }
		// jsonBytes, err = json.MarshalIndent(pv.LastSignState, "", "  ")
		// if err != nil {
		// 	return err
		// }
		// if err = tempfile.WriteFileAtomic(stateFilePath, jsonBytes, 0600); err != nil {
		// 	return err
		// }
	}
	return nil
}

// InitCmd returns a command that initializes all files needed for Tendermint and BandChain app.
func InitCmd(customAppState map[string]json.RawMessage, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)
			timeoutCommit, err := cmd.Flags().GetInt(FlagTimeoutCommit)
			if err != nil {
				return err
			}
			config.Consensus.TimeoutCommit = time.Duration(timeoutCommit) * time.Millisecond
			if err := genFilePVIfNotExists(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile()); err != nil {
				return err
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", tmrand.Str(6))
			}

			// Get bip39 mnemonic
			var mnemonic string
			recover, _ := cmd.Flags().GetBool(FlagRecover)
			if recover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				mnemonic, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			config.Moniker = args[0]

			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

			if !overwrite && tmos.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}
			appState, err := json.MarshalIndent(customAppState, "", "")
			if err != nil {
				return err
			}
			genDoc := &types.GenesisDoc{}
			if _, err := os.Stat(genFile); err != nil {
				if !os.IsNotExist(err) {
					return err
				}
			} else {
				genDoc, err = types.GenesisDocFromFile(genFile)
				if err != nil {
					return err
				}
			}
			genDoc.ChainID = chainID
			genDoc.Validators = nil
			genDoc.AppState = appState
			genDoc.ConsensusParams = types.DefaultConsensusParams()
			// TODO: Revisit max block size
			// genDoc.ConsensusParams.Block.MaxBytes = 1000000 // 1M bytes
			genDoc.ConsensusParams.Block.MaxGas = 40000000 // 40M gas
			genDoc.ConsensusParams.Block.TimeIotaMs = 1000 // 1 second
			genDoc.ConsensusParams.Validator.PubKeyTypes = []string{types.ABCIPubKeyTypeSecp256k1}
			if err = genutil.ExportGenesisFile(genDoc, genFile); err != nil {
				return err
			}
			toPrint := newPrintInfo(config.Moniker, chainID, nodeID, "", appState)
			cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
			return displayInfo(toPrint)
		},
	}
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Int(FlagTimeoutCommit, 1500, "timeout commit of the node in ms")
	return cmd
}
