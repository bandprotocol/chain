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
      - [Packet Generation](#packet-generation)
  - [State](#state)
    - [TunnelCount](#tunnelcount)
    - [TotalFee](#totalfee)
    - [ActiveTunnelID](#activetunnelid)
    - [Tunnel](#tunnel-1)
    - [Packet](#packet-1)
    - [LatestSignalPrices](#latestsignalprices)
    - [Deposit](#deposit)
    - [Params](#params)
  - [Msg](#msg)
    - [MsgCreateTunnel](#msgcreatetunnel)
    - [MsgUpdateAndResetTunnel](#msgupdateandresettunnel)
    - [MsgActivate](#msgactivate)
    - [MsgDeactivate](#msgdeactivate)
    - [MsgTriggerTunnel](#msgtriggertunnel)
    - [MsgDepositToTunnel](#MsgDepositToTunnel)
    - [MsgWithdrawFromTunnel](#MsgWithdrawFromTunnel)
  - [Events](#events)
    - [Event: `create_tunnel`](#event-create_tunnel)
    - [Event: `edit_tunnel`](#event-edit_tunnel)
    - [Event: `activate`](#event-activate)
    - [Event: `deactivate`](#event-deactivate)
    - [Event: `trigger_tunnel`](#event-trigger_tunnel)
    - [Event: `produce_packet_fail`](#event-produce_packet_fail)
  - [Clients](#clients)
    - [CLI Commands](#cli-commands)
      - [Query Commands](#query-commands)
        - [List All Tunnels](#list-all-tunnels)
        - [Get Tunnel by ID](#get-tunnel-by-id)
        - [Get Deposits for a Tunnel](#get-deposits-for-a-tunnel)
        - [Get Deposit by Depositor](#get-deposit-by-depositor)
        - [List All Packets for a Tunnel](#list-all-packets-for-a-tunnel)
        - [Get Packet by Sequence](#get-packet-by-sequence)

## Concepts

### Tunnel

The `x/tunnel` module defines a `Tunnel` type that specifies details such as the way to send the data to the destination ([Route](#route)), the type of price data to encode, the fee payer's address for packet fees, and the total deposit of the tunnel (which must meet a minimum requirement to activate). It also includes the interval and deviation settings for the price data used when producing packets at every end-block.

Users can create a new tunnel by sending a `MsgCreateTunnel` message to BandChain, specifying the desired signals and deviations. The available routes for tunnels are provided in the [Route](#route) concepts.

```go
type Tunnel struct {
    // ID is the unique identifier of the tunnel.
    ID uint64
    // Sequence represents the sequence number of the tunnel packets.
    Sequence uint64
    // Route defines the path for delivering the signal prices.
    Route *types1.Any
    // Encoder specifies the mode of encoding price signal data.
    Encoder Encoder
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

#### IBC Route

The IBC Route enables the transmission of data from BandChain to Cosmos-compatible chains via the Inter-Blockchain Communication (IBC) protocol. This route allows for secure and efficient cross-chain communication, leveraging the standardized IBC protocol to transmit packets of data between chains.

To create an IBC tunnel, use the following CLI command:

> **Note**: An example of the signalInfos-json-file can be found at scripts/tunnel/signal_deviations.json.

```bash
bandd tx tunnel create-tunnel ibc [channel-id] [encoder] [initial-deposit] [interval] [signalInfos-json-file]
```

#### TSS Route

### Packet

A Packet is the data unit produced and sent to the destination chain based on the specified route.

```go
type Packet struct {
    // tunnel_id is the tunnel ID
    TunnelID uint64
    // sequence is representing the sequence of the tunnel packet.
    Sequence uint64
    // signal_prices is the list of signal prices
    SignalPrices []SignalPrice
    // packet_content is the content of the packet that implements PacketContentI
    PacketContent *types1.Any
    // created_at is the timestamp when the packet is created
    CreatedAt int64
}
```

#### Packet Generation

At the end of every block, the tunnel generates packets by checking the deviations and intervals for each tunnel. We utilize both hard and soft deviations:

- **Hard Deviation**: If any signal reaches this threshold, the system triggers a check on all soft deviations.
- **Soft Deviation**: If any signal meets its soft deviation criteria during this check, the latest price is sent in the packet.

This mechanism helps reduce the number of transactions on the destination route during periods of market instability.

## State

### TunnelCount

Stores the number of tunnels existing on the chain.

- **TunnelCount**: `0x00 | -> BigEndian(#tunnels)`

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

### LatestSignalPrices

Stores the latest prices that the tunnel has sent to the destination route. These are used to compare intervals and deviations at the end of each block.

- **LatestSignalPrices**: `0x13 | TunnelID -> LatestSignalPrices`

### Deposit

Stores the total deposit per tunnel by each depositor.

- **Deposit**: `0x14 | TunnelID | DepositorAddress -> Deposit`

### Params

Stores the parameters in the state. These parameters can be updated via a governance proposal or by an authority address.

- **Params**: `0x90 -> Params`

The `x/tunnel` module contains the following parameters:

```go
type Params struct {
    // MinDeposit is the minimum deposit required to create a tunnel.
    MinDeposit sdk.Coins
    // MinInterval is the minimum interval in seconds.
    MinInterval uint64
    // MaxSignals defines the maximum number of signals allowed per tunnel.
    MaxSignals uint64
    // BasePacketFee is the base fee for each packet.
    BasePacketFee sdk.Coins
}
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
  google.protobuf.Any route = 3 [(cosmos_proto.accepts_interface) = "Route"];
  // encoder is the mode of encoding price signal data.
  Encoder encoder = 4;
  // initial_deposit is the deposit value that must be paid at tunnel creation.
  repeated cosmos.base.v1beta1.Coin initial_deposit = 5 [
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
- **Encoder Types**: Specifies the type of price value to be sent to the destination route.
  - Price
  - Tick
- **Initial Deposit**: The initial deposit can be set to zero. Other users can contribute to the tunnel's deposit until it reaches the required minimum deposit.

### MsgUpdateAndResetTunnel

**Editable Arguments**: The following parameters can be modified within the tunnel: `signal_deviations` and `Interval`

```protobuf
// MsgUpdateAndResetTunnel is the transaction message to update a tunnel information
// and reset the interval.
message MsgUpdateAndResetTunnel {
  option (cosmos.msg.v1.signer) = "creator";
  option (amino.name)           = "tunnel/MsgUpdateAndResetTunnel";

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
| encoder              | `{Encoder.String()}`                  |
| fee_payer            | `{FeePayer}`                          |
| is_active            | `{IsActive}`                          |
| created_at           | `{CreatedAt}`                         |
| creator              | `{Creator}`                           |
| signal_id[]          | `{SignalDeviation.SignalID}}`         |
| soft_deviation_bps[] | `{SignalDeviation.SoftDeviationBPS}}` |
| hard_deviation_bps[] | `{SignalDeviation.hardDeviationBPS}}` |

### Event: `update_and_reset_tunnel`

This event is emitted when an existing tunnel is edited.

| Attribute Key        | Attribute Value                       |
| -------------------- | ------------------------------------- |
| tunnel_id            | `{ID}`                                |
| interval             | `{Interval}`                          |
| signal_id[]          | `{SignalDeviation.SignalID}}`         |
| soft_deviation_bps[] | `{SignalDeviation.SoftDeviationBPS}}` |
| hard_deviation_bps[] | `{SignalDeviation.hardDeviationBPS}}` |

### Event: `activate`

This event is emitted when a tunnel is activated.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| tunnel_id     | `{ID}`          |
| is_active     | `true`          |

### Event: `deactivate`

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
