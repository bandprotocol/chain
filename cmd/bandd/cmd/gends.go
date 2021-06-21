package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/bandprotocol/chain/v2/pkg/filecache"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// AddGenesisDataSourceCmd returns add-data-source cobra Command.
func AddGenesisDataSourceCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-data-source [name] [description] [owner] [filepath] [fee] [treasury]",
		Short: "Add a data source to genesis.json",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

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
			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}
			fee, err := sdk.ParseCoinsNormalized(args[4])
			if err != nil {
				return err
			}
			treasury, err := sdk.AccAddressFromBech32(args[5])
			if err != nil {
				return err
			}

			oracleGenState := types.GetGenesisStateFromAppState(cdc, appState)
			oracleGenState.DataSources = append(oracleGenState.DataSources, types.NewDataSource(
				owner, args[0], args[1], filename, fee, treasury,
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
	return cmd
}
