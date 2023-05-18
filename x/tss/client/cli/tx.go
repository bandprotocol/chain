package cli

import (
	"encoding/hex"
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

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

const (
	flagExpiration = "expiration"
)

// NewTxCmd returns a root CLI command handler for all x/tss transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "TSS transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(MsgAddGrantee())
	txCmd.AddCommand(MsgRemoveGrantees())
	txCmd.AddCommand(MsgCreateGroupCmd())
	txCmd.AddCommand(MsgSubmitDKGRound1Cmd())
	txCmd.AddCommand(MsgSubmitDKGRound2Cmd())

	return txCmd
}

func MsgAddGrantee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-grantees [grantee1] [grantee2] ...",
		Short: "Add agents authorized to submit tss transactions.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add agents authorized to submit tss transactions.
Example:
$ %s tx oracle add-grantees band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
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
			msgs := []sdk.Msg{}

			for _, arg := range args {
				grantee, err := sdk.AccAddressFromBech32(arg)
				if err != nil {
					return err
				}

				gMsgs, err := combineGrantMsgs(granter, grantee, types.MsgGrants, &expTime)
				if err != nil {
					return err
				}

				msgs = append(msgs, gMsgs...)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	cmd.Flags().
		Int64(flagExpiration, time.Now().AddDate(2500, 0, 0).Unix(), "The Unix timestamp. Default is 2500 years(forever).")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgRemoveGrantees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-grantees [grantee1] [grantee2] ...",
		Short: "Remove agents from the list of authorized grantees.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Remove agents from the list of authorized grantees.
Example:
$ %s tx oracle remove-grantees band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
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
			msgs := []sdk.Msg{}

			for _, arg := range args {
				grantee, err := sdk.AccAddressFromBech32(arg)
				if err != nil {
					return err
				}

				rMsgs, err := combineRevokeMsgs(granter, grantee, types.MsgGrants)
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

func MsgCreateGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-group [member1,member2,...] [threshold]",
		Args:  cobra.ExactArgs(2),
		Short: "Make a new group for tss module",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a new group for sign tx in tss module.
Example:
$ %s tx tss create-group band15mxunzureevrg646khnunhrl6nxvrj3eree5tz,band1p2t43jx3rz84y4z05xk8dcjjhzzeqnfrt9ua9v,band18f55l8hf4l7zvy8tx28n4r4nksz79p6lp4z305 2 --from mykey
`,
				version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members := strings.Split(args[0], ",")

			threshold, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateGroup{
				Members:   members,
				Threshold: threshold,
				Sender:    clientCtx.GetFromAddress().String(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgSubmitDKGRound1Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dkg-round1 [group_id] [one_time_pub_key] [a0_sing] [one_time_sign] [coefficients-commit1] [coefficients-commit2] ...",
		Args:  cobra.MinimumNArgs(5),
		Short: "submit tss round 1 containing group_id, one_time_pub_key, a0_sing, one_time_sign and coefficients_commit",
		Example: fmt.Sprintf(
			`%s tx tss submit-dkg-round1 [group_id] [one_time_pub_key] [a0_sing] [one_time_sign] [coefficients-commit1] [coefficients-commit2] ...`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			oneTimePubKey, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			a0Sig, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			oneTimeSig, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			var coefficientsCommit tss.Points
			for i := 4; i < len(args); i++ {
				coefficientCommit, err := hex.DecodeString(args[i])
				if err != nil {
					return err
				}

				coefficientsCommit = append(coefficientsCommit, tss.Point(coefficientCommit))
			}

			msg := &types.MsgSubmitDKGRound1{
				GroupID:            tss.GroupID(groupID),
				CoefficientsCommit: coefficientsCommit,
				OneTimePubKey:      oneTimePubKey,
				A0Sig:              a0Sig,
				OneTimeSig:         oneTimeSig,
				Member:             clientCtx.GetFromAddress().String(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgSubmitDKGRound2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dkg-round2 [group_id] [encrypted-secret-share1,encrypted-secret-share2,...]",
		Args:  cobra.ExactArgs(2),
		Short: "submit tss round 2 containing group_id, and n encrypted-secret-shares",
		Example: fmt.Sprintf(
			`%s tx tss submit-dkg-round2 [group_id] [encrypted-secret-share1,encrypted-secret-share2,...]`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			var encryptedSecretShares tss.Scalars
			encryptedSecretSharesStr := strings.Split(args[1], ",")
			for _, essStr := range encryptedSecretSharesStr {
				ess, err := hex.DecodeString(essStr)
				if err != nil {
					return err
				}
				encryptedSecretShares = append(encryptedSecretShares, ess)
			}

			msg := &types.MsgSubmitDKGRound2{
				GroupID: tss.GroupID(groupID),
				Round2Share: &types.Round2Share{
					EncryptedSecretShares: encryptedSecretShares,
				},
				Member: clientCtx.GetFromAddress().String(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
