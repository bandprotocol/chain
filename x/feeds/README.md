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
    - [Feed](#feed-1)
      - [FeedByPowerIndex](#feedbypowerindex)
    - [ValidatorPrice](#validatorprice)
    - [Price](#price-1)
    - [DelegatorSignal](#delegatorsignal)
    - [Params](#params)
  - [Messages](#messages)
    - [MsgSubmitPrices](#msgsubmitprices)
    - [MsgUpdatePriceService](#msgupdatepriceservice)
    - [MsgUpdateParams](#msgupdateparams)
    - [MsgSubmitSignals](#msgsubmitsignals)
  - [End-Block](#end-block)
    - [Update Prices](#update-prices)
  - [Events](#events)
    - [EndBlocker](#endblocker)
    - [Handlers](#handlers)
      - [MsgSubmitPrices](#msgsubmitprices-1)
      - [MsgUpdatePriceService](#msgupdatepriceservice-1)
      - [MsgUpdateParams](#msgupdateparams-1)
      - [MsgSubmitSignals](#msgsubmitsignals-1)


## Concepts

### Delegator Signal

A Delegator Signal is a sign or vote from a delegator, instructing the chain to provide feed service for the designated ID.

A Delegator Signal consists of an ID and the power associated with that ID. The feeding interval and deviation are reduced by the sum of the power of the ID. The total power of a delegator cannot exceed their bonded delegation.

### Feed

A Feed is a data structure containing a signal ID, its total power, and calculated interval and deviation values. Essentially, it instructs the validator regarding which signal IDs' prices need to be submitted at each specified interval or deviation.

#### Feed Interval

The interval is calculated based on the power of its feed; the greater the power, the shorter the interval. The total power of a feed is the sum of the power of its signal IDs received from the delegators. The minimum and maximum intervals are determined by parameters called `MinInterval` and `MaxInterval`, respectively.

#### Feed Deviation

Deviation follows a similar logic to interval. On-chain deviation is measured in thousandths, meaning a deviation of 1 indicates a price tolerance within 0.1%. The minimum and maximum deviations are determined by parameters called `MinDeviationInThousandth` and `MaxDeviationInThousandth`, respectively.

#### How Feed Interval and Deviation are calculated

- Power is registered after surpassing the `PowerThreshold`.
- Then, the power factor is calculated as the floor(Power / PowerThreshold).
- Subsequently, the interval is calculated as the maximum of MinInterval or the floor(MaxInterval / power factor).
- The deviation is then calculated as the max(`MinDeviationInThousandth`, (`MaxDeviationInThousandth` / power factor).

You can visualize the interval/deviation as resembling the harmonic series times MaxInterval/MaxDeviationInThousandth, with step of PowerThreshold.

#### Supported Feeds

The list of currently supported feeds includes those with power exceeding the PowerThreshold parameter and ranking within the top MaxSupportedFeeds. Feeds outside of this list are considered unsupported, and validators do not need to submit their prices.

### Validator Price

The Validator Price refers to the price submitted by each validator before being aggregated into the final Price.

The module only contains the latest price of each validator and signal id.

### Price

A Price is a structure that maintains the current price state for a signal id, including its current price, price status, and the most recent timestamp.

Once the Validator Price is submitted, it will be weighted median which weight by how latest the price and how much validator power of the owner of the price to get the most accurate and trustworthy price.

The module only contains the latest price of each signal id.

### Price Service

The On-chain Price Service is the agreed-upon version of the price service suggested for validators to use when querying prices for the feeds.

## State

### PriceService

PriceService is stored in the global store `0x00` to hold Price Service information.

* PriceService: `0x00 | []byte("PriceService") -> ProtocolBuffer(PriceService)`

### Feed

The Feed is a space for holding current Feeds information.

* Feed: `0x01 -> ProtocolBuffer(Feed)`

#### FeedByPowerIndex

`FeedByPowerIndex` allow to retrieve Feeds by power:
`0x20| BigEndian(Power) | SignalIDLen (1 byte) | SignalID -> SignalID`

### ValidatorPrice

The ValidatorPrice is a space for holding current Validator Price information.

* ValidatorPrice: `0x02 -> ProtocolBuffer(ValidatorPrice)`

### Price

The Price is a space for holding current Priceinformation.

* Price: `0x03 -> ProtocolBuffer(Price)`

### DelegatorSignal

The DelegatorSignal is a space for holding current Delegator Signals information.

* DelegatorSignal: `0x04 -> ProtocolBuffer(Signal)`

### Params

The feeds module stores its params in state with the prefix of `0x10`,
it can be updated with governance or the address with authority.

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
}
```

## Messages

In this section we describe the processing of the feeds messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](#state) section.

### MsgSubmitPrices

Validator Prices are submitted using the `MsgSubmitPrices` message.
The Prices will be updated at endblock using this new Validator Prices.

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

* validator address is not correct
* validator status is not bonded
* price is submitted in `CooldownTime` param
* no Feed with the same signalID
  
### MsgUpdatePriceService

Price Service can be updated with the `MsgUpdatePriceService` message.
Only assigned admin can update the Price Service.

```protobuf
// MsgUpdatePriceService is the transaction message to update price service's information.
message MsgUpdatePriceService {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "feeds/MsgUpdateParams";

  // Admin is the address of the admin that is performing the operation.
  string admin = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // PriceService is the information of price service.
  PriceService price_service = 2 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* sender address do not match `Admin` param
* Price Service url is not in the correct format of an url

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
Signals are registered, and their power is added to the feeds of the same SignalID.
If the Feed's Interval is changed, its LastIntervalUpdateTimestamp will be marked as the block time.
If the updated Feed's Power is zero, it will be deleted from the state.
Every time there is an update to a Feed, `FeedByPowerIndex` will be re-indexed.

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

* delegator address is not correct
* delegator do not have less delegation than sum of the Powers
* no Feed with the same signalID

## End-Block

Each abci end block call, the operations to update prices.

### Update Prices

At every end block, the Validator Price of every Supported Feed will be obtained and checked if it is within the acceptance period (1 interval).
Any validator that does not submit a price within this period is considered to have miss-reported and will be deactivated, unless the Feed is in a transition period (where the interval has just been updated within TransitionTime).
Accepted Validator Prices of the same SignalID will be weighted and medianed based on the recency of the price and the power of the validator who submitted the price.
The medianed price is then set as the Price.

## Events

The feeds module emits the following events:

### EndBlocker

| Type                   | Attribute Key | Attribute Value |
| ---------------------- | ------------- | --------------- |
| calculate_price_failed | signal_id     | {signalID}      |
| calculate_price_failed | error_message | {error}         |
| update_price           | signal_id     | {signalID}      |
| update_price           | price         | {price}         |
| update_price           | timestamp     | {timestamp}     |

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

| Type         | Attribute Key           | Attribute Value         |
| ------------ | ----------------------- | ----------------------- |
| update_feed  | signal_id               | {signalID}              |
| update_feed  | power                   | {power}                 |
| update_feed  | interval                | {interval}              |
| update_feed  | timestamp               | {timestamp}             |
| update_feed  | deviation_in_thousandth | {deviationInThousandth} |
| deleate_feed | signal_id               | {signalID}              |
| deleate_feed | power                   | {power}                 |
| deleate_feed | interval                | {interval}              |
| deleate_feed | timestamp               | {timestamp}             |
| deleate_feed | deviation_in_thousandth | {deviationInThousandth} |
