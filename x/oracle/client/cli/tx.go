package cli

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

const (
	flagName          = "name"
	flagDescription   = "description"
	flagScript        = "script"
	flagOwner         = "owner"
	flagCalldata      = "calldata"
	flagClientID      = "client-id"
	flagSchema        = "schema"
	flagSourceCodeURL = "url"
	flagPrepareGas    = "prepare-gas"
	flagExecuteGas    = "execute-gas"
	flagFeeLimit      = "fee-limit"
	flagFee           = "fee"
	flagTreasury      = "treasury"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "oracle transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetCmdRequest(),
		GetCmdCreateDataSource(),
		GetCmdEditDataSource(),
		GetCmdCreateOracleScript(),
		GetCmdEditOracleScript(),
		GetCmdActivate(),
		GetCmdAddReporters(),
		GetCmdRemoveReporter(),
		GetCmdDepositRequestPool(),
	)

	return txCmd
}

// GetCmdRequest implements the request command handler.
func GetCmdRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request [oracle-script-id] [ask-count] [min-count] (-c [calldata]) (-m [client-id]) (--prepare-gas=[prepare-gas] (--execute-gas=[execute-gas])) (--fee-limit=[fee-limit])",
		Short: "Make a new data request via an existing oracle script",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a new request via an existing oracle script with the configuration flags.
