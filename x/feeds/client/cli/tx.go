package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/bandprotocol/chain/v3/pkg/grant"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	"github.com/bandprotocol/chain/v3/x/feeds/types"
)

const (
	flagExpiration = "expiration"
	flagFeeLimit   = "fee-limit"
)

// getGrantMsgTypes returns types for GrantMsg.
func getGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&types.MsgSubmitSignalPrices{}),
	}
}

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetTxCmdAddFeeders(),
		GetTxCmdRemoveFeeders(),
		GetTxCmdUpdateReferenceSourceConfig(),
		GetTxCmdVote(),
	)

	return txCmd
}

// GetTxCmdAddFeeders creates a CLI command for adding new feeders
func GetTxCmdAddFeeders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-feeders [feeder1] [feeder2] ...",
		Short: "Add agents authorized to submit signal prices transactions.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Add agents authorized to submit feeds transactions.
Example:
$ %s tx feeds add-feeders band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
`,
				version.AppName,
			),
		),
		RunE: grant.AddGranteeCmd(getGrantMsgTypes(), flagExpiration),
	}

	cmd.Flags().
		Int64(
			flagExpiration,
			time.Now().AddDate(2500, 0, 0).Unix(),
			"The Unix timestamp. Default is 2500 years(forever).",
		)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdRemoveFeeders creates a CLI command for removing feeders from granter
func GetTxCmdRemoveFeeders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-feeders [feeder1] [feeder2] ...",
		Short: "Remove agents from the list of authorized feeders.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Remove agents from the list of authorized feeders.
Example:
$ %s tx feeds remove-feeders band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
`,
				version.AppName,
			),
		),
		RunE: grant.RemoveGranteeCmd(getGrantMsgTypes()),
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdUpdateReferenceSourceConfig creates a CLI command for updating reference source config
func GetTxCmdUpdateReferenceSourceConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-reference-source-config [registry-ipfs-hash] [registry-version]",
		Short: "Update reference source config",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Update reference source configuration that will be use as the default service for price querying.
Example:
$ %s tx feeds update-reference-source-config <YOUR_IPFS_HASH> 1.0.0 --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			admin := clientCtx.GetFromAddress()
			referenceSourceConfig := types.NewReferenceSourceConfig(args[0], args[1])

			msg := types.NewMsgUpdateReferenceSourceConfig(admin.String(), referenceSourceConfig)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdVote creates a CLI command for voting signals
func GetTxCmdVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [signal_id1],[power1] [signal_id2],[power2] ...",
		Short: "Vote signal ids and their powers",
		Args:  cobra.MinimumNArgs(0),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Vote signal ids and their power.
Example:
$ %s tx feeds vote BTC,1000000 --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			voter := clientCtx.GetFromAddress()
			var signals []types.Signal
			for i, arg := range args {
				idAndPower := strings.SplitN(arg, ",", 2)
				if len(idAndPower) != 2 {
					return fmt.Errorf("argument %d is not valid", i)
				}
				power, err := strconv.ParseInt(idAndPower[1], 0, 64)
				if err != nil {
					return err
				}
				signals = append(
					signals, types.NewSignal(
						idAndPower[0],
						power,
					),
				)
			}

			msg := types.NewMsgVote(voter.String(), signals)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRequestSignature implements the request signature handler.
func GetCmdRequestSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feeds-prices [signal_id1,signal_id2,...] [encoder]",
		Short: "Request bandtss signature prices from list of signal id and encoder (1: fixed-point abi, 2: tick abi)",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Request bandtss signature from list of signal id and encoder (1: fixed-point abi, 2: tick abi)
Example:
$ %s tx bandtss request-signature feeds-prices CS:ETH-USD,CS:USDT-USD 1 --fee-limit 10uband
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
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

			signalIDs := strings.Split(args[0], ",")

			encoder, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress().String()
			content := types.NewFeedSignatureOrder(signalIDs, types.Encoder(encoder))

			msg, err := bandtsstypes.NewMsgRequestSignature(content, feeLimit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
