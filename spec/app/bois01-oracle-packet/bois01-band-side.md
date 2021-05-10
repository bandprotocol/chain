---
bois: 01
title: Band Protocol Oracle Requests
stage: Strawman
category: IBC/APP
kind: TODO
author: Nathachai Jaiboon <nathachai@bandprotocol.com>
created: 2021-04-05
modified: 2021-04-05
# requires: (optional list of ics numbers)
# required-by: (optional list of ics numbers)
# implements: (optional list of ics numbers)
---

TODO: Change `bois01-1` to correct number

<https://hackmd.io/@ntchjb/rJQirQdS_>

## Synopsis

This standard document specifies packet data structure, state machine handling logic, and encoding details for requesting oracle data, which is done by separated chains, from Bandchain over IBC channel.


### Motivation

Band Protocol is the decentralized oracle data service that expected to provide as many ways as possible to allow users requesting oracle data from Bandchain network. The main way to request oracle data is to send a transaction directly to the network.

Since Bandchain is based on Cosmos SDK, IBC is another option to allow other chain be able to send transaction between separate chains. Therefore, IBC is one of the ways to request oracle data to improve integration for other chains who prefer to use data provided by Bandchain network.

### Desired Properties

- A user who knows request key is able to pay fee for requesting oracle data using the escrow address linked to the request key.
- A user should add tokens to the escrow address first to be used as fee.
- The chain that request oracle data should handle packet processing themselves to process result that later fulfilled the oracle request.

## Technical Specification

### Data Structures

There is only one packet data type, which is `OracleRequestPacketData`, which specify details of a oracle request. The following interface indicating data structure of `OracleRequestPacketData`, which is implemented as a Protobuf message.

```protobuf
message OracleRequestPacketData {
  string client_id = 1;
  int64 oracle_script_id = 2;
  bytes calldata = 3;
  uint64 ask_count = 4;
  uint64 min_count = 5;
  repeated cosmos.base.v1beta1.Coin fee_limit = 6;
  string request_key = 7;
  uint64 prepare_gas = 8;
  uint64 execute_gas = 9;
}
```

- `client_id` is the unique identifier of this oracle request, as specified by the client. This same unique ID will be sent back to the requester with the oracle response.
- `oracle_script_id` is the unique identifier of the oracle script to be executed.
- `calldata` is the OBI-encoded data as argument parameters for oracle script's executor to read. It is usually used to specify which type of data is requested and how the data is shaped when returned to the requester. However, call data can be anything depending on how the oracle script is implemented.
- `ask_count` is the number of validators that are requested to respond to this oracle request. Higher value means more security, at a higher gas cost.
- `min_count` is the minimum number of validators necessary for the request to be proceeded to the execution phase. Higher value means more security, at the cost of liveness.
- `fee_limit` is the maximum amount of tokens that will be paid to all data source providers for this request.
- `request_key` is the key from request chain to match data source fee payer on Bandchain. It is an arbitrary secret string that is used as a part of finding address of fee payer on Bandchain.
- `prepare_gas` is an amount of gas reserved for preparing raw requests done by the oracle script's executor. The gas will be manually consumed by the oracle script's executor.
- `execute_gas` is an amount of gas reserved for processing raw requests done by the oracle script's executor. The gas will be manually consumed by the oracle script's executor.

Acknowledge data type describes an unique identifier assigned to the submitted request. It can be used to keep tracking of request's results.

```protobuf
message OracleRequestPacketAcknowledgement {
  int64 request_id = 1;
}
```

- `request_id` is BandChain's unique identifier for an oracle request.

Once the request has been fulfilled, the result will be sent by Bandchain back to origin chain using the following package named `OracleResponsePacketData`. The origin chain should handle the message for further processing of the data.

```protobuf
message OracleResponsePacketData {
  string client_id = 1;
  int64 request_id = 2;
  uint64 ans_count = 3;
  int64 request_time = 4;
  int64 resolve_time = 5;
  ResolveStatus resolve_status = 6;
  bytes result = 7;
}

enum ResolveStatus {
  // Open - the request is not yet resolved.
  RESOLVE_STATUS_OPEN_UNSPECIFIED = 0
  // Success - the request has been resolved successfully with no errors.
  RESOLVE_STATUS_SUCCESS = 1
  // Failure - an error occured during the request's resolve call.
  RESOLVE_STATUS_FAILURE = 2
  // Expired - the request does not get enough reports from validator within the
  // timeframe.
  RESOLVE_STATUS_EXPIRED = 3
}
```

