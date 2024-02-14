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
	flagGroupID    = "group-id"
	flagFeeLimit   = "fee-limit"
)

// GetTxCmd returns a root CLI command handler for all x/tss transaction commands.
func GetTxCmd(requestSignatureCmds []*cobra.Command) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "TSS transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// Create the command for requesting a signature.
	cmdRequestSignature := GetTxCmdRequestSignature()

	// Create the command for requesting a signature using text input.
	cmdTextRequestSignature := GetTxCmdTextRequestSignature()

	// Add the text signature command as a subcommand.
	flags.AddTxFlagsToCmd(cmdTextRequestSignature)
	cmdRequestSignature.AddCommand(cmdTextRequestSignature)

	// Loop through and add the provided request signature commands as subcommands.
	for _, cmd := range requestSignatureCmds {
		flags.AddTxFlagsToCmd(cmd)
		cmdRequestSignature.AddCommand(cmd)
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
		GetTxCmdActivate(),
		GetTxCmdHealthCheck(),

		cmdRequestSignature,
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

				gMsgs, err := combineGrantMsgs(granter, grantee, types.GetTSSGrantMsgTypes(), &expTime)
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

// GetTxCmdRemoveGrantees creates a CLI command for remove grantees from granter
func GetTxCmdRemoveGrantees() *cobra.Command {
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

				rMsgs, err := combineRevokeMsgs(granter, grantee, types.GetTSSGrantMsgTypes())
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

// GetTxCmdSubmitDKGRound1 creates a CLI command for CLI command for Msg/SubmitDKGRound1.
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

			msg := &types.MsgSubmitDKGRound1{
				GroupID: tss.GroupID(groupID),
				Round1Info: types.Round1Info{
					MemberID:           tss.MemberID(memberID),
					CoefficientCommits: coefficientCommits,
					OneTimePubKey:      oneTimePubKey,
					A0Signature:        a0Signature,
					OneTimeSignature:   oneTimeSignature,
				},
				Address: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdSubmitDKGRound2 creates a CLI command for CLI command for Msg/SubmitDKGRound2.
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

			msg := &types.MsgSubmitDKGRound2{
				GroupID: tss.GroupID(groupID),
				Round2Info: types.Round2Info{
					MemberID:              tss.MemberID(memberID),
					EncryptedSecretShares: encryptedSecretShares,
				},
				Address: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdComplain creates a CLI command for CLI command for Msg/Complaint.
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

			msg := &types.MsgComplain{
				GroupID:    tss.GroupID(groupID),
				Complaints: complaints,
				Address:    clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdConfirm creates a CLI command for CLI command for Msg/Confirm.
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

			msg := &types.MsgConfirm{
				GroupID:      tss.GroupID(groupID),
				MemberID:     tss.MemberID(memberID),
				OwnPubKeySig: ownPubKeySig,
				Address:      clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdSubmitDEs creates a CLI command for CLI command for Msg/SubmitDEPairs.
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

				des = append(des, types.DE{PubD: d, PubE: e})
			}

			msg := &types.MsgSubmitDEs{
				DEs:     des,
				Address: clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdRequestSignature creates a CLI command for CLI command for Msg/RequestSignature.
func GetTxCmdRequestSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-signature",
		Short: "request signature from the group",
	}

	cmd.PersistentFlags().String(flagFeeLimit, "", "The maximum tokens that will be paid for this request")
	cmd.PersistentFlags().Uint64(flagGroupID, 0, "The group that is requested to sign the result")

	_ = cmd.MarkPersistentFlagRequired(flagFeeLimit)
	_ = cmd.MarkPersistentFlagRequired(flagGroupID)

	return cmd
}

// GetTxCmdTextRequestSignature creates a CLI command for CLI command for Msg/TextRequestSignature.
func GetTxCmdTextRequestSignature() *cobra.Command {
	return &cobra.Command{
		Use:   "text [message]",
		Args:  cobra.ExactArgs(1),
		Short: "request signature of the message from the group",
		Example: fmt.Sprintf(
			`%s tx tss request-signature text [message] --group-id 1 --fee-limit 10uband`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			gid, err := cmd.Flags().GetUint64(flagGroupID)
			if err != nil {
				return err
			}

			data, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			content := types.NewTextSignatureOrder(data)

			coinStr, err := cmd.Flags().GetString(flagFeeLimit)
			if err != nil {
				return err
			}

			feeLimit, err := sdk.ParseCoinsNormalized(coinStr)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgRequestSignature(
				tss.GroupID(gid),
				content,
				feeLimit,
				clientCtx.GetFromAddress(),
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// GetTxCmdSubmitSignature creates a CLI command for CLI command for Msg/SubmitSignature.
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

			msg := &types.MsgSubmitSignature{
				SigningID: tss.SigningID(signingID),
				MemberID:  tss.MemberID(memberID),
				Signature: sig,
				Address:   clientCtx.GetFromAddress().String(),
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdActivate creates a CLI command for CLI command for Msg/Activate.
func GetTxCmdActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Args:  cobra.NoArgs,
		Short: "activate the status of the address",
		Example: fmt.Sprintf(
			`%s tx tss activate`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgActivate{
				Address: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdHealthCheck creates a CLI command for CLI command for Msg/HealthCheck.
func GetTxCmdHealthCheck() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health-check",
		Args:  cobra.NoArgs,
		Short: "update the active status of the address to ensure that the TSS process is still running",
		Example: fmt.Sprintf(
			`%s tx tss health-check`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgHealthCheck{
				Address: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
