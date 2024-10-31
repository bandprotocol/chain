package tss

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	tssv1beta1 "github.com/bandprotocol/chain/v3/api/band/tss/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: tssv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Counts",
					Use:       "counts",
					Short:     "Get current number of groups and signings on BandChain",
				},
				{
					RpcMethod: "Group",
					Use:       "group [id]",
					Short:     "Query group by group ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
				},
				{
					RpcMethod: "Groups",
					Use:       "groups",
					Short:     "Query a list of groups information",
				},
				{
					RpcMethod: "Members",
					Use:       "members [group-id]",
					Short:     "Query members by group id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "group_id"},
					},
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
					RpcMethod: "DE",
					Use:       "de-list [address]",
					Short:     "Query all DE for this address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "PendingGroups",
					Use:       "pending-groups [address]",
					Short:     "Query all pending groups for this address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "PendingSignings",
					Use:       "pending-signings [address]",
					Short:     "Query all pending signings for this address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
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
					RpcMethod: "Signings",
					Use:       "signings",
					Short:     "Query a list of signings information",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows parameter of the module",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: tssv1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
