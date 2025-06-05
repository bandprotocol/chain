# Changelog

## [v3.0.0](https://github.com/bandprotocol/chain/releases/tag/v3.0.0)

* (bump) Use Go 1.24.2
* (bump) Use Cosmos SDK v0.50.13 / IBC-Go v8.7.0
* (bump) Replace github.com/tendermint/tendermint with github.com/cometbft/cometbft v0.38.17
* (bump) Migrate from gogo/protobuf to cosmos/gogoproto
* (bump) Use go-owasm v0.3.1
* (chain) Remove --disablefeeless flag
* (chain) Remove old oracle ante
* (chain) Remove legacy migrate command
* (chain) Remove gogoproto.goproto_stringer option on param
* (chain) Return error on AllocationToken at BeginBlock instead of panic
* (chain) Migrate REST endpoints to GRPC
* (chain) Change the allocation of remaining tokens from the community pool to the proposer
* (chain) Add consensus module to manage Tendermint consensus parameters
* (chain) Add AutoCLI for oracle module
* (chain) Update oracle params to support Cosmos SDK 0.47+
* (chain) Add GlobalFee module to define protocol-level MinimumGasPrice
* (chain) Set MinimumGasPrice to 0.0025uband
* (chain) Add tss module
* (chain) Add bandtss module
* (chain) Add feeds module
* (chain) Add tunnel module (supports TSS, IBC, IBC Hook, Router, and Axelar routes)
* (chain) Add restake module
* (chain) Separate rolling seed logic into standalone module
* (chain) Update proof system to support new modules
* (chain) Introduce new mempool for flexible transaction processing and block proposal
* (chain) Add transaction prioritization with lanes for feeds, TSS, oracle reports, oracle requests, and default
* (grogu) Add Grogu service to support feed price submission
* (cylinder) Add Cylinder service to support TSS process
* (yoda) Add auto gas bump logic

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
