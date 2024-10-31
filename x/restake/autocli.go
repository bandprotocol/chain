package restake

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	restakev1beta1 "github.com/bandprotocol/chain/v3/api/band/restake/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: restakev1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Vaults",
					Use:       "vaults",
					Short:     "Shows all vaults",
				},
				{
					RpcMethod: "Vault",
					Use:       "vault [key]",
					Short:     "Shows information of the vault",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "key"},
					},
				},
				{
					RpcMethod: "Locks",
					Use:       "locks [staker-address]",
					Short:     "shows all locks of an staker address",
				},
				{
					RpcMethod: "Lock",
					Use:       "lock [staker-address] [key]",
					Short:     "Shows the lock of an staker address for the vault",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "staker_address"},
						{ProtoField: "key"},
					},
				},
				{
					RpcMethod: "Stake",
					Use:       "stake [staker-address]",
					Short:     "Shows all stakes of an staker address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "staker_address"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows parameter of the module",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: restakev1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
