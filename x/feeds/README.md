# `x/feeds`

## Abstract

This document specifies the Feeds module.

The Feeds module provides a mechanism for decentralized signal voting, price submission, and price updating within a blockchain. 

This module is used in the BandChain.

## Contents

- [`x/feeds`](#xfeeds)
  - [Abstract](#abstract)
  - [Contents](#contents)
  - [Concepts](#concepts)
    - [Delegator Signal](#delegator-signal)
    - [Feed](#feed)
      - [Feed Interval](#feed-interval)
      - [Feed Deviation](#feed-deviation)
      - [How Feed Interval and Deviation are calculated](#how-feed-interval-and-deviation-are-calculated)
      - [Supported Feeds](#supported-feeds)
    - [Validator Price](#validator-price)
    - [Price](#price)
    - [Price Service](#price-service)
  - [State](#state)
    - [PriceService](#priceservice)
    - [SupportedFeeds](#supportedfeeds)
    - [ValidatorPrice](#validatorprice)
    - [Price](#price-1)
    - [DelegatorSignal](#delegatorsignal)
    - [SignalTotalPower](#signaltotalpower)
      - [SignalTotalPowerByPowerIndex](#signaltotalpowerbypowerindex)
    - [Params](#params)
  - [Messages](#messages)
    - [MsgSubmitPrices](#msgsubmitprices)
    - [MsgUpdatePriceService](#msgupdatepriceservice)
    - [MsgUpdateParams](#msgupdateparams)
    - [MsgSubmitSignals](#msgsubmitsignals)
  - [End-Block](#end-block)
    - [Update Prices](#update-prices)
    - [Update Supported Feeds](#update-supported-feeds)
  - [Events](#events)
    - [EndBlocker](#endblocker)
    - [Handlers](#handlers)
      - [MsgSubmitPrices](#msgsubmitprices-1)
      - [MsgUpdatePriceService](#msgupdatepriceservice-1)
      - [MsgUpdateParams](#msgupdateparams-1)
      - [MsgSubmitSignals](#msgsubmitsignals-1)


## Concepts

### Delegator Signal

A Delegator Signal is a vote from a delegator, instructing the chain to provide feed service for the designated ID.

A Delegator Signal consists of an ID and the power associated with that ID. The feeding interval and deviation are reduced by the sum of the power of the ID. The total power of a delegator cannot exceed their total bonded delegation.

### Feed

A Feed is a data structure containing a signal ID and calculated interval and deviation values from the total power. Essentially, it instructs the validator regarding which signal IDs' prices need to be submitted at each specified interval or deviation.

#### Feed Interval

The interval is calculated based on the total power of the signal ID; the greater the power, the shorter the interval. The total power of a signal is the sum of the power of its signal IDs received from the delegators. The minimum and maximum intervals are determined by parameters called `MinInterval` and `MaxInterval`, respectively.

#### Feed Deviation

Deviation follows a similar logic to interval. On-chain deviation is measured in thousandths, meaning a deviation of 1 indicates a price tolerance within 0.1%. The minimum and maximum deviations are determined by parameters called `MinDeviationInThousandth` and `MaxDeviationInThousandth`, respectively.

#### How Feed Interval and Deviation are calculated

- Power is registered after surpassing the `PowerThreshold`.
- Then, the power factor is calculated as the floor(Power / PowerThreshold).
- Subsequently, the interval is calculated as the maximum of MinInterval or the floor(MaxInterval / power factor).
- The deviation is then calculated as the max(`MinDeviationInThousandth`, (`MaxDeviationInThousandth` / power factor).

You can visualize the interval/deviation as resembling the harmonic series times MaxInterval/MaxDeviationInThousandth, with the step of PowerThreshold.

#### Supported Feeds

The list of currently supported feeds includes those with power exceeding the PowerThreshold parameter and ranking within the top `MaxSupportedFeeds`. The supported feeds will be re-calculated on every `BlocksPerFeedsUpdate` block(s). Validators are only required to submit their prices for the supported feeds.

### Validator Price

The Validator Price refers to the price submitted by each validator before being aggregated into the final Price.

The module only contains the latest price of each validator and signal ID.

### Price

A Price is a structure that maintains the current price state for a signal ID, including its current price, price status, and the most recent timestamp.

Once the Validator Price is submitted, it will be weighted median which is weighted by how latest the price is and how much power the owner of the price has to get the most accurate and trustworthy price.

The module only contains the latest price of each signal ID.

### Price Service

The On-chain Price Service is the agreed-upon version of the price service suggested for validators to use when querying prices for the feeds. Only the admin address can update this configuration.

## State

### PriceService

PriceService is stored in the global store `0x00` to hold Price Service information.

* PriceService: `0x00 | []byte("PriceService") -> ProtocolBuffer(PriceService)`

### SupportedFeeds

SupportedFeeds is stored in the global store `0x00` to hold the list of supported feeds.

* SupportedFeeds: `0x00 | []byte("SupportedFeeds") -> ProtocolBuffer(SupportedFeeds)`

### ValidatorPrice

The ValidatorPrice is a space for holding the current price of signals of validators.

* ValidatorPrice: `0x01 -> ProtocolBuffer(ValidatorPrice)`

### Price

The Price is a space for holding the current price information of signals.

* Price: `0x02 -> ProtocolBuffer(Price)`

### DelegatorSignal

The DelegatorSignal is a space for holding current Delegator Signals information of validators.

* DelegatorSignal: `0x03 -> ProtocolBuffer(DelegatorSignals)`

### SignalTotalPower

The SignalTotalPower is a space for holding the total power of signals.

* SignalTotalPower: `0x04 -> ProtocolBuffer(Signal)`

#### SignalTotalPowerByPowerIndex

`SignalTotalPowerByPowerIndex` allow to retrieve SignalTotalPower by power:
`0x20| BigEndian(Power) | SignalIDLen (1 byte) | SignalID -> SignalID`

### Params

The feeds module stores its params in state with the prefix of `0x10`,
it can be updated with governance proposal or the address with authority.

* Params: `0x10 | ProtocolBuffer(Params)`

```protobuf
// Params is the data structure that keeps the parameters of the feeds module.
message Params {
  option (gogoproto.equal)            = true;  // Use gogoproto.equal for proto3 message equality checks
  option (gogoproto.goproto_stringer) = false; // Disable stringer generation for better control

  // Admin is the address of the admin that is allowed to perform operations on modules.
  string admin = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // AllowableBlockTimeDiscrepancy is the allowed discrepancy (in seconds) between validator price timestamp and
  // block_time.
  int64 allowable_block_time_discrepancy = 2;

  // TransitionTime is the time (in seconds) given for validators to adapt to changing in feed's interval.
  int64 transition_time = 3;

  // MinInterval is the minimum limit of every feeds' interval (in seconds).
  // If the calculated interval is lower than this, it will be capped at this value.
  int64 min_interval = 4;

  // MaxInterval is the maximum limit of every feeds' interval (in seconds).
  // If the calculated interval of a feed is higher than this, it will not be recognized as a supported feed.
  int64 max_interval = 5;

  // PowerThreshold is the amount of minimum power required to put feed in the supported list.
  int64 power_threshold = 6;

  // MaxSupportedFeeds is the maximum number of feeds supported at a time.
  int64 max_supported_feeds = 7;

  // CooldownTime represents the duration (in seconds) during which validators are prohibited from sending new prices.
  int64 cooldown_time = 8;

  // MinDeviationInThousandth is the minimum limit of every feeds' deviation (in thousandth).
  int64 min_deviation_in_thousandth = 9;

  // MaxDeviationInThousandth is the maximum limit of every feeds' deviation (in thousandth).
  int64 max_deviation_in_thousandth = 10;

  // MaxSignalIDCharacters is the maximum limit of characters of a signal id.
  uint64 max_signal_id_characters = 11 [(gogoproto.customname) = "MaxSignalIDCharacters"];

  // BlocksPerFeedsUpdate is the number of blocks after which the feed interval and deviation will be recalculated
  uint64 blocks_per_feeds_update = 12;
}
```

## Messages

In this section, we describe the processing of the `feeds` messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](#state) section.

### MsgSubmitPrices

Validator Prices are submitted using the `MsgSubmitPrices` message.
The price of signals will be updated at the end block using these new prices from validators.

```protobuf
// MsgSubmitPrices is the transaction message to submit multiple prices.
message MsgSubmitPrices {
  option (cosmos.msg.v1.signer) = "validator";
  option (amino.name)           = "feeds/MsgSubmitPrices";

  // Validator is the address of the validator that is performing the operation.
  string validator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // Timestamp is the timestamp use as reference of the data.
  int64 timestamp = 2;

  // Prices is a list of prices to submit.
  repeated SubmitPrice prices = 3 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* validator address is not correct.
* validator status is not bonded.
* validator's oracle status is not active.
* timestamp is too different from block time.
* the price is submitted in the `CooldownTime` param.
* the signals of the prices are not in the supported feeds.
  
### MsgUpdatePriceService

Price Service can be updated with the `MsgUpdatePriceService` message.
Only the assigned admin can update the Price Service.

```protobuf
// MsgUpdatePriceService is the transaction message to update price service's information.
message MsgUpdatePriceService {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "feeds/MsgUpdatePriceService";

  // Admin is the address of the admin that is performing the operation.
  string admin = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // PriceService is the information of price service.
  PriceService price_service = 2 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* sender address does not match the `Admin` param.
* Price Service's URL is not in the correct format of a URL.

### MsgUpdateParams

The `MsgUpdateParams` update the feeds module parameters.
The params are updated through a governance proposal where the signer is the gov module account address or other specified authority addresses.

```protobuf
// MsgUpdateParams is the transaction message to update parameters.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "feeds/MsgUpdateParams";

  // Authority is the address of the governance account.
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // Params is the x/feeds parameters to update.
  Params params = 2 [(gogoproto.nullable) = false];
}
```

The message handling can fail if:

* signer is not the authority defined in the feeds keeper (usually the gov module account).

### MsgSubmitSignals

Delegator Signals are submitted as a batch using the MsgSubmitSignals message.

Batched Signals replace the previous Signals of the same delegator as a batch.
Signals are registered, and their power is added to the SignalTotalPower of the same SignalID.

```protobuf
// MsgSubmitSignals is the transaction message to submit signals
message MsgSubmitSignals {
  option (cosmos.msg.v1.signer) = "delegator";
  option (amino.name)           = "feeds/MsgSubmitSignals";

  // Delegator is the address of the delegator that want to submit signals
  string delegator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // Signals is a list of submitted signal
  repeated Signal signals = 2 [(gogoproto.nullable) = false];
}
```

The message handling can fail if:

* The delegator's address is not correct.
* The delegator has less delegation than the sum of the Powers.
* The signal is not valid. (e.g. too long signal ID, power is a negative value).

## End-Block

Each abci end block call, the operations to update prices.

### Update Prices

At every end block, the Validator Price of all supported feeds will be obtained and checked if it is within the acceptance period (1 interval).
Any validator that does not submit a price within this period is considered to have miss-reported and will be deactivated unless the Supported feeds are in a transition period.
Accepted Validator Prices of the same SignalID will be weighted and median based on the recency of the price and the power of the validator who submitted the price.
The median price is then set as the Price. Here is the price aggregation logic:

#### Input

A list of PriceFeedInfo objects, each containing:
- `Price`: The reported price from the feeder
- `Deviation`: The price deviation
- `Power`: The feeder's power
- `Timestamp`: The time at which the price is reported

#### Objective

- An aggregated price from the list of priceFeedInfo.

#### Assumption

1. No PriceFeedInfo has a power that exceeds 25% of the total power in the list.

#### Procedure

1. Order the List:

- Sort the list by `Timestamp` in descending order (latest timestamp first).
- For entries with the same `Timestamp`, sort by `Power` in descending order.

2. Apply Power Weights:

- Calculate the total power from the list.
- Assign weights to the powers in segments as follows:
    - The first 1/32 of the total power is multiplied by 6.
    - The next 1/16 of the total power is multiplied by 4.
    - The next 1/8 of the total power is multiplied by 2.
    - The next 1/4 of the total power is multiplied by 1.1.
- If PriceFeedInfo overlaps between segments, split it into parts corresponding to each segment and assign the respective multiplier.
- Any power that falls outside these segments will have a multiplier of 1.

3. Generate Points:

- For each PriceFeedInfo (or its parts if split), generate three points:
    - One at the `Price` with the assigned `Power`.
    - One at `Price + Deviation` with the assigned `Power`.
    - One at `Price - Deviation` with the assigned `Power`.

4. Calculating Weight Median

- Compute the weighted median of the generated points to determine the final aggregated price.
- The weighted median price is the price at which the cumulative power (sorted by increasing price) crosses half of the total weighted power.

### Update supported feeds

At every `BlocksPerFeedsUpdate` block(s), the supported feeds will be re-calculated based on the parameters of the module (e.g. `MinInterval`, `MaxSupportedFeeds`). 

## Events

The feeds module emits the following events:

### EndBlocker

| Type                    | Attribute Key         | Attribute Value |
| ----------------------- | --------------------- | --------------- |
| calculate_price_failed  | signal_id             | {signalID}      |
| calculate_price_failed  | error_message         | {error}         |
| update_price            | signal_id             | {signalID}      |
| update_price            | price                 | {price}         |
| update_price            | timestamp             | {timestamp}     |
| updated_supported_feeds | last_update_timestamp | {timestamp}     |
| updated_supported_feeds | last_update_block     | {block_height}  |

### Handlers

#### MsgSubmitPrices

| Type         | Attribute Key | Attribute Value    |
| ------------ | ------------- | ------------------ |
| submit_price | price_status  | {priceStatus}      |
| submit_price | validator     | {validatorAddress} |
| submit_price | signal_id     | {signalID}         |
| submit_price | price         | {price}            |
| submit_price | timestamp     | {timestamp}        |


#### MsgUpdatePriceService

| Type                 | Attribute Key | Attribute Value |
| -------------------- | ------------- | --------------- |
| update_price_service | hash          | {hash}          |
| update_price_service | version       | {version}       |
| update_price_service | url           | {url}           |

#### MsgUpdateParams

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| update_params | params        | {params}        |

#### MsgSubmitSignals

| Type                      | Attribute Key | Attribute Value |
| ------------------------- | ------------- | --------------- |
| update_signal_total_power | signal_id     | {signalID}      |
| update_signal_total_power | power         | {power}         |
