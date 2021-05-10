---
bois: 01
title: Band Protocol Oracle Requests (consumer)
stage: Strawman
category: IBC/APP
kind: TODO
author: Songwong Tasneeyapant <songwong@bandprotocol.com>
created: 2021-04-06
modified: 2021-04-06
# requires: (optional list of ics numbers)
# required-by: (optional list of ics numbers)
# implements: (optional list of ics numbers)
---

`bois01-1` consumer side

<https://hackmd.io/@songwongtp/rye4QgYHO>

## Technical Specification

### Subprotocols

*Note*: change `portName` accordingly to your module's name

#### Port & Channel Setup

In order to send an oracle request and recieve an oracle response via IBC, a port needs to be binded to the module using a method of that module's keeper called `BindPort`. This is done in `InitGenesis` method of the module.

```go
func InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
    // ...

    // portKeeper is defined in github.com/cosmos/cosmos-sdk/x/ibc/core/05-port/keeper
    capability = portKeeper.bindPort(ctx, "portName")
    // scopedKeeper is defined in github.com/cosmos/cosmos-sdk/x/capability/keeper
    err = scopedKeeper.claimCapability(ctx, capability, "ports/portName")
    
    // ...
}
```

Once the module obtains its port, new channels are created via IBC routing module.

#### Routing Callbacks

Routing callbacks are responsible for handling channel initialization and relaying packets with a counterparty chain (bandchain in this case).

##### Channel Lifecycle Management


`OnChanOpenInit` validates channel parameters before claiming the newly created channel capability that is associated with the module port.
```javascript
function onChanOpenInit(
    order:          channeltypes.Order,
    connectionHops: [Identifier],
    portID:         Identifier,
    channelID:      Identifier,
    chanCap:        Capability,
    counterparty:   Counterparty,
    version:        string) {
    
    abortTransactionUnless(channelSequence <= MAX_UINT32)
    abortTransactionUnless(order === "UNORDERED")
    abortTransactionUnless(version === "bois01-1")
    abortTransactionUnless(portID === "portName")

    claimCapability(chanCap, `capabilities/ports/${portID}/channel/${channelID}`)
}
```

`OnChanOpenTry` validates channel parameters as `OnChanOpenInit` along with checking the counterpart version. Then it claims the channel capability if it has not done so.
```javascript 
function onChanOpenTry(
    order: ChannelOrder,
    connectionHops: [Identifier],
    portIdentifier: Identifier,
    channelIdentifier: Identifier,
    counterpartyPortIdentifier: Identifier,
    counterpartyChannelIdentifier: Identifier,
    version: string,
    counterpartyVersion: string) {
    
    abortTransactionUnless(order === "UNORDERED")
    abortTransactionUnless(version === "bois01-1")
    abortTransactionUnless(portID === "portName")
    abortTransactionUnless(counterpartyVersion === "bois01-1")

    const capabilityPath = `capabilities/ports/${portID}/channel/${channelID}`
    if (!isCapabilityAuthenticated(capabilityPath)) {
        claimCapability(chanCap, capabilityPath)
    }
}
```

`OnChanOpenAck` mainly checks BOIS-01 version. Additional commands can be added based on a module's specific necessities. 
```javascript
function onChanOpenAck(
    portIdentifier: Identifier,
    channelIdentifier: Identifier,
    version: string) {
    abortTransactionUnless(version === "bois01-1")
}
```

`OnChanOpenConfirm` is optional. Ones can modify it as they see fit.
```javascript
function onChanOpenConfirm(
    portIdentifier: Identifier,
    channelIdentifier: Identifier) {
    // no action necessary
}
```

`OnChanCloseInit` is optional. Ones can modify it as they see fit.
```javascript
function onChanCloseInit(
    portIdentifier: Identifier,
    channelIdentifier: Identifier) {
    // no action necessary
}
```

`OnChanCloseInit` is optional. Ones can modify it as they see fit.
```javascript 
function onChanCloseConfirm(
    portIdentifier: Identifier,
    channelIdentifier: Identifier) {
    // no action necessary
}
```

##### Handling oracle requests and responses

`CreateOracleRequestPacket` is a template for a method to create an oracle request packet. 
```javascript 
function createOracleRequestPacket(
    destPort:         string,
    destChannel:      string,
    sourcePort:       string,
    sourceChannel:    string,
    timeoutHeight:    Height,
    timeoutTimestamp: uint64,
    clientID:         string,
    oracleScript:     uint64,
    callData:         []byte,
    askCount:         uint64,
    minCount:         uint64,
    requestKey:       string,
    prepareGas:       uint64,
    executeGas:       uint64) {
    
    sequence = channelKeeper.getNextSequenceSend(sourcePort, sourceChannel)
    channelCap = scopedKeeper.getCapability(`capabilities/ports/${sourcePort}/channel/${sourceChannel}`)
    
    packetData := newOracleRequestPacketData(
        clientID, 
        oracleScript, 
        callData, 
        askCount, 
        minCount, 
        requestKey, 
        prepareGas, 
        executeGas, 
    )
    packet = newPacket(
        packetData.getBytes(),
        sequence,
        sourcePort,
        sourceChannel,
        destinationPort,
        destinationChannel,
        timeoutHeight,
        timeoutTimestamp,
    )
    channelKeeper.sendPacket(channelCap, packet)
}
```

`OnAcknowledgePacket` below is a template of how to extract the oracle request id from the returned acknowledgement. **Note**: This method is required by BOIS-01.
```javascript
function onAcknowledgePacket(
    packet:          Packet,
    acknowledgement: []byte) {
    
    Acknowledgement ack = acknowledgement
    
    if (ack.Response instanceof Acknowledgement_Result) {
        OracleRequestPacketAcknowledgement oracleAck = resp.Result
        // oracleAck.RequestID ...
    } else if (ack.Response instanceof Acknowledgement_Error) {
        // no action necessary
    } else {
        // no action necessary
    }
    // ...
}
```

`OnRecvPacket` below is a template to handle oracle response packet sent from bandchain. **Note**: This method is required by BOIS-01. `OracleResult` contains only a single field of type `uint64` named `price` which is the result from bandchain. 

```javascript
function onRecvPacket(
    packet: Packet) {
    
    OracleResponsePacketData data = packet.data
    string clientId = data.clientId 
    OracleResult res = data.result
    // clientId, res.price ...
    
    // acknowledgements are not necessary as bandchain ignores them
    acknowledgement = NewAcknowledgement()
    return acknowledgement
}
```
