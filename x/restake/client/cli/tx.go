package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// GetTxCmd returns a root CLI command handler for all x/restake transaction commands.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Restake transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetTxCmdClaimRewards(),
		GetTxCmdAddRewards(),
		GetTxCmdLockPower(),
		GetTxCmdDeactivateKey(),
	)

	return txCmd
}

// GetTxCmdClaimRewards creates a CLI command for claming rewards
func GetTxCmdClaimRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards",
		Short: "Claim rewards",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaimRewards(
				clientCtx.GetFromAddress(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdLockPower creates a CLI command for locking power
func GetTxCmdLockPower() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-power [key] [amount]",
		Short: "lock power to the key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			key := args[0]
			amount, ok := sdk.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			msg := types.NewMsgLockPower(
				clientCtx.GetFromAddress(),
				key,
				amount,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdAddRewards creates a CLI command for adding rewards
func GetTxCmdAddRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-rewards [key] [coins]",
		Short: "Add rewards to the key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			key := args[0]
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgAddRewards(
				clientCtx.GetFromAddress(),
				key,
				coins,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetTxCmdDeactivateKey creates a CLI command for deactivating key
func GetTxCmdDeactivateKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate [key]",
		Short: "Deactivate key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			key := args[0]

			msg := types.NewMsgDeactivateKey(
				clientCtx.GetFromAddress(),
				key,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
