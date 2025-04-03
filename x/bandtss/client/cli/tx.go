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
	flagExpiration = "expiration"
	flagFeeLimit   = "fee-limit"
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
	cmdRequestSignature.AddCommand(cmdTextRequestSignature)

	// Loop through and add the provided request signature commands as subcommands.
	for _, cmd := range requestSignatureCmds {
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
		Short: "Request signature from the group",
	}

	cmd.PersistentFlags().String(flagFeeLimit, "", "The maximum tokens that will be paid for this request")

	_ = cmd.MarkPersistentFlagRequired(flagFeeLimit)

	return cmd
}

// GetTxCmdTextRequestSignature creates a CLI command for create a signature on text message.
func GetTxCmdTextRequestSignature() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "text [message]",
		Args:  cobra.ExactArgs(1),
		Short: "Request signature of the message from the current group",
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
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetTxCmdActivate creates a CLI command for activate the sender.
func GetTxCmdActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activate",
		Args:    cobra.NoArgs,
		Short:   "Activate the status of the address",
		Example: fmt.Sprintf(`%s tx bandtss activate`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupIDs, err := getActivatingGroupIDs(clientCtx)
			if err != nil {
				return err
			}
			if len(groupIDs) == 0 {
				return fmt.Errorf("no group to activate")
			}

			msgs := make([]sdk.Msg, len(groupIDs))
			for i, groupID := range groupIDs {
				msgs[i] = &types.MsgActivate{
					Sender:  clientCtx.GetFromAddress().String(),
					GroupID: groupID,
				}
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// getActivatingGroupIDs returns the group IDs that the sender should activate.
func getActivatingGroupIDs(clientCtx client.Context) ([]tss.GroupID, error) {
	queryClient := types.NewQueryClient(clientCtx)

	// Get the member information in both current and incoming group.
	memberResp, err := queryClient.Member(
		context.Background(),
		&types.QueryMemberRequest{
			Address: clientCtx.GetFromAddress().String(),
		},
	)
	if err != nil {
		return nil, err
	}
	memberInfos := []types.Member{
		memberResp.CurrentGroupMember,
		memberResp.IncomingGroupMember,
	}

	// Get penalty duration from the params
	paramResp, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}
	penaltyDuration := paramResp.Params.InactivePenaltyDuration

	// Get the time when the penalty will expire.
	status, err := clientCtx.Client.Status(context.Background())
	if err != nil {
		return nil, err
	}
	latestBlockTime := status.SyncInfo.LatestBlockTime

	groupTypeText := []string{"current group", "incoming group"}

	// Check what group that the member should activate.
	var activatingGroupIDs []tss.GroupID
	for i, info := range memberInfos {
		displayedText := ""

		if info.Address == "" {
			displayedText = "skip; not belong to this group"
		} else if info.IsActive {
			displayedText = "skip; member is already active"
		} else if info.Since.Add(penaltyDuration).After(latestBlockTime) {
			displayedText = "skip; penalty not expired"
		} else {
			activatingGroupIDs = append(activatingGroupIDs, info.GroupID)
			displayedText = "activating"
		}

		fmt.Printf("checking %s: %s\n", groupTypeText[i], displayedText)
	}

	fmt.Println() // extra-newline for better readability

	return activatingGroupIDs, nil
}
