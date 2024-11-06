package cli

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

const (
	flagExpiration    = "expiration"
	flagFeeLimit      = "fee-limit"
	flagIncomingGroup = "incoming-group"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(requestSignatureCmds []*cobra.Command) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bandtss transactions subcommands",
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
		GetTxCmdActivate(),
		cmdRequestSignature,
	)

	return txCmd
}

// GetTxCmdRequestSignature creates a CLI command for create a signature request.
func GetTxCmdRequestSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-signature",
		Short: "request signature from the group",
	}

	cmd.PersistentFlags().String(flagFeeLimit, "", "The maximum tokens that will be paid for this request")

	_ = cmd.MarkPersistentFlagRequired(flagFeeLimit)

	return cmd
}

// GetTxCmdTextRequestSignature creates a CLI command for create a signature on text message.
func GetTxCmdTextRequestSignature() *cobra.Command {
	return &cobra.Command{
		Use:   "text [message]",
		Args:  cobra.ExactArgs(1),
		Short: "request signature of the message from the current group",
		Example: fmt.Sprintf(
			`%s tx bandtss request-signature text [message] --fee-limit 10uband`,
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			data, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			content := tsstypes.NewTextSignatureOrder(data)

			coinStr, err := cmd.Flags().GetString(flagFeeLimit)
			if err != nil {
				return err
			}

			feeLimit, err := sdk.ParseCoinsNormalized(coinStr)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgRequestSignature(
				content,
				feeLimit,
				clientCtx.GetFromAddress().String(),
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// GetTxCmdActivate creates a CLI command for activate the sender.
func GetTxCmdActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activate",
		Args:    cobra.NoArgs,
		Short:   "activate the status of the address",
		Example: fmt.Sprintf(`%s tx bandtss activate`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			isIncomingGroup, err := cmd.Flags().GetBool(flagIncomingGroup)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			groupID, err := getGroupID(queryClient, isIncomingGroup)
			if err != nil {
				return err
			}

			msg := &types.MsgActivate{
				Sender:  clientCtx.GetFromAddress().String(),
				GroupID: groupID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().
		Bool(flagIncomingGroup, false, "Whether the activation is for the incoming group or current group.")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func getGroupID(queryClient types.QueryClient, isIncomingGroup bool) (tss.GroupID, error) {
	if isIncomingGroup {
		resp, err := queryClient.IncomingGroup(
			context.Background(),
			&types.QueryIncomingGroupRequest{},
		)
		if err != nil {
			return 0, err
		}

		return resp.GroupID, nil
	} else {
		resp, err := queryClient.CurrentGroup(
			context.Background(),
			&types.QueryCurrentGroupRequest{},
		)
		if err != nil {
			return 0, err
		}

		return resp.GroupID, nil
	}
}
