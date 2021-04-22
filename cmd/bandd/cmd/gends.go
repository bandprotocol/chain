package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/GeoDB-Limited/odin-core/pkg/filecache"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

const (
	feeFlag = "fee"
)

// AddGenesisDataSourceCmd returns add-data-source cobra Command.
func AddGenesisDataSourceCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-data-source [name] [description] [owner] [filepath] (--fee [fee])",
		Short: "Add a data source to genesis.json",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			name := args[0]
			description := args[1]

			f := filecache.New(filepath.Join(defaultNodeHome, "files"))
			data, err := ioutil.ReadFile(args[3])
			if err != nil {
				return err
			}
			filename := f.AddFile(data)
			owner, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}
			coinsRaw, err := cmd.Flags().GetString(feeFlag)
			if err != nil {
				return err
			}

			var coins sdk.Coins
			if coinsRaw == "" {
				coins = sdk.NewCoins()
			} else {
				coins, err = sdk.ParseCoinsNormalized(coinsRaw)
				if err != nil {
					return err
				}
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			oracleGenState := types.GetGenesisStateFromAppState(cdc, appState)

			oracleGenState.DataSources = append(oracleGenState.DataSources, types.NewDataSource(
				owner, name, description, filename, coins,
			))

			oracleGenStateBz, err := cdc.MarshalJSON(oracleGenState)

			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}
			appState[types.ModuleName] = oracleGenStateBz

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}
	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(feeFlag, "", "fee data requesters should pay in data providers pool")
	return cmd
}
