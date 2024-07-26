package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/restake/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the restake module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetQueryCmdKey(),
		GetQueryCmdKeys(),
		GetQueryCmdRewards(),
		GetQueryCmdReward(),
		GetQueryCmdLocks(),
		GetQueryCmdLock(),
	)

	return queryCmd
}

func GetQueryCmdKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key [name]",
		Short: "shows information of the key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Key(
				context.Background(),
				&types.QueryKeyRequest{
					Key: args[0],
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "shows all keys",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Keys(context.Background(), &types.QueryKeysRequest{Pagination: pageReq})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "keys")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards [locker_address]",
		Short: "shows all rewards of an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Rewards(context.Background(), &types.QueryRewardsRequest{
				LockerAddress: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "rewards")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward [locker_address] [key_name]",
		Short: "shows the reward of an locker address for the key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Reward(context.Background(), &types.QueryRewardRequest{
				LockerAddress: args[0],
				Key:           args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdLocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locks [locker_address]",
		Short: "shows all locks of an locker address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Locks(context.Background(), &types.QueryLocksRequest{
				LockerAddress: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "locks")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetQueryCmdLock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock [locker_address] [key_name]",
		Short: "shows the lock of an locker address for the key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Lock(context.Background(), &types.QueryLockRequest{
				LockerAddress: args[0],
				Key:           args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
