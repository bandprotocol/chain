package tunnel

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	tunnelv1beta1 "github.com/bandprotocol/chain/v3/api/band/tunnel/v1beta1"
)

// AutoCLIOptions returns the AutoCLI options for the tunnel module
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: tunnelv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the tunnel module",
				},
				{
					RpcMethod: "Tunnels",
					Use:       "tunnels",
					Short:     "Query all tunnels",
				},
				{
					RpcMethod: "Tunnel",
					Use:       "tunnel [tunnel-id]",
					Short:     "Query a specific tunnel by ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "tunnel_id"},
					},
				},
				{
					RpcMethod: "Deposits",
					Use:       "deposits [tunnel-id]",
					Short:     "Query all deposits of a tunnel",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "tunnel_id"},
					},
				},
				{
					RpcMethod: "Deposit",
					Use:       "deposit [tunnel-id] [depositor]",
					Short:     "Query a specific deposit by tunnel ID and depositor address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "tunnel_id"},
						{ProtoField: "depositor"},
					},
				},
				{
					RpcMethod: "Packets",
					Use:       "packets [tunnel-id]",
					Short:     "Query all packets of a tunnel",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "tunnel_id"},
					},
				},
				{
					RpcMethod: "Packet",
					Use:       "packet [tunnel-id] [sequence]",
					Short:     "Query a specific packet by tunnel ID and sequence",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "tunnel_id"},
						{ProtoField: "sequence"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: tunnelv1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
