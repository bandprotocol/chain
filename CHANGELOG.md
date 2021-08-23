# Changelog

## [Unreleased]

- (yoda) Fix yoda to send report by MsgExec.
- (yoda) Add feature on yoda keys list to show grant status of reporter
- (chain) Remove MsgAddReporter/MsgRemoveReporter + Using Grant in authz module to manage authorization of reporter
- (chain) Refactor QueryRequestVerification
- (chore) Change max owasm gas to be the same as block gas limit (8000000)
- (chore) Limit block gas to 8000000
- (test) Refactor tests
- (ibc) Remove escrowAddress for IBC oracle request and use a given relayer account instead
- (chain) Replace int64 with uint64 for ids and counts
- (patch) Use ibc-go package v1.0.0
- (patch) Upgrade SDK version to v0.43.0
- (chain) Support oracle script functions - GetPrepareTime() and GetExecuteTime() - for retrieving prepare and execute blocktime respectively.

## [v2.1.1](https://github.com/bandprotocol/chain/releases/tag/v2.1.1)

- (bandd) Increase max block size for evidence size

## [v2.1.0](https://github.com/bandprotocol/chain/releases/tag/v2.1.0)

- (app) Adjust block params on init and migrate command
- (bandd) Bump SDK to 0.42.9 to resolve IBC channel restart SDK issue [9800](https://github.com/cosmos/cosmos-sdk/issues/9800).
- (yoda) Add retry logic when query data from node
- (bandd) Parameterized max data report size