- `client_id` is the unique identifier of this oracle request, as specified by the requester.
- `request_id` is BandChain's unique identifier for this oracle request.
- `ans_count` is the number of validators among to the asked validators that have responded to this oracle request.
- `request_time` is the UNIX epoch time at which the request was sent to BandChain.
- `resolve_time` is the UNIX epoch time at which the request was resolved to the final result.
- `resolve_status` is the status of this oracle request, which can be OK, FAILURE, or EXPIRED.
- `result` is the final aggregated value encoded in OBI format. Only available if status is OK.

### Sub-protocols

This section describes initial setup and procedures done after received `OracleRequestPacketData` packet data.

#### Port & Channel Setup

In order to receive oracle requests via IBC, a port need to be binded for oracle module. So, the name of the port uses the same name as oracle module name, which is `oracle`. Also, module callbacks need to be registered during binding the port. After capability is returned by binded the port, it is needed to be claimed under the name `ports/oracle`. All of these operations is done in `InitGenesis` method.

```go
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	// ...

    // portKeeper is defined in github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/keeper
    capability := portKeeper.BindPort(ctx, "oracle")
    // scopedKeeper is defined in github.com/cosmos/cosmos-sdk/x/capability/keeper
    err := scopedKeeper.ClaimCapability(ctx, capability, "ports/oracle")
    
    // ...
}
```

Once the setup is done, new channels can be created via IBC routing module.

#### Routing Callbacks

Routing callback need to be setup to response on channel initialization and packet sent by other chain. The following section describes procedures of the routing callbacks.

##### OnChanOpenInit

```typescript
function OnChanOpenInit(
	order:          channeltypes.Order,
	connectionHops: [Identifier],
	portID:         Identifier,
	channelID:      Identifier,
	chanCap:        Capability,
	counterparty:   Counterparty,
	version:        string,
) {
  abortTransactionUnless(channelSequence <= MAX_UINT32)
  abortTransactionUnless(order === "UNORDERED")
  abortTransactionUnless(version === "bois01-1")
  abortTransactionUnless(portID === "oracle")

  claimCapability(chanCap, `capabilities/ports/${portID}/channel/${channelID}`)
}
```

`OnChanOpenInit` validates channel parameters before create the channel and claim capability that the module owned the channel associated with the port.

##### OnChanOpenTry

```typescript
function OnChanOpenTry() {

}
func (am AppModule) OnChanOpenTry(
	order: ChannelOrder,
  connectionHops: [Identifier],
  portIdentifier: Identifier,
  channelIdentifier: Identifier,
  counterpartyPortIdentifier: Identifier,
  counterpartyChannelIdentifier: Identifier,
  version: string,
  counterpartyVersion: string),
) {
  abortTransactionUnless(order === "UNORDERED")
  abortTransactionUnless(version === "bois01-1")
  abortTransactionUnless(portID === "oracle")
  abortTransactionUnless(counterpartyVersion === "bois01-1")
  
  const capabilityPath = `capabilities/ports/${portID}/channel/${channelID}`
  if (!isCapabilityAuthenticated(capabilityPath)) {
    claimCapability(chanCap, capabilityPath)
  }
}
```
`OnChanOpenTry` validates channel parameters the same way as `OnChanOpenInit` plus comparing countryparty version, and claim capability in case that `OnChainOpenInit` has not been run yet.


##### OnChanOpenAck

```typescript
function OnChanOpenAck(
  portIdentifier: Identifier,
  channelIdentifier: Identifier,
  version: string
) {
  abortTransactionUnless(version === "bois01-1")
}
```

For `OnChanOpenAck`, it only checks BOIS-01 version.

##### OnChanOpenConfirm

For `OnChanOpenConfirm`, there is no action necessary.

##### OnChanCloseInit

```go
function OnChanCloseInit(
  portIdentifier: Identifier,
  channelIdentifier: Identifier
) {
  alwaysAbort()
}
```

For `OnChanCloseInit`, it disallow user-initiated channel closing oracle channels.

##### OnChanCloseConfirm

For `OnChanCloseConfirm`. There is no action necessary.

##### OnRecvPacket

```typescript
function OnRecvPacket(
	packet: Packet
) {
  OracleRequestPacketData data = packet.data

  getEscrowAddress(data.RequestKey, packet.DestinationChannel, packet.DestinationPort)
  source = IBCSource{packet.DestinationChannel,  packet.DestinationPort}
  result = prepareRequest(data, escrowAddress, source)
	if result.err !== null {
		acknowledgement = NewErrorAcknowledgement(err.Error())
	} else {
		acknowledgement = NewResultAcknowledgement(NewOracleRequestPacketAcknowledgement(id))
	}
    
    return acknowledgement
}
```

