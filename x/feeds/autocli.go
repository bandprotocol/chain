package feeds

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	feedsv1beta1 "github.com/bandprotocol/chain/v3/api/band/feeds/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: feedsv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "AllCurrentPrices",
					Use:       "all-current-prices",
					Short:     "Get a list of all current prices of supported feeds",
				},
				{
					RpcMethod: "CurrentFeeds",
					Use:       "current-feeds",
					Short:     "Get a list of all currently supported feeds",
				},
				{
					RpcMethod:      "CurrentPrices",
					Use:            "current-prices [signal-ids]",
					Short:          "Get current prices for the provided signal IDs",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "signal_ids"}},
				},
				{
					RpcMethod: "IsFeeder",
					Use:       "is-feeder [validator-address] [feeder-address]",
					Short:     "Check if the given account is a feeder for the validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
						{ProtoField: "feeder_address"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get all the parameters of the feeds module",
				},
				{
					RpcMethod:      "Price",
					Use:            "price [signal-id]",
					Short:          "Get the price for a given signal ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "signal_id"}},
				},
				{
					RpcMethod: "Prices",
					Use:       "prices",
					Short:     "Get all prices",
				},
				{
					RpcMethod: "ReferenceSourceConfig",
					Use:       "reference-source-config",
					Short:     "Get the configuration of the reference price source",
				},
				{
					RpcMethod: "SignalTotalPowers",
					Use:       "signal-total-powers",
					Short:     "Get total powers for all signals or specific ones",
				},
				{
					RpcMethod:      "ValidValidator",
					Use:            "valid-validator [validator-address]",
					Short:          "Check if the validator is valid to send prices",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}},
				},
				{
					RpcMethod:      "ValidatorPrices",
					Use:            "validator-prices [validator-address]",
					Short:          "Get prices submitted by a validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}},
				},
				{
					RpcMethod:      "Vote",
					Use:            "vote [voter]",
					Short:          "Get signals submitted by a voter",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "voter"}},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: feedsv1beta1.Msg_ServiceDesc.ServiceName,
		},
	}
}
