package cli

import (
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
		GetQueryCmdKeys(),
		GetQueryCmdKey(),
		GetQueryCmdRewards(),
		GetQueryCmdReward(),
		GetQueryCmdLock(),
		GetQueryCmdLocks(),
	)

	return queryCmd
}

// GetQueryCmdKeys implements the keys query command.
func GetQueryCmdKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "shows all keys",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Keys(cmd.Context(), &types.QueryKeysRequest{Pagination: pageReq})
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

// GetQueryCmdKey implements the key query command.
func GetQueryCmdKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key [name]",
		Short: "shows information of the key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Key(
				cmd.Context(),
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

// GetQueryCmdRewards implements the rewards query command.
func GetQueryCmdRewards() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rewards [locker_address]",
		Short: "shows all rewards of an address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Rewards(cmd.Context(), &types.QueryRewardsRequest{
				LockerAddress: args[0],
				Pagination:    pageReq,
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

// GetQueryCmdReward implements the reward query command.
func GetQueryCmdReward() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward [locker_address] [key_name]",
		Short: "shows the reward of an locker address for the key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Reward(cmd.Context(), &types.QueryRewardRequest{
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

// GetQueryCmdLocks implements the locks query command.
func GetQueryCmdLocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locks [locker_address]",
		Short: "shows all locks of an locker address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Locks(cmd.Context(), &types.QueryLocksRequest{
				LockerAddress: args[0],
				Pagination:    pageReq,
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

// GetQueryCmdLock implements the lock query command.
func GetQueryCmdLock() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock [locker_address] [key_name]",
		Short: "shows the lock of an locker address for the key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Lock(cmd.Context(), &types.QueryLockRequest{
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
