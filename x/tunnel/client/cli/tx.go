package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

// GetTxCmd returns a root CLI command handler for all x/tunnel transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "tunnel transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(GetTxCmdCreateTSSTunnel())
	txCmd.AddCommand(GetTxCmdActivateTunnel())
	txCmd.AddCommand(GetTxCmdDeactivateTunnel())
	txCmd.AddCommand(GetTxCmdManualTriggerTunnel())

	return txCmd
}

func GetTxCmdCreateTSSTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-tss-tunnel [interval] [feed-type] [destination-chain-id] [destination-contract-address] [deposit] [signalInfos-json-file]",
		Short: "Create a new TSS tunnel",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			interval, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			feedTypeCli, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}
			if feedTypeCli <= 0 || feedTypeCli >= int64(len(feedstypes.FeedType_value)) {
				return fmt.Errorf("invalid feed type; got %d", feedTypeCli)
			}
			feedType := feedstypes.FeedType(feedTypeCli)

			destChainID := args[2]
			destContractAddr := args[3]

			deposit, err := sdk.ParseCoinsNormalized(args[4])
			if err != nil {
				return err
			}

			signalInfos, err := parseSignalInfos(args[5])
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateTSSTunnel(
				signalInfos.ToSignalInfos(),
				interval,
				feedType,
				destChainID,
				destContractAddr,
				deposit,
				clientCtx.GetFromAddress(),
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

func GetTxCmdActivateTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate-tunnel [tunnel-id]",
		Short: "Activate a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgActivateTunnel(id, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdDeactivateTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-tunnel [tunnel-id]",
		Short: "Deactivate a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgDeactivateTunnel(id, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetTxCmdManualTriggerTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manual-trigger-tunnel [tunnel-id]",
		Short: "Manual trigger a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			tunnelID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgManualTriggerTunnel(tunnelID, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
