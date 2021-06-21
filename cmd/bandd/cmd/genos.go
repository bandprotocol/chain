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
	"github.com/bandprotocol/go-owasm/api"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// AddGenesisOracleScriptCmd returns add-oracle-script cobra Command.
func AddGenesisOracleScriptCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-oracle-script [name] [description] [schema] [url] [owner] [filepath]",
		Short: "Add a data source to genesis.json",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			config.SetRoot(clientCtx.HomeDir)

			f := filecache.New(filepath.Join(defaultNodeHome, "files"))
			data, err := ioutil.ReadFile(args[5])
			if err != nil {
				return err
			}
			vm, err := api.NewVm(0) // The compilation doesn't use cache
			if err != nil {
				return err
			}
			compiledData, err := vm.Compile(data, types.MaxCompiledWasmCodeSize)
			if err != nil {
				return err
			}
			filename := f.AddFile(compiledData)
			owner, err := sdk.AccAddressFromBech32(args[4])
			if err != nil {
				return err
			}
			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			oracleGenState := types.GetGenesisStateFromAppState(cdc, appState)
			oracleGenState.OracleScripts = append(oracleGenState.OracleScripts, types.NewOracleScript(
				owner, args[0], args[1], filename, args[2], args[3],
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
