package cli

import (
	"fmt"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	oracleCmd := &cobra.Command{
		Use:                        oracletypes.ModuleName,
		Short:                      "oracle transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	oracleCmd.AddCommand(
		GetCmdCreateDataSource(),
		GetCmdEditDataSource(),
		GetCmdCreateOracleScript(),
		GetCmdEditOracleScript(),
		GetCmdRequest(),
		GetCmdActivate(),
		GetCmdAddReporters(),
		GetCmdRemoveReporter(),
	)

	return oracleCmd
}

// GetCmdRequest implements the request command handler.
func GetCmdRequest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request [oracle-script-id] [ask-count] [min-count] (-l [fee-limit]) (-p [prepare-gas]) (-e [execute-gas]) (-c [calldata]) (-m [client-id])",
		Short: "Make a new data request via an existing oracle script",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a new request via an existing oracle script with the configuration flags.
Example:
$ %s tx oracle request 1 4 3 -c 1234abcdef -m client-id -l 100loki -p 4000 -e 3000000 --from mykey
$ %s tx oracle request 1 4 3 --calldata 1234abcdef --client-id cliend-id --fee-limit 100loki --prepare-gas 4000 --execute-gas 300000 --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			rawOsId, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			oracleScriptID := oracletypes.OracleScriptID(rawOsId)

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

			rawFeeLimit, err := cmd.Flags().GetString(flagFeeLimit)
			if err != nil {
				return err
			}

			feeLimit, err := sdk.ParseCoinsNormalized(rawFeeLimit)
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

			msg := oracletypes.NewMsgRequestData(
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
	cmd.Flags().StringP(flagFeeLimit, "l", oracletypes.DefaultFeeLimit.String(), "Gas used for execution phase")
	cmd.Flags().Uint64P(flagPrepareGas, "p", oracletypes.DefaultPrepareGas, "Gas used for preparation phase")
	cmd.Flags().Uint64P(flagExecuteGas, "e", oracletypes.DefaultExecuteGas, "Gas used for execution phase")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdCreateDataSource implements the create data source command handler.
func GetCmdCreateDataSource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-data-source (--name [name]) (--description [description]) (--script [path-to-script]) (--fee [fee]) (--owner [owner])",
		Short: "Create a new data source",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a new data source that will be used by oracle scripts.
Example:
$ %s tx oracle create-data-source --name coingecko-price --description "The script that queries crypto price from cryptocompare" --script ../price.sh --owner odin15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --fee 10loki,100geo --from mykey
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

			rawFee, err := cmd.Flags().GetString(flagFee)
			if err != nil {
				return err
			}

			fee, err := sdk.ParseCoinsNormalized(rawFee)
			if err != nil {
				return err
			}

			rawOwner, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}

			owner, err := sdk.AccAddressFromBech32(rawOwner)
			if err != nil {
				return err
			}

			msg := oracletypes.NewMsgCreateDataSource(
				name,
				description,
				execBytes,
				fee,
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
	cmd.Flags().String(flagFee, "", "Fee for usage of this data source")
	cmd.Flags().String(flagOwner, "", "Owner of this data source")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TODO add args to examples
// GetCmdEditDataSource implements the edit data source command handler.
func GetCmdEditDataSource() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-data-source [id] (--name [name]) (--description [description]) (--script [path-to-script]) (--fee [fee]) (--owner [owner])",
		Short: "Edit data source",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Edit an existing data source. The caller must be the current data source's owner.
Example:
$ %s tx oracle edit-data-source 1 --name coingecko-price --description The script that queries crypto price from cryptocompare --script ../price.sh --owner band15d4apf20449ajvwycq8ruaypt7v6d345n9fpt9 --fee 10loki,100geo --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			rawID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			dataSourceID := oracletypes.DataSourceID(rawID)

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

			execBytes := oracletypes.DoNotModifyBytes
			if scriptPath != oracletypes.DoNotModify {
				execBytes, err = ioutil.ReadFile(scriptPath)
				if err != nil {
					return err
				}
			}

			rawFee, err := cmd.Flags().GetString(flagFee)
			if err != nil {
				return err
			}

			fee, err := sdk.ParseCoinsNormalized(rawFee)
			if err != nil {
				return err
			}

			rawOwner, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return err
			}

			owner, err := sdk.AccAddressFromBech32(rawOwner)
			if err != nil {
				return err
			}

			msg := oracletypes.NewMsgEditDataSource(
				dataSourceID,
				name,
				description,
				execBytes,
				fee,
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

	cmd.Flags().String(flagName, oracletypes.DoNotModify, "Name of this data source")
	cmd.Flags().String(flagDescription, oracletypes.DoNotModify, "Description of this data source")
	cmd.Flags().String(flagScript, oracletypes.DoNotModify, "Path to this data source script")
	cmd.Flags().String(flagFee, "", "Fee for usage of this data source")
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

			msg := oracletypes.NewMsgCreateOracleScript(
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
			oracleScriptID := oracletypes.OracleScriptID(id)
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
			scriptCode := oracletypes.DoNotModifyBytes
			if scriptPath != oracletypes.DoNotModify {
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

			msg := oracletypes.NewMsgEditOracleScript(
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
	cmd.Flags().String(flagName, oracletypes.DoNotModify, "Name of this oracle script")
	cmd.Flags().String(flagDescription, oracletypes.DoNotModify, "Description of this oracle script")
	cmd.Flags().String(flagScript, oracletypes.DoNotModify, "Path to this oracle script")
	cmd.Flags().String(flagOwner, "", "Owner of this oracle script")
	cmd.Flags().String(flagSchema, oracletypes.DoNotModify, "Schema of this oracle script")
	cmd.Flags().String(flagSourceCodeURL, oracletypes.DoNotModify, "URL for the source code of this oracle script")

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
			msg := oracletypes.NewMsgActivate(validator)
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
				msgs[i] = oracletypes.NewMsgAddReporter(
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
			msg := oracletypes.NewMsgRemoveReporter(
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
