package oracle

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	oraclev1 "github.com/bandprotocol/chain/v3/api/oracle/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: oraclev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Counts",
					Use:       "counts",
					Short:     "Get number of requests, oracle scripts, and data source scripts currently deployed on Bandchain",
				},
				{
					RpcMethod:      "Data",
					Use:            "data [data-hash]",
					Short:          "Get a content of the data source or oracle script for given SHA256 file hash",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "data_hash"}},
				},
				{
					RpcMethod:      "DataSource",
					Use:            "data-source [id]",
					Short:          "Get summary information of a data source",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "data_source_id"}},
				},
				{
					RpcMethod:      "OracleScript",
					Use:            "oracle-script [id]",
					Short:          "Get summary information of an oracle script",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "oracle_script_id"}},
				},
				{
					RpcMethod:      "Request",
					Use:            "request [id]",
					Short:          "Get an oracle request details",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "request_id"}},
				},
				{
					RpcMethod:      "PendingRequests",
					Use:            "pending-requests [validator-address]",
					Short:          "Get list of pending request IDs assigned to given validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}},
				},
				{
					RpcMethod:      "Validator",
					Use:            "validator [validator-address]",
					Short:          "Get active status of a validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}},
				},
				{
					RpcMethod: "IsReporter",
					Use:       "is-reporter [validator-address] [reporter-address]",
					Short:     "Check report grant of reporter",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_address"},
						{ProtoField: "reporter_address"},
					},
				},
				{
					RpcMethod:      "Reporters",
					Use:            "reporters [validator-address]",
					Short:          "Get an oracle request details",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "validator_address"}},
				},
				{
					RpcMethod: "ActiveValidators",
					Use:       "active-validators",
					Short:     "Get all active oracle validators",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Get current parameters of Bandchain's oracle module",
				},
				{
					RpcMethod: "RequestVerification",
					Use:       "verify-request [chain-id] [validator-addr] [request-id] [data-source-external-id] [reporter-pubkey] [reporter-signature-hex]",
					Short:     "Verify validity of pending oracle requests",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "chain_id"},
						{ProtoField: "validator"},
						{ProtoField: "request_id"},
						{ProtoField: "external_id"},
						{ProtoField: "reporter"},
						{ProtoField: "signature"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: oraclev1.Msg_ServiceDesc.ServiceName,
		},
	}
}
