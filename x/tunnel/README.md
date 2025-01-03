# `x/tunnel`

## Abstract

The Tunnel module is designed to decentralize the creation of push-based price data packets by enabling users to configure intervals and deviations for transmitting price data. It ensures secure and efficient transmission to Ethereum Virtual Machine (EVM)-compatible blockchains, Cosmos-based blockchains, and other blockchain networks. Users have the flexibility to select their preferred transmission route, such as the Inter-Blockchain Communication (IBC) protocol or bridge technologies that integrate with Band Protocol. This versatility allows for seamless and reliable delivery of price data across diverse blockchain ecosystems.

## Contents

- [`x/tunnel`](#xtunnel)
  - [Abstract](#abstract)
  - [Contents](#contents)
  - [Concepts](#concepts)
    - [Tunnel](#tunnel)
    - [Route](#route)
      - [IBC Route](#ibc-route)
      - [TSS Route](#tss-route)
    - [Packet](#packet)
      - [Packet Generation Workflow](#packet-generation-workflow)
  - [State](#state)
    - [TunnelCount](#tunnelcount)
    - [TotalFee](#totalfee)
    - [ActiveTunnelID](#activetunnelid)
    - [Tunnel](#tunnel-1)
    - [Packet](#packet-1)
    - [LatestPrices](#latestprices)
    - [Deposit](#deposit)
    - [Params](#params)
  - [Msg](#msg)
    - [MsgCreateTunnel](#msgcreatetunnel)
    - [MsgUpdateRoute](#msgupdateroute)
    - [MsgUpdateSignalsAndInterval](#msgupdatesignalsandinterval)
    - [MsgWithdrawFeePayerFunds](#msgwithdrawfeepayerfunds)
    - [MsgActivate](#msgactivate)
    - [MsgDeactivate](#msgdeactivate)
    - [MsgTriggerTunnel](#msgtriggertunnel)
    - [MsgDepositToTunnel](#msgdeposittotunnel)
    - [MsgWithdrawFromTunnel](#msgwithdrawfromtunnel)
  - [Events](#events)
    - [Event: `create_tunnel`](#event-create_tunnel)
    - [Event: `update_signals_and_interval`](#event-update_signals_and_interval)
    - [Event: `activate_tunnel`](#event-activate_tunnel)
    - [Event: `deactivate_tunnel`](#event-deactivate_tunnel)
    - [Event: `trigger_tunnel`](#event-trigger_tunnel)
    - [Event: `produce_packet_fail`](#event-produce_packet_fail)
    - [Event: `produce_packet_success`](#event-produce_packet_success)
    - [Event: `deposit_to_tunnel`](#event-deposit_to_tunnel)
    - [Event: `withdraw_from_tunnel`](#event-withdraw_from_tunnel)
  - [Clients](#clients)
    - [CLI Commands](#cli-commands)
      - [Query Commands](#query-commands)
        - [List All Tunnels](#list-all-tunnels)
        - [Get Tunnel by ID](#get-tunnel-by-id)
        - [Get Deposits for a Tunnel](#get-deposits-for-a-tunnel)
        - [Get Deposit by Depositor](#get-deposit-by-depositor)
        - [List All Packets for a Tunnel](#list-all-packets-for-a-tunnel)
        - [Get Packet by Sequence](#get-packet-by-sequence)
        - [Get Total Fees](#get-total-fees)

## Concepts

### Tunnel

The `x/tunnel` module defines a `Tunnel` type that specifies details such as the way to send the data to the destination ([Route](#route)), the type of price data to encode, the address of the fee payer responsible for covering packet fees, and the total deposit of the tunnel (which must meet a minimum requirement to activate) and the interval and deviation settings applied to price data during packet production at each end-block.

Users can create a new tunnel by submitting a `MsgCreateTunnel` message to BandChain, specifying the desired signals, deviations, interval, and the route to which the data should be sent. The available routes for tunnels are detailed in the [Route](#route) section.

The Tunnel type represents a structure with the following fields:

```go
type Tunnel struct {
    // ID is the unique identifier of the tunnel.
    ID uint64
    // Sequence represents the sequence number of the tunnel packets.
    Sequence uint64
    // Route defines the path for delivering the signal prices.
    Route *types1.Any
    // FeePayer is the address responsible for paying the packet fees.
    FeePayer string
    // SignalDeviations is a list of signal deviations.
    SignalDeviations []SignalDeviation
    // Interval determines how often the signal prices are delivered.
    Interval uint64
    // TotalDeposit is the total amount of deposit in the tunnel.
    TotalDeposit sdk.Coins
    // IsActive indicates whether the tunnel is active.
    IsActive bool
    // CreatedAt is the timestamp when the tunnel was created.
    CreatedAt int64
    // Creator is the address of the tunnel's creator.
    Creator string
}
```

### Route

A Route defines the secure method for transmitting price data to a destination chain using a tunnel. It specifies the pathway and protocols that ensure safe and reliable data delivery from BandChain to other EVM-compatible chains or Cosmos-based blockchains.

The Route must be implemented using the RouteI interface to ensure compatibility and functionality

```go
type RouteI interface {
  proto.Message

  ValidateBasic() error
}
```

#### IBC Route

The IBC Route enables the transmission of data from BandChain to Cosmos-compatible chains via the Inter-Blockchain Communication (IBC) protocol. This route allows for secure and efficient cross-chain communication, leveraging the standardized IBC protocol to transmit packets of data between chains.

We also provide a library, cw-band, that enables the use of the Tunnel via WASM contracts on the destination chain. You can find an example and further details here: [cw-band](https://github.com/bandprotocol/cw-band)

To create an IBC tunnel, use the following CLI command:

> **Note**: You must create a tunnel before establishing an IBC connection using the tunnel ID. For example, if you create a tunnel and receive tunnelID 1, then create a channel with the port: `tunnel.1`.

> **Note**: An example of the signalInfos-json-file can be found at scripts/tunnel/signal_deviations.json.

```bash
bandd tx tunnel create-tunnel ibc [initial-deposit] [interval] [signalInfos-json-file]
```

#### TSS Route

The TSS Route enables the tunnel to send data securely from BandChain to destination chain using a TSS (Threshold Signature Scheme) signature. This approach ensures secure data signing within a decentralized network.

The tunnel requests the BandTSS module to sign the tunnel packet. Once the signing process is complete, a relayer captures the signed message and relays it to the destination chain. The destination chain can also verify that the data originates from BandChain without any modifications, ensuring data integrity.

To create a TSS tunnel, use the following CLI command:

```bash
bandd tx tunnel create-tunnel tss [destination-chain-id] [destination-contract-address] [encoder] [initial-deposit] [interval] [signalDeviations-json-file]
```

### Packet

A Packet represents the signal price data produced at the end of a block, based on the interval and deviation configured by the tunnel's creator. This data is then sent to the destination according to the specified route.

The Packet type represents a structure with the following fields:

```go
type Packet struct {
  // tunnel_id is the tunnel ID
  TunnelID uint64
  // sequence is representing the sequence of the tunnel packet.
  Sequence uint64
  // prices is the list of prices information from feeds module.
  Prices []feedstypes.Price
  // receipt represents the confirmation of the packet delivery to the destination via the specified route.
  Receipt *codectypes.Any
  // base_fee is the base fee of the packet
  BaseFee sdk.Coins
  // route_fee is the route fee of the packet
  RouteFee sdk.Coins
  // created_at is the timestamp when the packet is created
  CreatedAt int64
}
```

#### Packet Generation Workflow

At the end of each block, the tunnel generates packets by evaluating the deviations and intervals for all tunnels. The system uses two types of deviations: hard deviations and soft deviations, which are explained in detail below.

If any signal exceeds the hard deviation threshold, it is appended to the list of signals to be sent. Additionally, if any signal meets its soft deviation criteria while another signal surpasses the hard deviation threshold, that signal is also added to the list.

This mechanism is designed to optimize transaction efficiency on the destination route, particularly during periods of market instability, by reducing the number of unnecessary transactions.

## State

### TunnelCount

Stores the number of tunnels existing on the chain.

- **TunnelCount**: `0x00 | -> BigEndian(count)`

### TotalFee

Stores the total fees collected by tunnels when producing packets.

- **TotalFee**: `0x01 | -> TotalFee`

### ActiveTunnelID

Stores the IDs of active tunnels for quick querying at the end of a block.

- **ActiveTunnelID**: `0x10 | TunnelID -> []byte{0x01}`

### Tunnel

Stores information about each tunnel.

- **Tunnel**: `0x11 | TunnelID -> Tunnel`

### Packet

Stores information about packets sent via the routes declared in tunnels.

- **Packet**: `0x12 | TunnelID | Sequence -> Packet`

### LatestPrices

Stores the latest prices that the tunnel has sent to the destination route. These are used to compare intervals and deviations at the end of each block.

- **LatestPrices**: `0x13 | TunnelID -> LatestPrices`

### Deposit

Stores the total deposit per tunnel by each depositor.

- **Deposit**: `0x14 | TunnelID | DepositorAddress -> Deposit`

### Params

Stores the parameters in the state. These parameters can be updated via a governance proposal or by an authority address.

- **Params**: `0x90 -> Params`

The `x/tunnel` module contains the following parameters:

```go
type Params struct {
  // min_deposit is the minimum deposit required to create a tunnel.
  MinDeposit sdk.Coins
  // min_interval is the minimum interval in seconds.
  MinInterval uint64
  // max_interval is the maximum interval in seconds.
  MaxInterval uint64
  // min_deviation_bps is the minimum deviation in basis points.
  MinDeviationBPS uint64
  // max_deviation_bps is the maximum deviation in basis points.
  MaxDeviationBPS uint64
  // max_signals defines the maximum number of signals allowed per tunnel.
  MaxSignals uint64
  // base_packet_fee is the base fee for each packet.
  BasePacketFee sdk.Coins
```

## Msg

In this section, we describe the processing of the tunnel messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](#state) section.

### MsgCreateTunnel

```protobuf
// MsgCreateTunnel is the transaction message to create a new tunnel.
message MsgCreateTunnel {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgCreateTunnel";

  // signal_deviations is the list of signal deviations.
  repeated SignalDeviation signal_deviations = 1 [(gogoproto.nullable) = false];
  // interval is the interval for delivering the signal prices.
  uint64 interval = 2;
  // route is the route for delivering the signal prices
  google.protobuf.Any route = 3 [(cosmos_proto.accepts_interface) = "RouteI"];
  // initial_deposit is the deposit value that must be paid at tunnel creation.
  repeated cosmos.base.v1beta1.Coin initial_deposit = 4 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (amino.dont_omitempty)   = true
  ];
  // creator is the address of the creator.
  string creator = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

- **Deviation and Interval Settings**: Each tunnel must specify the deviation per signal and the interval per tunnel.
- **Route Selection**: Only one route can be chosen per tunnel.
- **Initial Deposit**: The initial deposit can be set to zero. Other users can contribute to the tunnel's deposit by send [MsgDepositToTunnel](#msgdeposittotunnel) message until it reaches the required minimum deposit.

### MsgUpdateRoute

To update the route details based on the route type, allowing certain arguments to be updated.

```protobuf
// MsgUpdateRoute is the transaction message to update a route tunnel
message MsgUpdateRoute {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgUpdateRoute";

  // tunnel_id is the ID of the tunnel to edit.
  uint64 tunnel_id = 1 [(gogoproto.customname) = "TunnelID"];
  // route is the route for delivering the signal prices
  google.protobuf.Any route = 2 [(cosmos_proto.accepts_interface) = "RouteI"];
  // creator is the address of the creator.
  string creator = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgUpdateSignalsAndInterval

Allows the creator of a tunnel to update the list of signal deviations and the interval for the tunnel.

```protobuf
// MsgUpdateSignalsAndInterval is the transaction message to update signals and interval of the tunnel.
message MsgUpdateSignalsAndInterval {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgUpdateSignalsAndInterval";

  // tunnel_id is the ID of the tunnel to edit.
  uint64 tunnel_id = 1 [(gogoproto.customname) = "TunnelID"];
  // signal_deviations is the list of signal deviations.
  repeated SignalDeviation signal_deviations = 2 [(gogoproto.nullable) = false];
  // interval is the interval for delivering the signal prices.
  uint64 interval = 3;
  // creator is the address of the creator.
  string creator = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgWithdrawFeePayerFunds

Allows the creator of a tunnel to withdraw funds from the fee payer to the creator.

```protobuf
// MsgWithdrawFeePayerFunds is the transaction message to withdraw fee payer funds to creator.
message MsgWithdrawFeePayerFunds {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgWithdrawFeePayerFunds";

  // tunnel_id is the ID of the tunnel to withdraw fee payer coins.
  uint64 tunnel_id = 1 [(gogoproto.customname) = "TunnelID"];
  // amount is the coins to withdraw.
  repeated cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (amino.dont_omitempty)   = true
  ];
  // creator is the address of the creator.
  string creator = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgActivate

To activate the tunnel for processing at the EndBlock, the following conditions must be met:

1. The total deposit must exceed the minimum deposit.
2. The fee payer must have sufficient Band tokens in their account to cover the base fee (will deactivate if tunnel didn’t have band).

```protobuf
// Activate is the transaction message to activate a tunnel.
message MsgActivate {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgActivate";

  // tunnel_id is the ID of the tunnel to activate.
  uint64 tunnel_id = 1 [(gogoproto.customname) = "TunnelID"];
  // creator is the address of the creator.
  string creator = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgDeactivate

To stop producing new packets, the tunnel can be deactivated.

```protobuf
// MsgDeactivate is the transaction message to deactivate a tunnel.
message MsgDeactivate {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgDeactivate";

  // tunnel_id is the ID of the tunnel to deactivate.
  uint64 tunnel_id = 1 [(gogoproto.customname) = "TunnelID"];
  // creator is the address of the creator.
  string creator = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgTriggerTunnel

Allows the manual creation of a packet without waiting for the deviation or interval conditions to be met.

```protobuf
// MsgTriggerTunnel is the transaction message to manually trigger a tunnel.
message MsgTriggerTunnel {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgTriggerTunnel";

  // tunnel_id is the ID of the tunnel to manually trigger.
  uint64 tunnel_id = 1 [(gogoproto.customname) = "TunnelID"];
  // creator is the address of the creator.
  string creator = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgDepositToTunnel

Increase the `total_deposit` for the tunnel by depositing more coins.

```protobuf
// MsgDepositToTunnel defines a message to submit a deposit to an existing tunnel.
message MsgDepositToTunnel {
  option (cosmos.msg.v1.signer) = "depositor";
  option (amino.name)           = "tunnel/MsgDepositToTunnel";

  // tunnel_id defines the unique id of the tunnel.
  uint64 tunnel_id = 1
      [(gogoproto.customname) = "TunnelID", (gogoproto.jsontag) = "tunnel_id", (amino.dont_omitempty) = true];

  // amount to be deposited by depositor.
  repeated cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (amino.dont_omitempty)   = true
  ];

  // depositor defines the deposit addresses from the tunnel.
  string depositor = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgWithdrawFromTunnel

Allows users to withdraw their deposited coins from the tunnel.

```protobuf
// MsgWithdrawFromTunnel is the transaction message to withdraw a deposit from an existing tunnel.
message MsgWithdrawFromTunnel {
  option (cosmos.msg.v1.signer) = "withdrawer";
  option (amino.name)           = "tunnel/MsgWithdrawFromTunnel";

  // tunnel_id defines the unique id of the tunnel.
  uint64 tunnel_id = 1
      [(gogoproto.customname) = "TunnelID", (gogoproto.jsontag) = "tunnel_id", (amino.dont_omitempty) = true];

  // amount to be withdrawn by withdrawer.
  repeated cosmos.base.v1beta1.Coin amount = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins",
    (amino.dont_omitempty)   = true
  ];

  // withdrawer defines the withdraw addresses from the tunnel.
  string withdrawer = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

## Events

The `x/tunnel` module emits several events that can be used to track the state changes and actions within the module. These events are helpful for developers and users to monitor tunnel creation, updates, activations, deactivations, and packet production.

### Event: `create_tunnel`

This event is emitted when a new tunnel is created.

| Attribute Key        | Attribute Value                       |
| -------------------- | ------------------------------------- |
| tunnel_id            | `{ID}`                                |
| interval             | `{Interval}`                          |
| route                | `{Route.String()}`                    |
| fee_payer            | `{FeePayer}`                          |
| is_active            | `{IsActive}`                          |
| created_at           | `{CreatedAt}`                         |
| creator              | `{Creator}`                           |
| signal_id[]          | `{SignalDeviation.SignalID}}`         |
| soft_deviation_bps[] | `{SignalDeviation.SoftDeviationBPS}}` |
| hard_deviation_bps[] | `{SignalDeviation.hardDeviationBPS}}` |

### Event: `update_signals_and_interval`

This event is emitted when an existing tunnel is edited.

| Attribute Key        | Attribute Value                       |
| -------------------- | ------------------------------------- |
| tunnel_id            | `{ID}`                                |
| interval             | `{Interval}`                          |
| signal_id[]          | `{SignalDeviation.SignalID}}`         |
| soft_deviation_bps[] | `{SignalDeviation.SoftDeviationBPS}}` |
| hard_deviation_bps[] | `{SignalDeviation.hardDeviationBPS}}` |

### Event: `activate_tunnel`

This event is emitted when a tunnel is activated.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| tunnel_id     | `{ID}`          |
| is_active     | `true`          |

### Event: `deactivate_tunnel`

This event is emitted when a tunnel is deactivated.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| tunnel_id     | `{ID}`          |
| is_active     | `false`         |

### Event: `trigger_tunnel`

This event is emitted when a tunnel is triggered to produce a packet due to deviations or intervals.

| Attribute Key | Attribute Value     |
| ------------- | ------------------- |
| tunnel_id     | `{ID}`              |
| sequence      | `{packet_sequence}` |

### Event: `produce_packet_fail`

This event is emitted when the tunnel fails to produce a packet.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| tunnel_id     | `{ID}`          |
| reason        | `{err.Error()}` |

### Event: `produce_packet_success`

This event is emitted when the tunnel succeeds to produce a packet.

| Attribute Key | Attribute Value     |
| ------------- | ------------------- |
| tunnel_id     | `{ID}`              |
| sequence      | `{packet.Sequence}` |

### Event: `deposit_to_tunnel`

This event is emitted when a deposit is made to the tunnel.

| Attribute Key | Attribute Value            |
| ------------- | -------------------------- |
| tunnel_id     | `{tunnelID}`               |
| depositor     | `{depositor.String()}`     |
| amount        | `{depositAmount.String()}` |

### Event: `withdraw_from_tunnel`

This event is emitted when a withdrawal deposit is made to the tunnel.

| Attribute Key | Attribute Value            |
| ------------- | -------------------------- |
| tunnel_id     | `{tunnelID}`               |
| depositor     | `{depositor.String()}`     |
| amount        | `{depositAmount.String()}` |

## Clients

Users can interact with the `x/tunnel` module via the Command-Line Interface (CLI). The CLI allows for querying tunnel states and performing various operations.

### CLI Commands

To access the tunnel module commands, use:

```bash
bandd query tunnel --help
```

#### Query Commands

The query commands enable users to retrieve information about tunnels, deposits, and packets.

##### List All Tunnels

To query all tunnels in the `x/tunnel` module:

```bash
bandd query tunnel tunnels
```

##### Get Tunnel by ID

To query a specific tunnel by its ID:

```bash
bandd query tunnel tunnel [tunnel-id]
```

##### Get Deposits for a Tunnel

To query the total deposits for a tunnel by its ID:

```bash
bandd query tunnel deposits [tunnel-id]
```

##### Get Deposit by Depositor

To query the total deposit of a depositor for a specific tunnel:

```bash
bandd query tunnel deposit [tunnel-id] [depositor-address]
```

##### List All Packets for a Tunnel

To query all packets produced by a tunnel:

```bash
bandd query tunnel packets [tunnel-id]
```

##### Get Packet by Sequence

To query a specific packet produced by a tunnel using its sequence number:

```bash
bandd query tunnel packet [tunnel-id] [sequence]
```

##### Get Total Fees

To query the total fees collected by the tunnel module:

```bash
bandd query tunnel total-fees
```
