# Changelog

## v2.1.2
- (bandd) Support oracle script functions - GetPrepareTime() and GetExecuteTime() - for retrieving prepare and execute blocktime respectively.

## v2.1.1

- (bandd) Increase max block size for evidence size

## v2.1.0

- (app) Adjust block params on init and migrate command
- (bandd) Bump SDK to 0.42.9 to resolve IBC channel restart issue (9800)[https://github.com/cosmos/cosmos-sdk/issues/9800].
- (yoda) Add retry logic when query data from node
- (bandd) Parameterized max data report size
