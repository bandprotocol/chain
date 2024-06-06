package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/pkg/grant"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	flagExpiration = "expiration"
	flagFeeLimit   = "fee-limit"
)

// getGrantMsgTypes returns types for GrantMsg.
func getGrantMsgTypes() []string {
	return []string{
		sdk.MsgTypeURL(&types.MsgSubmitPrices{}),
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
		GetTxCmdAddGrantees(),
		GetTxCmdRemoveGrantees(),
		GetTxCmdSubmitSignals(),
		GetTxCmdUpdatePriceService(),
	)

	return txCmd
}

// GetTxCmdSubmitSignals creates a CLI command for submitting signals
func GetTxCmdSubmitSignals() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signal [signal_id1]:[power1] [signal_id2]:[power2] ...",
		Short: "Signal signal ids and their powers",
		Args:  cobra.MinimumNArgs(0),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Signal signal ids and their power.
Example:
$ %s tx feeds signal BTC:1000000 --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delegator := clientCtx.GetFromAddress()
			var signals []types.Signal
			for i, arg := range args {
				idAndPower := strings.SplitN(arg, ":", 2)
				if len(idAndPower) != 2 {
					return fmt.Errorf("argument %d is not valid", i)
				}
				power, err := strconv.ParseInt(idAndPower[1], 0, 64)
				if err != nil {
					return err
				}
				signals = append(
					signals, types.Signal{
						ID:    idAndPower[0],
						Power: power,
					},
				)
			}

			msg := types.MsgSubmitSignals{
				Delegator: delegator.String(),
				Signals:   signals,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdAddGrantees creates a CLI command for adding new grantees
func GetTxCmdAddGrantees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-grantees [grantee1] [grantee2] ...",
		Short: "Add agents authorized to submit prices transactions.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Add agents authorized to submit feeds transactions.
Example:
$ %s tx feeds add-grantees band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
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

// GetTxCmdRemoveGrantees creates a CLI command for removing grantees from granter
func GetTxCmdRemoveGrantees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-grantees [grantee1] [grantee2] ...",
		Short: "Remove agents from the list of authorized grantees.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Remove agents from the list of authorized grantees.
Example:
$ %s tx feeds remove-grantees band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
`,
				version.AppName,
			),
		),
		RunE: grant.RemoveGranteeCmd(getGrantMsgTypes()),
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdUpdatePriceService creates a CLI command for updating price service
func GetTxCmdUpdatePriceService() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-price-service [hash] [version] [url]",
		Short: "Update reference price service",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(
				`Update reference price service that will be use as the default service for price querying.
Example:
$ %s tx feeds update-price-service 1234abcedf 1.0.0 http://www.example.com --from mykey
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
			priceService := types.PriceService{
				Hash:    args[0],
				Version: args[1],
				Url:     args[2],
			}

			msg := types.MsgUpdatePriceService{
				Admin:        admin.String(),
				PriceService: priceService,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdRequestSignature implements the request signature handler.
func GetCmdRequestSignature() *cobra.Command {
	return &cobra.Command{
		Use:   "feeds-prices [signal_id1,signal_id2,...] [feeds-type]",
		Short: "Request bandtss signature prices from list of signal id and feeds-type (1: default, 2: tick)",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Request bandtss signature from list of signal id and feeds-type (1: default, 2: tick)
Example:
$ %s tx bandtss request-signature feeds-prices crypto_price.ethusd,crypto_price.usdtusd 1 --fee-limit 10uband
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

			feedsType, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewFeedSignatureOrder(signalIDs, types.FeedType(feedsType))

			msg, err := bandtsstypes.NewMsgRequestSignature(content, feeLimit, from)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
