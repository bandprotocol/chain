package bandtss

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	bandtssv1beta1 "github.com/bandprotocol/chain/v3/api/band/bandtss/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: bandtssv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Counts",
					Use:       "counts",
					Short:     "Get current number of signing requests to bandtss module on BandChain",
				},
				{
					RpcMethod: "IsGrantee",
					Use:       "is-grantee [granter_address] [grantee_address]",
					Short:     "Query grantee status",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "granter"},
						{ProtoField: "grantee"},
					},
				},
				{
					RpcMethod: "Members",
					Use:       "members",
					Short:     "Query the members information",
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"is_incoming_group": {
							Name:  "incoming-group",
							Usage: "Whether the heartbeat is for the incoming group or current group.",
						},
						"status": {
							Name:  "status",
							Usage: "Filter members by status",
						},
					},
				},
				{
					RpcMethod: "Member",
					Use:       "member [address]",
					Short:     "Query the member by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod:      "CurrentGroup",
					Use:            "current-group",
					Short:          "Query the current group information",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod:      "IncomingGroup",
					Use:            "incoming-group",
					Short:          "Query the incoming group information",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod: "Signing",
					Use:       "signing [id]",
					Short:     "Query a signing by signing ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "signing_id"},
					},
				},
				{
					RpcMethod:      "GroupTransition",
					Use:            "group-transition",
					Short:          "Query the group transition information",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows parameter of the module",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: bandtssv1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
