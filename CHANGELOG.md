# Changelog

## [Unreleased]
- (app) Upgrade cosmos sdk version to v0.43.x
- (app) Use ibc-go package v1.0.0
- (app) Limit block gas to 8000000
- (bandd) Change max owasm gas to be the same as block gas limit (8000000)
- (bandd) Replace int64 with uint64 for ids and counts
- (bandd) Remove escrowAddress for IBC oracle request and use a given relayer account instead
- (bandd) Support oracle script functions - GetPrepareTime() and GetExecuteTime() - for retrieving prepare and execute blocktime respectively.
- (bandd) Refactor test

## [v2.1.1](https://github.com/bandprotocol/chain/releases/tag/v2.1.1))

- (bandd) Increase max block size for evidence size

## [v2.1.0](https://github.com/bandprotocol/chain/releases/tag/v2.1.0)

- (app) Adjust block params on init and migrate command
- (bandd) Bump SDK to 0.42.9 to resolve IBC channel restart issue (9800)[https://github.com/cosmos/cosmos-sdk/issues/9800].
- (yoda) Add retry logic when query data from node
- (bandd) Parameterized max data report size
