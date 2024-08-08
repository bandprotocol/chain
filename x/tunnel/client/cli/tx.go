package cli

import (
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
	txCmd.AddCommand(GetTxCmdCreateIBCTunnel())
	txCmd.AddCommand(GetTxCmdActivateTunnel())
	txCmd.AddCommand(GetTxCmdManualTriggerTunnel())

	return txCmd
}

func GetTxCmdCreateTSSTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-tss-tunnel [feed-type] [destination-chain-id] [destination-contract-address] [deposit] [signalInfos-json-file]",
		Short: "Create a new TSS tunnel",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			feedType, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			signalInfos, err := parseSignalInfos(args[4])
			if err != nil {
				return err
			}

			var route types.Route
			tssRoute := types.TSSRoute{
				DestinationChainID:         args[1],
				DestinationContractAddress: args[2],
			}
			route = &tssRoute

			msg, err := types.NewMsgCreateTunnel(
				signalInfos,
				feedstypes.FeedType(feedType),
				route,
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

func GetTxCmdCreateIBCTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-ibc-tunnel [feed-type] [channel-id] [deposit] [signalInfos-json-file]",
		Short: "Create a new IBC tunnel",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			feedType, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			signalInfos, err := parseSignalInfos(args[3])
			if err != nil {
				return err
			}

			var route types.Route
			ibcRoute := types.IBCRoute{
				ChannelID: args[1],
			}
			route = &ibcRoute

			msg, err := types.NewMsgCreateTunnel(
				signalInfos,
				feedstypes.FeedType(feedType),
				route,
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
		Use:   "activate-tunnel [id]",
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

func GetTxCmdManualTriggerTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manual-trigger-tunnel [id]",
		Short: "Manual trigger a tunnel",
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

			msg := types.NewMsgManualTriggerTunnel(id, clientCtx.GetFromAddress().String())
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
