package cli

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

const flagTunnelStatusFilter = "status"

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the tunnel module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetQueryCmdTunnels(),
		GetQueryCmdTunnel(),
		GetQueryCmdPackets(),
		GetQueryCmdPacket(),
		GetQueryCmdParams(),
	)

	return queryCmd
}

// GetQueryCmdTunnel implements the query tunnel command.
func GetQueryCmdTunnel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tunnel [tunnel-id]",
		Short: "Query the tunnel by tunnel id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			tunnelID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Tunnel(context.Background(), &types.QueryTunnelRequest{
				TunnelId: tunnelID,
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

// GetQueryCmdTunnels implements the query tunnels command.
func GetQueryCmdTunnels() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tunnels",
		Short: "Query all tunnels",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			statusFilterFlag, err := cmd.Flags().GetBool(flagTunnelStatusFilter)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			var statusFilter types.TunnelStatusFilter
			if !cmd.Flags().Changed(flagTunnelStatusFilter) {
				statusFilter = types.TUNNEL_STATUS_FILTER_UNSPECIFIED
			} else if statusFilterFlag {
				statusFilter = types.TUNNEL_STATUS_FILTER_ACTIVE
			} else {
				statusFilter = types.TUNNEL_STATUS_FILTER_INACTIVE
			}

			res, err := queryClient.Tunnels(context.Background(), &types.QueryTunnelsRequest{
				IsActive:   statusFilter,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool(flagTunnelStatusFilter, false, "Filter tunnels by active status")
	flags.AddPaginationFlagsToCmd(cmd, "tunnels")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdPackets implements the query packets command.
func GetQueryCmdPackets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packets [tunnel-id]",
		Short: "Query the packets of a tunnel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			tunnelID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Packets(cmd.Context(), &types.QueryPacketsRequest{
				TunnelId:   tunnelID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, "packets")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetQueryCmdPacket implements the query packet command.
func GetQueryCmdPacket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packet [tunnel-id] [nonce]",
		Short: "Query a packet by tunnel id and nonce",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			tunnelID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			nonce, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			res, err := queryClient.Packet(cmd.Context(), &types.QueryPacketRequest{
				TunnelId: tunnelID,
				Nonce:    nonce,
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

// GetQueryCmdParams implements the query params command.
func GetQueryCmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Shows the parameters of the module",
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
