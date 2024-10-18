package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/bandprotocol/chain/v3/pkg/grant"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

const (
	flagExpiration = "expiration"
)

// GetTxCmd returns a root CLI command handler for all x/tss transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "TSS transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetTxCmdAddGrantees(),
		GetTxCmdRemoveGrantees(),
		GetTxCmdSubmitDKGRound1(),
		GetTxCmdSubmitDKGRound2(),
		GetTxCmdComplain(),
		GetTxCmdConfirm(),
		GetTxCmdSubmitDEs(),
		GetTxCmdSubmitSignature(),
	)

	return txCmd
}

// GetTxCmdAddGrantees creates a CLI command for add new grantees
func GetTxCmdAddGrantees() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-grantees [grantee1] [grantee2] ...",
		Short: "Add agents authorized to submit tss transactions.",
		Args:  cobra.MinimumNArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Add agents authorized to submit tss transactions.
Example:
$ %s tx tss add-grantees band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
`,
				version.AppName,
			),
		),
		RunE: grant.AddGranteeCmd(types.GetGrantMsgTypes(), flagExpiration),
	}

	cmd.Flags().
		Int64(flagExpiration, time.Now().AddDate(2500, 0, 0).Unix(), "The Unix timestamp. Default is 2500 years(forever).")
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
			fmt.Sprintf(`Remove agents from the list of authorized grantees.
Example:
$ %s tx tss remove-grantees band1p40yh3zkmhcv0ecqp3mcazy83sa57rgjp07dun band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs --from mykey
`,
				version.AppName,
			),
		),
		RunE: grant.RemoveGranteeCmd(types.GetGrantMsgTypes()),
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdSubmitDKGRound1 creates a CLI command for submitting DKG round 1 information.
func GetTxCmdSubmitDKGRound1() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dkg-round1 [group_id] [member_id] [one_time_pub_key] [a0_sign] [one_time_sign] [coefficients-commit1] [coefficients-commit2] ...",
		Args:  cobra.MinimumNArgs(6),
		Short: "submit tss round 1 containing group_id, member_id, one_time_pub_key, a0_sign, one_time_sign and coefficients_commit",
		Example: fmt.Sprintf(
			`%s tx tss submit-dkg-round1 [group_id] [member_id] [one_time_pub_key] [a0_sign] [one_time_sign] [coefficients-commit1] [coefficients-commit2] ...`,
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

			a0Signature, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			oneTimeSignature, err := hex.DecodeString(args[4])
			if err != nil {
				return err
			}

			var coefficientCommits tss.Points
			for i := 5; i < len(args); i++ {
				coefficientCommit, err := hex.DecodeString(args[i])
				if err != nil {
					return err
				}

				point, err := tss.NewPoint(coefficientCommit)
				if err != nil {
					return err
				}

				coefficientCommits = append(coefficientCommits, point)
			}

			r1Info := types.NewRound1Info(
				tss.MemberID(memberID),
				coefficientCommits,
				oneTimePubKey,
				a0Signature,
				oneTimeSignature,
			)
			sender := clientCtx.GetFromAddress().String()
			msg := types.NewMsgSubmitDKGRound1(tss.GroupID(groupID), r1Info, sender)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdSubmitDKGRound2 creates a CLI command for submitting DKG round 2 information.
func GetTxCmdSubmitDKGRound2() *cobra.Command {
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

			var encryptedSecretShares tss.EncSecretShares
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

			r2Info := types.NewRound2Info(tss.MemberID(memberID), encryptedSecretShares)
			sender := clientCtx.GetFromAddress().String()
			msg := types.NewMsgSubmitDKGRound2(tss.GroupID(groupID), r2Info, sender)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdComplain creates a CLI command for submitting complaint message.
func GetTxCmdComplain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complain [group_id] [complaints-json-file]",
		Args:  cobra.ExactArgs(2),
		Short: "complain containing group_id and complaints data",
		Example: fmt.Sprintf(`
%s tx tss complain [group_id] [complaints-json-file]

Where complaints.json contains:
{
	complaints: [
		{
            "complainant": 1,
            "respondent": 2,
            "key_sym": "symmetric key between complainant and respondent",
            "signature": "signature that complain by complainant"
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

			complaints, err := parseComplaints(args[1])
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress().String()
			msg := types.NewMsgComplain(tss.GroupID(groupID), complaints, sender)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdConfirm creates a CLI command for submitting confirm message.
func GetTxCmdConfirm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm [group_id] [member_id] [own_pub_key_sig]",
		Args:  cobra.ExactArgs(3),
		Short: "submit confirm containing group_id, member_id, and own_pub_key_sig",
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

			msg := types.NewMsgConfirm(
				tss.GroupID(groupID),
				tss.MemberID(memberID),
				ownPubKeySig,
				clientCtx.GetFromAddress().String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdSubmitDEs creates a CLI command for submitting DE message.
func GetTxCmdSubmitDEs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-multi-de [d,e] [d,e] ...",
		Args:  cobra.MinimumNArgs(1),
		Short: "submit multiple DE containing address and DEs",
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

				des = append(des, types.NewDE(d, e))
			}

			msg := types.NewMsgSubmitDEs(des, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdSubmitSignature creates a CLI command for submitting a signature.
func GetTxCmdSubmitSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-signature [signing_id] [member_id] [signature]",
		Args:  cobra.ExactArgs(3),
		Short: "submit-signature the message by sending signing ID, member ID and signature",
		Example: fmt.Sprintf(
			`%s tx tss submit-signature [signing_id] [member_id] [signature]`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			signingID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			memberID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			sig, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmitSignature(
				tss.SigningID(signingID),
				tss.MemberID(memberID),
				sig,
				clientCtx.GetFromAddress().String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
