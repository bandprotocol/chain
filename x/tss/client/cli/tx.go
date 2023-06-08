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

	txCmd.AddCommand(
		NewAddGranteesCmd(),
		NewRemoveGranteesCmd(),
		NewCreateGroupCmd(),
		NewSubmitDKGRound1Cmd(),
		NewSubmitDKGRound2Cmd(),
		NewComplainCmd(),
		NewConfirmCmd(),
		NewSubmitDEsCmd(),
		NewRequestSignCmd(),
	)

	return txCmd
}

// NewAddGranteesCmd creates a CLI command for add new grantees
func NewAddGranteesCmd() *cobra.Command {
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

				gMsgs, err := combineGrantMsgs(granter, grantee, types.GetMsgGrants(), &expTime)
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

// NewRemoveGranteesCmd creates a CLI command for remove grantees from granter
func NewRemoveGranteesCmd() *cobra.Command {
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

				rMsgs, err := combineRevokeMsgs(granter, grantee, types.GetMsgGrants())
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

// NewCreateGroupCmd creates a CLI command for CLI command for Msg/CreateGroup.
func NewCreateGroupCmd() *cobra.Command {
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

			threshold, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateGroup{
				Members:   members,
				Threshold: threshold,
				Sender:    clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSubmitDKGRound1Cmd creates a CLI command for CLI command for Msg/SubmitDKGRound1.
func NewSubmitDKGRound1Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dkg-round1 [group_id] [member_id] [one_time_pub_key] [a0_sing] [one_time_sign] [coefficients-commit1] [coefficients-commit2] ...",
		Args:  cobra.MinimumNArgs(6),
		Short: "submit tss round 1 containing group_id, member_id, one_time_pub_key, a0_sing, one_time_sign and coefficients_commit",
		Example: fmt.Sprintf(
			`%s tx tss submit-dkg-round1 [group_id] [member_id] [one_time_pub_key] [a0_sing] [one_time_sign] [coefficients-commit1] [coefficients-commit2] ...`,
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

			memberID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			oneTimePubKey, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			a0Sig, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			oneTimeSig, err := hex.DecodeString(args[4])
			if err != nil {
				return err
			}

			var coefficientsCommit tss.Points
			for i := 5; i < len(args); i++ {
				coefficientCommit, err := hex.DecodeString(args[i])
				if err != nil {
					return err
				}

				coefficientsCommit = append(coefficientsCommit, tss.Point(coefficientCommit))
			}

			msg := &types.MsgSubmitDKGRound1{
				GroupID: tss.GroupID(groupID),
				Round1Data: types.Round1Data{
					MemberID:           tss.MemberID(memberID),
					CoefficientsCommit: coefficientsCommit,
					OneTimePubKey:      oneTimePubKey,
					A0Sig:              a0Sig,
					OneTimeSig:         oneTimeSig,
				},
				Member: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSubmitDKGRound2Cmd creates a CLI command for CLI command for Msg/SubmitDKGRound2.
func NewSubmitDKGRound2Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dkg-round2 [group_id] [member_id] [encrypted-secret-share1,encrypted-secret-share2,...]",
		Args:  cobra.MinimumNArgs(2),
		Short: "submit tss round 2 containing group_id, member_id, and n-1 encrypted-secret-shares",
		Example: fmt.Sprintf(
			`%s tx tss submit-dkg-round2 [group_id] [member_id] [encrypted-secret-share1,encrypted-secret-share2,...]`,
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

			memberID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			var encryptedSecretShares tss.Scalars
			if len(args) > 2 {
				encryptedSecretSharesStr := strings.Split(args[2], ",")
				for _, essStr := range encryptedSecretSharesStr {
					ess, err := hex.DecodeString(essStr)
					if err != nil {
						return err
					}
					encryptedSecretShares = append(encryptedSecretShares, ess)
				}
			}

			msg := &types.MsgSubmitDKGRound2{
				GroupID: tss.GroupID(groupID),
				Round2Data: types.Round2Data{
					MemberID:              tss.MemberID(memberID),
					EncryptedSecretShares: encryptedSecretShares,
				},
				Member: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewComplainCmd creates a CLI command for CLI command for Msg/Complain.
func NewComplainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complain [group_id] [complains-json-file]",
		Args:  cobra.ExactArgs(2),
		Short: "complain containing group_id and complains data",
		Example: fmt.Sprintf(`
%s tx tss complain [group_id] [complains-json-file]

Where complains.json contains:
{
	complains: [
		{
			"i": 1,
			"j": 2,
			"key_sym": "symmetric key between i and j",
			"signature": "signature that complain by i",
			"nonce_sym": "symmetric nonce"
		},
		...
	]
}`,

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

			complains, err := parseComplains(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgComplain{
				GroupID:   tss.GroupID(groupID),
				Complains: complains,
				Member:    clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewConfirmCmd creates a CLI command for CLI command for Msg/Confirm.
func NewConfirmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm [group_id] [member_id] [own_pub_key_sig]",
		Args:  cobra.ExactArgs(3),
		Short: "submit tss confirm containing group_id, member_id, and own_pub_key_sig",
		Example: fmt.Sprintf(
			`%s tx tss confirm [group_id] [member_id] [own_pub_key_sig]`,
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

			memberID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			ownPubKeySig, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgConfirm{
				GroupID:      tss.GroupID(groupID),
				MemberID:     tss.MemberID(memberID),
				OwnPubKeySig: ownPubKeySig,
				Member:       clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewSubmitDEsCmd creates a CLI command for CLI command for Msg/SubmitDEPairs.
func NewSubmitDEsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-multi-de [d,e] [d,e] ...",
		Args:  cobra.MinimumNArgs(1),
		Short: "submit tss submit-multi-de containing address and DEs",
		Example: fmt.Sprintf(
			`%s tx tss submit-multi-de [d,e] [d,e] ...`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var des []types.DE
			for i := 0; i < len(args); i++ {
				de := strings.Split(args[i], ",")
				if len(de) != 2 {
					return fmt.Errorf("DE must be 2 value not %v", de)
				}

				d, err := hex.DecodeString(de[0])
				if err != nil {
					return err
				}

				e, err := hex.DecodeString(de[1])
				if err != nil {
					return err
				}

				des = append(des, types.DE{PubD: d, PubE: e})
			}

			msg := &types.MsgSubmitDEs{
				DEs:    des,
				Member: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewRequestSignCmd creates a CLI command for CLI command for Msg/RequestSign.
func NewRequestSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-sign [group_id] [message]",
		Args:  cobra.ExactArgs(2),
		Short: "request tss sign of the message from a group",
		Example: fmt.Sprintf(
			`%s tx tss request-sign [group_id] [message]`,
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

			data, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			msg := &types.MsgRequestSign{
				GroupID: tss.GroupID(groupID),
				Message: data,
				Sender:  clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
