package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	captypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	v040 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v040"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibcxfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	ibccoretypes "github.com/cosmos/cosmos-sdk/x/ibc/core/types"
	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	v039oracle "github.com/bandprotocol/chain/v2/x/oracle/legacy/v039"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	flagGenesisTime   = "genesis-time"
	flagChainID       = "chain-id"
	flagInitialHeight = "initial-height"
)

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*tmtypes.GenesisDoc, error) {
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read GenesisDoc file: %w", err)
	}
	genDoc, err := tmtypes.GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return nil, fmt.Errorf("error reading GenesisDoc at %s: %w", genDocFile, err)
	}

	genDoc.ConsensusParams.Block.MaxBytes = 1000000 // 1M bytes
	genDoc.ConsensusParams.Block.MaxGas = 40000000  // 40M gas
	genDoc.ConsensusParams.Block.TimeIotaMs = 1000  // 1 second

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return genDoc, nil
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
// nolint: funlen
func MigrateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate [genesis-file]",
		Short: "Migrate Guanyu version to Laozi version",
		Long: fmt.Sprintf(`Migrate the the Guanyu version into Laozi version and print to STDOUT.

Example:
$ %s migrate /path/to/genesis.json --chain-id=band-laozi --genesis-time=2020-08-11T17:00:00Z
`, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			var err error

			importGenesis := args[0]
			genDoc, err := GenesisDocFromFile(importGenesis)
			if err != nil {
				return errors.Wrapf(err, "failed to read genesis document from file %s", importGenesis)
			}

			var initialState genutiltypes.AppMap
			if err := json.Unmarshal(genDoc.AppState, &initialState); err != nil {
				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
			}

			// Migrate from guanyu (0.39 like genesis file) to cosmos-sdk v0.40
			newGenState := v040.Migrate(initialState, clientCtx)

			ibcTransferGenesis := ibcxfertypes.DefaultGenesisState()
			ibcCoreGenesis := ibccoretypes.DefaultGenesisState()
			capGenesis := captypes.DefaultGenesis()
			oracleGenesis := oracletypes.DefaultGenesisState()

			ibcTransferGenesis.Params.ReceiveEnabled = false
			ibcTransferGenesis.Params.SendEnabled = false

			newGenState[ibcxfertypes.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(ibcTransferGenesis)
			newGenState[host.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(ibcCoreGenesis)
			newGenState[captypes.ModuleName] = clientCtx.JSONMarshaler.MustMarshalJSON(capGenesis)

			v039Codec := codec.NewLegacyAmino()
			v040Codec := clientCtx.JSONMarshaler
			var oracleGenesisV039 v039oracle.GenesisState
			v039Codec.MustUnmarshalJSON(initialState[oracletypes.ModuleName], &oracleGenesisV039)

			oracleGenesis.Params.IBCRequestEnabled = false
			oracleGenesis.OracleScripts = oracleGenesisV039.OracleScripts
			oracleGenesis.Reporters = oracleGenesisV039.Reporters
			for _, dataSource := range oracleGenesisV039.DataSources {
				oracleGenesis.DataSources = append(oracleGenesis.DataSources, oracletypes.DataSource{
					Owner:       dataSource.Owner,
					Name:        dataSource.Name,
					Description: dataSource.Description,
					Filename:    dataSource.Filename,
					Treasury:    dataSource.Owner,
					Fee:         sdk.NewCoins(),
				})
			}
			newGenState[oracletypes.ModuleName] = v040Codec.MustMarshalJSON(oracleGenesis)

			genDoc.AppState, err = json.Marshal(newGenState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}

			genesisTime, _ := cmd.Flags().GetString(flagGenesisTime)
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal genesis time")
				}

				genDoc.GenesisTime = t
			}

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID != "" {
				genDoc.ChainID = chainID
			}

			initialHeight, _ := cmd.Flags().GetInt(flagInitialHeight)
			genDoc.InitialHeight = int64(initialHeight)

			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}

			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))
			return nil
		},
	}
	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().String(flagChainID, "band-laozi", "override chain_id with this flag")
	cmd.Flags().Int(flagInitialHeight, 0, "Set the starting height for the chain")

	return cmd
}
