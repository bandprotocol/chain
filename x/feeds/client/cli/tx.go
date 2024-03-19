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

	"github.com/bandprotocol/chain/v2/x/feeds/types"
)

const (
	flagExpiration = "expiration"
)

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
	)

	return txCmd
}

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
				power, err := strconv.ParseUint(idAndPower[1], 0, 64)
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
			msgs := []sdk.Msg{&msg}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdAddGrantees creates a CLI command for add new grantees
func GetTxCmdAddGrantees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-grantees [grantee1] [grantee2] ...",
		Short: "Add agents authorized to submit feeds transactions.",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			exp, err := cmd.Flags().GetInt64(flagExpiration)
			if err != nil {
				return err
			}
			expTime := time.Unix(exp, 0)

			granter := clientCtx.GetFromAddress()
			var msgs []sdk.Msg

			for _, arg := range args {
				grantee, err := sdk.AccAddressFromBech32(arg)
				if err != nil {
					return err
				}

				gMsgs, err := combineGrantMsgs(granter, grantee, types.GetGrantMsgTypes(), &expTime)
				if err != nil {
					return err
				}

				msgs = append(msgs, gMsgs...)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
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

// GetTxCmdRemoveGrantees creates a CLI command for remove grantees from granter
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
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			granter := clientCtx.GetFromAddress()
			var msgs []sdk.Msg

			for _, arg := range args {
				grantee, err := sdk.AccAddressFromBech32(arg)
				if err != nil {
					return err
				}

				rMsgs, err := combineRevokeMsgs(granter, grantee, types.GetGrantMsgTypes())
				if err != nil {
					return err
				}

				msgs = append(msgs, rMsgs...)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
