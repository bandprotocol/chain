package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/bandprotocol/chain/v3/x/restake/types"
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
		GetQueryCmdVaults(),
		GetQueryCmdVault(),
		GetQueryCmdLocks(),
		GetQueryCmdLock(),
		GetQueryCmdStake(),
		GetQueryCmdParams(),
	)

	return queryCmd
}

// GetQueryCmdVaults implements the vaults query command.
func GetQueryCmdVaults() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vaults",
		Short: "shows all vaults",
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

			res, err := queryClient.Vaults(cmd.Context(), &types.QueryVaultsRequest{Pagination: pageReq})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "vaults")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdVault implements the vault query command.
func GetQueryCmdVault() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault [key]",
		Short: "shows information of the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Vault(
				cmd.Context(),
				&types.QueryVaultRequest{
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

// GetQueryCmdLocks implements the locks query command.
func GetQueryCmdLocks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locks [staker_address]",
		Short: "shows all locks of an staker address",
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
				StakerAddress: args[0],
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
		Use:   "lock [staker_address] [key]",
		Short: "shows the lock of an staker address for the vault",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Lock(cmd.Context(), &types.QueryLockRequest{
				StakerAddress: args[0],
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

// GetQueryCmdStake implements the stake query command.
func GetQueryCmdStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stakes [staker_address]",
		Short: "shows all stakes of an staker address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Stake(cmd.Context(), &types.QueryStakeRequest{
				StakerAddress: args[0],
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

// GetQueryCmdParams implements the params query command.
func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Show params",
		Long:  "Show parameter of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
