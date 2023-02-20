package band

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"
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

	genDoc.ConsensusParams.Block.MaxBytes = 3000000 // 3M bytes
	genDoc.ConsensusParams.Block.MaxGas = 50000000  // 50M gas
	genDoc.ConsensusParams.Block.TimeIotaMs = 1000  // 1 second

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return genDoc, nil
}

// MigrateGenesisCmd returns a command to execute genesis state migration.
// nolint: funlen
// TODO: implement for support cosmos-sdk (v0.46.10)
func MigrateGenesisCmd() *cobra.Command {
	// 	cmd := &cobra.Command{
	// 		Use:   "migrate [genesis-file]",
	// 		Short: "Migrate Guanyu version to Laozi version",
	// 		Long: fmt.Sprintf(`Migrate the the Guanyu version into Laozi version and print to STDOUT.

	// Example:
	// $ %s migrate /path/to/genesis.json --chain-id=band-laozi --genesis-time=2020-08-11T17:00:00Z
	// `, version.AppName),
	// 		Args: cobra.ExactArgs(1),
	// 		RunE: func(cmd *cobra.Command, args []string) error {
	// 			clientCtx := client.GetClientContextFromCmd(cmd)

	// 			var err error

	// 			importGenesis := args[0]
	// 			genDoc, err := GenesisDocFromFile(importGenesis)
	// 			if err != nil {
	// 				return errors.Wrapf(err, "failed to read genesis document from file %s", importGenesis)
	// 			}

	// 			var initialState genutiltypes.AppMap
	// 			if err := json.Unmarshal(genDoc.AppState, &initialState); err != nil {
	// 				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
	// 			}

	// 			genesisTimeStr, _ := cmd.Flags().GetString(flagGenesisTime)
	// 			genesisTime := genDoc.GenesisTime
	// 			if genesisTimeStr != "" {
	// 				err := genesisTime.UnmarshalText([]byte(genesisTimeStr))
	// 				if err != nil {
	// 					return errors.Wrap(err, "failed to unmarshal genesis time")
	// 				}

	// 				genDoc.GenesisTime = genesisTime
	// 			}

	// 			// Get Guanyu oracle genesis state
	// 			v039Codec := codec.NewLegacyAmino()
	// 			v043Codec := clientCtx.Codec
	// 			var oracleGenesisV039 v039oracle.GenesisState
	// 			v039Codec.MustUnmarshalJSON(initialState[oracletypes.ModuleName], &oracleGenesisV039)

	// 			// Migrate from guanyu (0.39 like genesis file) to cosmos-sdk v0.43
	// 			newGenState := v043.Migrate(v040.Migrate(initialState, clientCtx), clientCtx)

	// 			// Add new module genesis state
	// 			ibcCoreGenesis := ibccoretypes.DefaultGenesisState()
	// 			newGenState[host.ModuleName] = clientCtx.Codec.MustMarshalJSON(ibcCoreGenesis)

	// 			capGenesis := captypes.DefaultGenesis()
	// 			newGenState[captypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(capGenesis)

	// 			ibcTransferGenesis := ibcxfertypes.DefaultGenesisState()
	// 			ibcTransferGenesis.Params.ReceiveEnabled = false
	// 			ibcTransferGenesis.Params.SendEnabled = false
	// 			newGenState[ibcxfertypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(ibcTransferGenesis)

	// 			feegrantGenesis := feegrant.DefaultGenesisState()
	// 			newGenState[feegrant.ModuleName] = clientCtx.Codec.MustMarshalJSON(feegrantGenesis)

	// 			// Adjust distribute params BaseProposer/Bonus to 3/12 %
	// 			var distrGenesis distrtypes.GenesisState
	// 			clientCtx.Codec.MustUnmarshalJSON(newGenState[distrtypes.ModuleName], &distrGenesis)
	// 			distrGenesis.Params.BaseProposerReward = sdk.NewDecWithPrec(3, 2)   // 3%
	// 			distrGenesis.Params.BonusProposerReward = sdk.NewDecWithPrec(12, 2) // 12%
	// 			newGenState[distrtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&distrGenesis)

	// 			// Authz module
	// 			entries := make([]authz.GrantAuthorization, 0)
	// 			auth, err := codectypes.NewAnyWithValue(
	// 				authz.NewGenericAuthorization(sdk.MsgTypeURL(&oracletypes.MsgReportData{})),
	// 			)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			// Using genesis time + 2500 years as expiration of grant
	// 			expirationTime := genesisTime.AddDate(2500, 0, 0)
	// 			for _, reps := range oracleGenesisV039.Reporters {
	// 				val, err := sdk.ValAddressFromBech32(reps.Validator)
	// 				if err != nil {
	// 					return err
	// 				}
	// 				v := sdk.AccAddress(val).String()
	// 				for _, r := range reps.Reporters {
	// 					if v != r {
	// 						entries = append(entries, authz.GrantAuthorization{
	// 							Granter:       v,
	// 							Grantee:       r,
	// 							Authorization: auth,
	// 							Expiration:    expirationTime,
	// 						})
	// 					}
	// 				}
	// 			}
	// 			authzGenesis := authz.NewGenesisState(entries)
	// 			newGenState[authz.ModuleName] = clientCtx.Codec.MustMarshalJSON(authzGenesis)

	// 			oracleGenesis := oracletypes.DefaultGenesisState()
	// 			oracleGenesis.Params.IBCRequestEnabled = false
	// 			oracleGenesis.OracleScripts = oracleGenesisV039.OracleScripts

	// 			for _, dataSource := range oracleGenesisV039.DataSources {
	// 				oracleGenesis.DataSources = append(oracleGenesis.DataSources, oracletypes.DataSource{
	// 					Owner:       dataSource.Owner,
	// 					Name:        dataSource.Name,
	// 					Description: dataSource.Description,
	// 					Filename:    dataSource.Filename,
	// 					Treasury:    dataSource.Owner,
	// 					Fee:         sdk.NewCoins(),
	// 				})
	// 			}
	// 			newGenState[oracletypes.ModuleName] = v043Codec.MustMarshalJSON(oracleGenesis)

	// 			genDoc.AppState, err = json.Marshal(newGenState)
	// 			if err != nil {
	// 				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
	// 			}

	// 			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
	// 			if chainID != "" {
	// 				genDoc.ChainID = chainID
	// 			}

	// 			initialHeight, _ := cmd.Flags().GetInt(flagInitialHeight)
	// 			genDoc.InitialHeight = int64(initialHeight)

	// 			bz, err := tmjson.Marshal(genDoc)
	// 			if err != nil {
	// 				return errors.Wrap(err, "failed to marshal genesis doc")
	// 			}

	// 			sortedBz, err := sdk.SortJSON(bz)
	// 			if err != nil {
	// 				return errors.Wrap(err, "failed to sort JSON genesis doc")
	// 			}

	// 			fmt.Println(string(sortedBz))
	// 			return nil
	// 		},
	// 	}
	// 	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	// 	cmd.Flags().String(flagChainID, "band-laozi", "override chain_id with this flag")
	// 	cmd.Flags().Int(flagInitialHeight, 0, "Set the starting height for the chain")
	// return cmd
	return nil
}
