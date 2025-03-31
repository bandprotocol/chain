# Changelog

## [v3.0.0](https://github.com/bandprotocol/chain/releases/tag/v3.0.0-rc2)

* (bump) Use go 1.22.3
* (bump) Use cosmos-sdk package v0.50.11 / ibc-go v8.5.2
* (bump) replace github.com/tendermint/tendermint by github.com/cometbft/cometbft v0.38.12
* (bump) migrated from gogo/protobuf to cosmos/gogoproto
* (bump) Use go-owasm v0.3.1
* (chain) Remove disablefeeless flag
* (chain) Remove old oracle ante
* (chain) Remove legacy migrate cmd
* (chain) Remove `gogoproto.goproto_stringer` option on param
* (chain) Return error on AllocationToken at beginblock instead of panic
* (chain) Migrate REST Endpoint to GRPC
* (chain) Change the allocation of remaining tokens from the community pool to the Proposer
* (chain) Add consensus module to handle managing Tendermint consensus parameters
* (chain) Add autocli for oracle module
* (chain) update oracle params to support cosmos-sdk 0.47+
* (chain) Add Global fee module to set the min-gas-price on protocol level (MinimumGasPrice param)
* (chain) Set MinimumGasPrice to 0.0025uband
* (chain) Use priority based on gas_prices of uband but MsgReport will have the highest priority
* (chain) Add TSS module
* (cylinder) Add Cylinder to provide a service for supporting TSS process
* (chain) Add Bandtss module
* (chain) Add Feeds module
* (grogu) Add Grogu to provide a service for supporting Feeds price submission
* (chain) Add Tunnel module
* (chain) Add Restake module
* (chain) Separate rolling seed from Oracle module to be a standalone Rolling seed module
* (chain) Update proof to support newly added module

## [v2.5.3](https://github.com/bandprotocol/chain/releases/tag/v2.5.3)

* (bump) Use cometbft v0.34.29
* (yoda) Get information of requests through endpoint instead of events

## [v2.5.2](https://github.com/bandprotocol/chain/releases/tag/v2.5.2)

* (bump) Use cosmos-sdk v0.45.16 / ibc-go v4.3.1 / cometbft v0.34.28

## [v2.5.1](https://github.com/bandprotocol/chain/releases/tag/v2.5.1)

* (bump) Use cosmos-sdk package v0.45.15 / tendermint v0.34.27
* (bump) Use go-owasm v0.2.3
* (chain) Support statically linked binary for bandd

## [v2.5.0](https://github.com/bandprotocol/chain/releases/tag/v2.5.0)

* (bump) Use cosmos-sdk package v0.45.14 / tendermint v0.34.26 / ibc-go v4.3.0
* (chain) add new rest paths to prepare for the moving from rest to grpc in 2.6.x

## [v2.4.1](https://github.com/bandprotocol/chain/releases/tag/v2.4.1)

* (bump) Use cosmos-sdk package v0.45.10 / tendermint v0.34.22 / ibc-go v3.3.1

## [v2.4.0](https://github.com/bandprotocol/chain/releases/tag/v2.4.0)

* (bump) Use go 1.19
* (bump) Use cosmos-sdk package v0.45.9 / tendermint v0.34.21 / ibc-go v3.3.0
* (bump) Use go-owasm v0.2.2
* (chain) Add ICA host module
* (chain) Add MaxDelay parameter for request verification query
* (chain) Add IsDelay parameter for request verification response
* (chain) Add snapshot extension for oracle module
* (chain) change DefaultBlockMaxGas to 50M
* (chain) change DefaultBaseRequestGas to 50k
* (chain) change multiplier of cosmos gas to owasm gas to 20M
* (yoda) Add BAND_DATA_SOURCE_ID in header
* (yoda) Update to broadcast transactions by sync mode

## [v2.3.3](https://github.com/bandprotocol/chain/releases/tag/v2.3.3)

* (yoda) Change severity of error when query log
* (bump) Use cosmos-sdk package v0.44.5 / tendermint v0.34.14 / ibc-go v1.1.5

## [v2.3.2](https://github.com/bandprotocol/chain/releases/tag/v2.3.2)

* (bump) Use cosmos-sdk package v0.44.2
* (yoda) Fix Yoda can't cache file

## [v2.3.0](https://github.com/bandprotocol/chain/releases/tag/v2.3.0)

* (bump) Use cosmos-sdk package v0.44.0
* (bump) Use ibc-go package v1.1.0

## [v2.2.0](https://github.com/bandprotocol/chain/releases/tag/v2.2.0)

* (bump) Use ibc-go package v1.0.1
* (chain) Replace report authorization with generic authorization
* (yoda) Fix yoda to send report by MsgExec.
* (yoda) Add feature on yoda keys list to show grant status of reporter
* (chain) Remove MsgAddReporter/MsgRemoveReporter + Using Grant in authz module to manage authorization of reporter
* (chain) Refactor QueryRequestVerification
* (chore) Change max owasm gas to be the same as block gas limit (8000000)
* (chore) Limit block gas to 8000000
* (test) Refactor tests
* (ibc) Remove escrowAddress for IBC oracle request and use a given relayer account instead
* (chain) Replace int64 with uint64 for ids and counts
* (patch) Use ibc-go package v1.0.0
* (patch) Upgrade SDK version to v0.43.0
* (chain) Support oracle script functions - GetPrepareTime() and GetExecuteTime() - for retrieving prepare and execute blocktime respectively.

## [v2.1.1](https://github.com/bandprotocol/chain/releases/tag/v2.1.1)

* (bandd) Increase max block size for evidence size

## [v2.1.0](https://github.com/bandprotocol/chain/releases/tag/v2.1.0)

* (app) Adjust block params on init and migrate command
* (bandd) Bump SDK to 0.42.9 to resolve IBC channel restart SDK issue [9800](https://github.com/cosmos/cosmos-sdk/issues/9800).
* (yoda) Add retry logic when query data from node
* (bandd) Parameterized max data report size