Example:
$ %s tx oracle request 1 4 3 -c 1234abcdef -x 20 -m client-id --from mykey
$ %s tx oracle request 1 4 3 --calldata 1234abcdef --client-id cliend-id --fee-limit 10uband --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			int64OracleScriptID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			oracleScriptID := types.OracleScriptID(int64OracleScriptID)

			askCount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			minCount, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			calldata, err := cmd.Flags().GetBytesHex(flagCalldata)
			if err != nil {
				return err
			}

			clientID, err := cmd.Flags().GetString(flagClientID)
			if err != nil {
				return err
			}

			prepareGas, err := cmd.Flags().GetUint64(flagPrepareGas)
			if err != nil {
				return err
			}

			executeGas, err := cmd.Flags().GetUint64(flagExecuteGas)
			if err != nil {
				return err
			}

			coinStr, err := cmd.Flags().GetString(flagFeeLimit)
			if err != nil {
				return err
			}

			feeLimit, err := sdk.ParseCoinsNormalized(coinStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgRequestData(
				oracleScriptID,
				calldata,
				askCount,
				minCount,
				clientID,
				feeLimit,
				prepareGas,
				executeGas,
				clientCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().BytesHexP(flagCalldata, "c", nil, "Calldata used in calling the oracle script")
	cmd.Flags().StringP(flagClientID, "m", "", "Requester can match up the request with response by clientID")
	cmd.Flags().Uint64(flagPrepareGas, 50000, "Prepare gas used in fee counting for prepare request")
	cmd.Flags().Uint64(flagExecuteGas, 300000, "Execute gas used in fee counting for execute request")
	cmd.Flags().String(flagFeeLimit, "", "the maximum tokens that will be paid to all data source providers")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdCreateDataSource implements the create data source command handler.
func GetCmdCreateDataSource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-data-source (--name [name]) (--description [description]) (--script [path-to-script]) (--owner [owner]) (--treasury [treasury]) (--fee [fee])",
		Short: "Create a new data source",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new data source that will be used by oracle scripts.
Example:
$ %s tx oracle create-data-source --name coingecko-price --description "The script that queries crypto price from cryptocompare" --script ../price.sh --owner band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --treasury band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --fee 10uband --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(flagDescription)
			if err != nil {
				return err
			}

			scriptPath, err := cmd.Flags().GetString(flagScript)
			if err != nil {
				return err
			}
			execBytes, err := ioutil.ReadFile(scriptPath)
			if err != nil {
				return err
			}

			ownerStr, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(ownerStr)
			if err != nil {
				return err
			}

			coinStr, err := cmd.Flags().GetString(flagFee)
			if err != nil {
				return err
			}

			fee, err := sdk.ParseCoinsNormalized(coinStr)
			if err != nil {
				return err
			}

			treasuryStr, err := cmd.Flags().GetString(flagTreasury)
			if err != nil {
				return err
			}
			treasury, err := sdk.AccAddressFromBech32(treasuryStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateDataSource(
				name,
				description,
				execBytes,
				fee,
				treasury,
				owner,
				clientCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(flagName, "", "Name of this data source")
	cmd.Flags().String(flagDescription, "", "Description of this data source")
	cmd.Flags().String(flagScript, "", "Path to this data source script")
	cmd.Flags().String(flagOwner, "", "Owner of this data source")
	cmd.Flags().String(flagTreasury, "", "Treasury of this data source")
	cmd.Flags().String(flagFee, "", "Fee of this data source")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdEditDataSource implements the edit data source command handler.
func GetCmdEditDataSource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-data-source [id] (--name [name]) (--description [description]) (--script [path-to-script]) (--owner [owner]) (--treasury [treasury]) (--fee [fee])",
		Short: "Edit data source",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Edit an existing data source. The caller must be the current data source's owner.
Example:
$ %s tx oracle edit-data-source 1 --name coingecko-price --description The script that queries crypto price from cryptocompare --script ../price.sh --owner band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --treasury band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --fee 10uband --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			int64ID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			dataSourceID := types.DataSourceID(int64ID)
			name, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(flagDescription)
			if err != nil {
				return err
			}

			scriptPath, err := cmd.Flags().GetString(flagScript)
			if err != nil {
				return err
			}
			execBytes := types.DoNotModifyBytes
			if scriptPath != types.DoNotModify {
				execBytes, err = ioutil.ReadFile(scriptPath)
				if err != nil {
					return err
				}
			}

			ownerStr, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(ownerStr)
			if err != nil {
				return err
			}

			// TODO: Support do-not-modify fee
			coinStr, err := cmd.Flags().GetString(flagFee)
			if err != nil {
				return err
			}
			fee, err := sdk.ParseCoinsNormalized(coinStr)
			if err != nil {
				return err
			}

			treasuryStr, err := cmd.Flags().GetString(flagTreasury)
			if err != nil {
				return err
			}
			treasury, err := sdk.AccAddressFromBech32(treasuryStr)
			if err != nil {
				return err
			}

			msg := types.NewMsgEditDataSource(
				dataSourceID,
				name,
				description,
				execBytes,
				fee,
				treasury,
				owner,
				clientCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(flagName, types.DoNotModify, "Name of this data source")
	cmd.Flags().String(flagDescription, types.DoNotModify, "Description of this data source")
	cmd.Flags().String(flagScript, types.DoNotModify, "Path to this data source script")
	cmd.Flags().String(flagTreasury, "", "Treasury of this data source")
	cmd.Flags().String(flagFee, "", "Fee of this data source")
	cmd.Flags().String(flagOwner, "", "Owner of this data source")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdCreateOracleScript implements the create oracle script command handler.
func GetCmdCreateOracleScript() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-oracle-script (--name [name]) (--description [description]) (--script [path-to-script]) (--owner [owner]) (--schema [schema]) (--url [source-code-url])",
		Short: "Create a new oracle script that will be used by data requests.",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new oracle script that will be used by data requests.
Example:
$ %s tx oracle create-oracle-script --name eth-price --description "Oracle script for getting Ethereum price" --script ../eth_price.wasm --owner band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			name, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return err
			}
			description, err := cmd.Flags().GetString(flagDescription)
			if err != nil {
				return err
			}

			scriptPath, err := cmd.Flags().GetString(flagScript)
			if err != nil {
				return err
			}
			scriptCode, err := ioutil.ReadFile(scriptPath)
			if err != nil {
				return err
			}

			ownerStr, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(ownerStr)
			if err != nil {
				return err
			}

			schema, err := cmd.Flags().GetString(flagSchema)
			if err != nil {
				return err
			}

			sourceCodeURL, err := cmd.Flags().GetString(flagSourceCodeURL)
			if err != nil {
				return err
			}

			msg := types.NewMsgCreateOracleScript(
				name,
				description,
				schema,
				sourceCodeURL,
				scriptCode,
				owner,
				clientCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(flagName, "", "Name of this oracle script")
	cmd.Flags().String(flagDescription, "", "Description of this oracle script")
	cmd.Flags().String(flagScript, "", "Path to this oracle script")
	cmd.Flags().String(flagOwner, "", "Owner of this oracle script")
	cmd.Flags().String(flagSchema, "", "Schema of this oracle script")
	cmd.Flags().String(flagSourceCodeURL, "", "URL for the source code of this oracle script")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdEditOracleScript implements the editing of oracle script command handler.
func GetCmdEditOracleScript() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-oracle-script [id] (--name [name]) (--description [description]) (--script [path-to-script]) (--owner [owner]) (--schema [schema]) (--url [source-code-url])",
		Short: "Edit an existing oracle script that will be used by data requests.",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Edit an existing oracle script that will be used by data requests.
Example:
$ %s tx oracle edit-oracle-script 1 --name eth-price --description "Oracle script for getting Ethereum price" --script ../eth_price.wasm --owner band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			oracleScriptID := types.OracleScriptID(id)
			name, err := cmd.Flags().GetString(flagName)
			if err != nil {
				return err
			}

			description, err := cmd.Flags().GetString(flagDescription)
			if err != nil {
				return err
			}

			scriptPath, err := cmd.Flags().GetString(flagScript)
			if err != nil {
				return err
			}
			scriptCode := types.DoNotModifyBytes
			if scriptPath != types.DoNotModify {
				scriptCode, err = ioutil.ReadFile(scriptPath)
				if err != nil {
					return err
				}
			}

			ownerStr, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}
			owner, err := sdk.AccAddressFromBech32(ownerStr)
			if err != nil {
				return err
			}

			schema, err := cmd.Flags().GetString(flagSchema)
			if err != nil {
				return err
			}

			sourceCodeURL, err := cmd.Flags().GetString(flagSourceCodeURL)
			if err != nil {
				return err
			}

			msg := types.NewMsgEditOracleScript(
				oracleScriptID,
				name,
				description,
				schema,
				sourceCodeURL,
				scriptCode,
				owner,
				clientCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().String(flagName, types.DoNotModify, "Name of this oracle script")
	cmd.Flags().String(flagDescription, types.DoNotModify, "Description of this oracle script")
	cmd.Flags().String(flagScript, types.DoNotModify, "Path to this oracle script")
	cmd.Flags().String(flagOwner, "", "Owner of this oracle script")
	cmd.Flags().String(flagSchema, types.DoNotModify, "Schema of this oracle script")
	cmd.Flags().String(flagSourceCodeURL, types.DoNotModify, "URL for the source code of this oracle script")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdActivate implements the activate command handler.
func GetCmdActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate myself to become an oracle validator.",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Activate myself to become an oracle validator.
Example:
$ %s tx oracle activate --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			validator := sdk.ValAddress(clientCtx.GetFromAddress())
			msg := types.NewMsgActivate(validator)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAddReporters implements the add reporters command handler.
func GetCmdAddReporters() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-reporters [reporter1] [reporter2] ...",
		Short: "Add agents authorized to submit report transactions.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add agents authorized to submit report transactions.
Example:
$ %s tx oracle add-reporters band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			validator := sdk.ValAddress(clientCtx.GetFromAddress())
			msgs := make([]sdk.Msg, len(args))
			for i, raw := range args {
				reporter, err := sdk.AccAddressFromBech32(raw)
				if err != nil {
					return err
				}
				msgs[i] = types.NewMsgAddReporter(
					validator,
					reporter,
				)
				err = msgs[i].ValidateBasic()
				if err != nil {
					return err
				}
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdRemoveReporter implements the remove reporter command handler.
func GetCmdRemoveReporter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-reporter [reporter]",
		Short: "Remove an agent from the list of authorized reporters.",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove an agent from the list of authorized reporters.
Example:
$ %s tx oracle remove-reporter band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			validator := sdk.ValAddress(clientCtx.GetFromAddress())
			reporter, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			msg := types.NewMsgRemoveReporter(
				validator,
				reporter,
			)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdDepositRequestPool implements the deposit request pool command handler.
func GetCmdDepositRequestPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-request-pool [request-key] [port-id] [channel-id] [amount]",
		Short: "Deposit the coins to request pool",
		Args:  cobra.ExactArgs(4),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit the coins to request pool.
Example:
$ %s tx oracle deposit-request-pool request-key port-1 channel-1 1000uband --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			sender := sdk.AccAddress(clientCtx.GetFromAddress())
			amt, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			portID := args[1]
			if err := host.PortIdentifierValidator(portID); err != nil {
				return err
			}
			channelID := args[2]
			if err := host.ChannelIdentifierValidator(channelID); err != nil {
				return err
			}
			pool := types.GetEscrowAddress(args[0], portID, channelID)
			msg := banktypes.NewMsgSend(sender, pool, amt)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