For `OnRecvPacket`, it parses packet data and store it into `OracleRequestPacketData` variable. Then, calculate an escrow address based on `request_key` in the variable. After that, they are passed to a function called `prepareRequest` to start creating oracle request and preparing for fulfilling the request. `prepareRequest` is the same function that is run when sending `MsgRequestData` directly to Bandchain.

If the request is successfully prepared, then a request ID referred to the request will be attached to the acknowledgement, otherwise, an error is attached instead.

In order to create an escrow address based on channel ID, port ID, and request key, the following procedure is used

```go
func GetEscrowAddress(requestKey, portID, channelID string) sdk.AccAddress {
  contents := fmt.Sprintf("%s/%s/%s", requestKey, portID, channelID)

  preImage := []byte(Version)
  preImage = append(preImage, 0)
  preImage = append(preImage, contents...)
  hash := sha256.Sum256(preImage)

  return hash[:20]
}
```

The escrow address is the first 20 byte of SHA-256 hash of the preImage byte array, which is constructed from concatenation of BOIS-01 version and content string separated by a zero byte. The content string consists of request key, portID, and channel ID separated by slash.

Note that the algorithm stated above is based on [ADR-028](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-028-public-key-addresses.md).



##### OnAcknowledgementPacket

For `OnAcknowledgementPacket`. There is no action necessary.

##### OnTimeoutPacket

For `OnTimeoutPacket`. There is no action necessary.

#### Handling results of oracle requests

After a oracle request has been fulfilled with results, Bandchain sends a packet message named `OracleResponsePacketData` to the chain who sent the request. The following pseudocode describes steps to send the packet data to origin chain.

```typescript
function SaveResult(
  requestId: Identifier,
  resolveStatus: ResolveStatusEnum,
  blockTime: Time
  result: bytes
) {

  // ...
  
  request = getRequest(requestId)
  reportCount = getReportCount(requestId)
  
  if (request.isFromIbc()) {
    sourceChannel = request.ibc.sourceChannel
    sourcePort = request.ibc.sourcePort
    destinationChannel = sourceChannel.counterParty.channelId
    destinationPort = sourcePort.counterParty.portId
    sequence = channelKeeper.getNextSequenceSend(sourcePort, sourceChannel)
    channelCap = scopedKeeper.getCapability(`capabilities/ports/${sourcePort}/channel/${sourceChannel}`)

    packetData = newOracleResponsePacketData(
      request.clientId,
      requestId,
      reportCount,
      request.requestTime,
      blockTime,
      resolveStatus,
      result
    )
    packet = newPacket(
      packetData.getBytes(),
      sequence,
      sourcePort,
      sourceChannel,
      destinationPort,
      destinationChannel,
      BlockHeight{0, 0},
      blockTime.addMinutes(10) // The amount of time is not final. It can be changes in future.
    )
    
    channelKeeper.sendPacket(channelCap, packet)
  }
  
  // ...
}
```

Firstly, IDs of source port/channel and destination port/channel are prepared along with packet sequence number, block time, and other information used for creating a packet. Then, construct the packet data using `OracleResponsePacketData` and serialize it into byte array, and use the array to construct generalized packet data with other information that we have prepared. For packet timeout, the value can be changed in the future. Currently, we have planned to set the timeout time of the packet to `currentBlockTime + 10 minutes`. After generalized packet has been constructed, it is sent to other chain.

Therefore, the counterparty of Bandchain should have their own method to handle `OracleResponsePacketData` packets.

## Backwards Compatibility

Not applicable.

## Forwards Compatibility

This initial standard uses version "bois01-1" in the channel handshake.

A future version of this standard could use a different version in the channel handshake, and safely alter the packet data format & packet handler semantics.

## Example Implementation

- IBC routing callbacks of oracle module can be seen in [this link](https://github.com/bandprotocol/chain/blob/master/x/oracle/module.go).
- All protobuf messages can be found [here](https://github.com/bandprotocol/chain/blob/master/proto/oracle/v1).

Methods of handling `OracleResponsePacketData` in client chain can be done by following this [BOIS01-consumer-side](https://hackmd.io/@songwongtp/rye4QgYHO).

## Other Implementations

An implementation of `prepareRequest()` method can be found [here](https://github.com/bandprotocol/chain/blob/b9079719786048dfae805b1dbcad4160534f3fe2/x/oracle/keeper/owasm.go#L54).

## History

(changelog and notable inspirations / references)

## Copyright

All content herein is licensed under [Apache 2.0](https://www.apache.org/licenses/LICENSE-2.0).
