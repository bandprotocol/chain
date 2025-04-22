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
    - [Vote](#vote)
    - [Feed](#feed)
      - [Feed Interval](#feed-interval)
      - [Feed Deviation](#feed-deviation)
      - [How Feed Interval and Deviation are calculated](#how-feed-interval-and-deviation-are-calculated)
      - [Current Feeds](#current-feeds)
    - [Validator Price](#validator-price)
      - [Status](#status)
    - [Price](#price)
      - [Status](#status-1)
    - [Reference Source Config](#reference-source-config)
  - [State](#state)
    - [ReferenceSourceConfig](#referencesourceconfig)
    - [CurrentFeeds](#currentfeeds)
    - [ValidatorPriceList](#validatorpricelist)
    - [Price](#price-1)
    - [Vote](#vote-1)
    - [SignalTotalPower](#signaltotalpower)
      - [SignalTotalPowerByPowerIndex](#signaltotalpowerbypowerindex)
    - [Params](#params)
  - [Messages](#messages)
    - [MsgVote](#msgvote)
    - [MsgSubmitSignalPrices](#msgsubmitsignalprices)
    - [MsgUpdateReferenceSourceConfig](#msgupdatereferencesourceconfig)
    - [MsgUpdateParams](#msgupdateparams)
  - [End-Block](#end-block)
    - [Update Prices](#update-prices)
      - [Input](#input)
      - [Objective](#objective)
      - [Assumption](#assumption)
      - [Constraint](#constraint)
      - [Procedure](#procedure)
    - [Update current feeds](#update-current-feeds)
  - [Events](#events)
    - [EndBlocker](#endblocker)
    - [Handlers](#handlers)
      - [MsgSubmitSignalPrices](#msgsubmitsignalprices-1)
      - [MsgUpdateReferenceSourceConfig](#msgupdatereferencesourceconfig-1)
      - [MsgUpdateParams](#msgupdateparams-1)
      - [MsgVote](#msgvote-1)

## Concepts

### Vote

A Vote is a decision made by a voter, directing the network to deliver feed service for specified signal IDs.

A vote can contain multiple signals for each distinct signal ID.

A signal consists of a signal ID and the power associated with that signal. The feeding interval and deviation are reduced by the sum of the power of the signal. The total power of all signals of voter cannot exceed their total bonded delegation and staked tokens.

### Feed

A Feed is a data structure containing a signal ID and calculated interval and deviation values from the total power. Essentially, it instructs the validator regarding which signal IDs' prices need to be submitted at each specified interval or deviation.

#### Feed Interval

The interval is calculated based on the total power of the signal ID; the greater the power, the shorter the interval. The total power of a signal is the sum of the power of its signal IDs received from the voters. The minimum and maximum intervals are determined by parameters called `MinInterval` and `MaxInterval` , respectively.

#### Feed Deviation

Deviation follows a similar logic to interval. On-chain deviation is measured in basis point, meaning a deviation of 1 indicates a price tolerance within 0.01%. The minimum and maximum deviations are determined by parameters called `MinDeviationBasisPoint` and `MaxDeviationBasisPoint` , respectively.

It should be noted that while feed deviation is calculated, it is only used as a reference value for the price service. This is because the chain cannot penalize validators for not reporting on price deviations, unlike time intervals.

#### How Feed Interval and Deviation are calculated

* Power is registered after surpassing the `PowerStepThreshold`.
* Then, the power factor is calculated as the floor(Power / `PowerStepThreshold`).
* Subsequently, the interval is calculated as the maximum of `MinInterval` or the floor(`MaxInterval` / power factor).
* The deviation is then calculated as the max(`MinDeviationBasisPoint`, (`MaxDeviationBasisPoint` / power factor).

You can visualize the interval/deviation as resembling the harmonic series times MaxInterval/MaxDeviationBasisPoint, with the step of PowerStepThreshold.

#### Current Feeds

The list of currently supported feeds includes those with power exceeding the PowerStepThreshold parameter and ranking within the top `MaxCurrentFeeds` . The current feeds will be re-calculated on every `CurrentFeedsUpdateInterval` block(s). Validators are only required to submit their prices for the current feeds.

### Validator Price

The Validator Price refers to the price submitted by each validator before being aggregated into the final Price.

The module only contains the latest price of each validator and signal ID.

#### Status

Validator Price can report three valid price status:

1. `SIGNAL_PRICE_STATUS_UNSUPPORTED`: Indicates that the requested signal ID is not supported by this validator and will not be available in the foreseeable future unless an upgrade occurs.

2. `SIGNAL_PRICE_STATUS_UNAVAILABLE`: Indicates that the requested signal ID is currently unavailable but is expected to become available in the near future.

3. `SIGNAL_PRICE_STATUS_AVAILABLE`: Indicates that the price for the requested signal ID is available.

### Price

A Price is a structure that maintains the current price state for a signal ID, including its current price, price status, and the most recent timestamp.

Once the Validator Price is submitted, it will be weighted median which is weighted by how latest the price is and how much power the owner of the price has to get the most accurate and trustworthy price.

The module only contains the latest price of each signal ID of Current feeds.

#### Status

The price status includes the following valid states:

1. `PRICE_STATUS_UNKNOWN_SIGNAL_ID`: Indicates that the price for this signal ID is not supported by the majority of price feeder and will not be available in the foreseeable future unless the feeders undergo an upgrade of their price service registry.

2. `PRICE_STATUS_NOT_READY`: Indicates that the price for this signal ID is currently not ready but is expected to become available in the near future.

3. `PRICE_STATUS_AVAILABLE`: Indicates that the price for this signal ID is available.

4. `PRICE_STATUS_NOT_IN_CURRENT_FEEDS`: Indicates that this signal ID is not included in the currently supported feeds but can be added through a voting process.

### Reference Source Config

The On-chain Reference Source Config is the agreed-upon version of the reference source suggested for validators to use when querying prices for the feeds. Only the admin address can update this configuration.

## State

### ReferenceSourceConfig

ReferenceSourceConfig is a single-value store that hold Reference Source information.

* ReferenceSourceConfig: `0x00 -> ProtocolBuffer(ReferenceSourceConfig)`

### CurrentFeeds

CurrentFeeds is a single-value store that hold currently supported feeds.

* CurrentFeeds: `0x01 -> ProtocolBuffer(CurrentFeeds)`

### ValidatorPriceList

The ValidatorPrice is a space for holding the current lists of validator prices.

* ValidatorPrice: `0x10 -> ProtocolBuffer(ValidatorPriceList)`

### Price

The Price is a space for holding the current price information of signals.

* Price: `0x11 -> ProtocolBuffer(Price)`

### Vote

The Vote is a space for holding current vote information of voters.

* Vote: `0x12 -> ProtocolBuffer(Vote)`

### SignalTotalPower

The SignalTotalPower is a space for holding the total power of signals.

* SignalTotalPower: `0x13 -> ProtocolBuffer(Signal)`

#### SignalTotalPowerByPowerIndex

`SignalTotalPowerByPowerIndex` allows to retrieve SignalTotalPower by power:
 `0x80| BigEndian(Power) | SignalIDLen (1 byte) | SignalID -> SignalID`

### Params

The feeds module stores its params in state with the prefix of `0x10` , 
it can be updated with governance proposal or the address with authority.

* Params: `0x90 | ProtocolBuffer(Params)`

```protobuf
// Params is the data structure that keeps the parameters of the feeds module.
message Params {
  option (gogoproto.equal) = true; // Use gogoproto.equal for proto3 message equality checks

  // admin is the address of the admin that is allowed to perform operations on modules.
  string admin = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // allowable_block_time_discrepancy is the allowed discrepancy (in seconds) between validator price timestamp and
  // block_time.
  int64 allowable_block_time_discrepancy = 2;

  // grace_period is the time (in seconds) given for validators to adapt to changing in feed's interval.
  int64 grace_period = 3;

  // min_interval is the minimum limit of every feeds' interval (in seconds).
  // If the calculated interval is lower than this, it will be capped at this value.
  int64 min_interval = 4;

  // max_interval is the maximum limit of every feeds' interval (in seconds).
  // If the calculated interval of a feed is higher than this, it will not be capped at this value.
  int64 max_interval = 5;

  // power_step_threshold is the amount of minimum power required to put feed in the current feeds list.
  int64 power_step_threshold = 6;

  // max_current_feeds is the maximum number of feeds supported at a time.
  uint64 max_current_feeds = 7;

  // cooldown_time represents the duration (in seconds) during which validators are prohibited from sending new prices.
  int64 cooldown_time = 8;

  // min_deviation_basis_point is the minimum limit of every feeds' deviation (in basis point).
  int64 min_deviation_basis_point = 9;

  // max_deviation_basis_point is the maximum limit of every feeds' deviation (in basis point).
  int64 max_deviation_basis_point = 10;

  // current_feeds_update_interval is the number of blocks after which the current feeds will be re-calculated.
  int64 current_feeds_update_interval = 11;

  // price_quorum is the minimum percentage of power that needs to be reached for a price to be processed.
  string price_quorum = 12;

  // MaxSignalIDsPerSigning is the maximum number of signals allowed in a single tss signing request.
  uint64 max_signal_ids_per_signing = 13 [(gogoproto.customname) = "MaxSignalIDsPerSigning"];
}
```

## Messages

In this section, we describe the processing of the `feeds` messages and the corresponding updates to the state. All created/modified state objects specified by each message are defined within the [state](#state) section.

### MsgVote

Vote contain a batch of signal and power.

Batched Signals replace the previous Signals of the same voter as a batch.
Signals are registered, and their power is added to the SignalTotalPower of the same SignalID.

```protobuf
// MsgVote is the transaction message to submit signals.
message MsgVote {
  option (cosmos.msg.v1.signer) = "voter";
  option (amino.name)           = "feeds/MsgVote";

  // voter is the address of the voter that wants to vote signals.
  string voter = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // signals is a list of submitted signals.
  repeated Signal signals = 2 [(gogoproto.nullable) = false];
}
```

The message handling can fail if:

* The voter's address is not correct.
* The voter has less power than the sum of the Powers.
* The signal is not valid. (e.g. too long signal ID, power is a negative value).
* The size of the list of signal is too large.

### MsgSubmitSignalPrices

Validator Prices are submitted using the `MsgSubmitSignalPrices` message.
The price of signals will be updated at the end block using these new prices from validators.

```protobuf
// MsgSubmitSignalPrices is the transaction message to submit multiple signal prices.
message MsgSubmitSignalPrices {
  option (cosmos.msg.v1.signer) = "validator";
  option (amino.name)           = "feeds/MsgSubmitSignalPrices";

  // validator is the address of the validator that is performing the operation.
  string validator = 1 [(cosmos_proto.scalar) = "cosmos.ValidatorAddressString"];

  // timestamp is the timestamp used as reference for the data.
  int64 timestamp = 2;

  // signal_prices is a list of signal prices to submit.
  repeated SignalPrice signal_prices = 3 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* validator address is not correct.
* validator status is not bonded.
* validator's oracle status is not active.
* timestamp is too different from block time.
* the price is submitted in the `CooldownTime` param.
* the signals of the prices are not in the current feeds.
  

### MsgUpdateReferenceSourceConfig

Reference Source can be updated with the `MsgUpdateReferenceSourceConfig` message.
Only the assigned admin can update the Reference Source.

```protobuf
// MsgUpdateReferenceSourceConfig is the transaction message to update reference price source's configuration.
message MsgUpdateReferenceSourceConfig {
  option (cosmos.msg.v1.signer) = "admin";
  option (amino.name)           = "feeds/MsgUpdateReferenceSourceConfig";

  // admin is the address of the admin that is performing the operation.
  string admin = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // reference_source_config is the information of reference price source.
  ReferenceSourceConfig reference_source_config = 2 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* sender address does not match the `Admin` param.
* Reference Source's URL is not in the correct format of a URL.

### MsgUpdateParams

The `MsgUpdateParams` update the feeds module parameters.
The params are updated through a governance proposal where the signer is the gov module account address or other specified authority addresses.

```protobuf
// MsgUpdateParams is the transaction message to update parameters.
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "feeds/MsgUpdateParams";

  // authority is the address of the governance account.
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // params is the x/feeds parameters to update.
  Params params = 2 [(gogoproto.nullable) = false];
}
```

The message handling can fail if:

* signer is not the authority defined in the feeds keeper (usually the gov module account).

## End-Block

Each abci end block call, the operations to update prices.

### Update Prices

At every end block, the Validator Price of all current feeds will be obtained and checked if it is within the acceptance period (1 interval).
Any validator that does not submit a price within this period is considered to have miss-reported and will be deactivated unless the current feeds are in a grace period.
Accepted Validator Prices of the same SignalID will be weighted and median based on the recency of the price and the power of the validator who submitted the price.
The median price is then set as the Price. Here is the price aggregation logic:

#### Input

A list of ValidatorPriceInfo objects, each containing:
* `Price`: The reported price from the feeder
* `SignalPriceStatus`: The status of price
* `Power`: The feeder's power
* `Timestamp`: The time at which the price is reported

#### Objective

* An aggregated price from the list of ValidatorPriceInfo.

#### Assumption

1. No ValidatorPriceInfo has a power that exceeds 25% of the total power in the list.

#### Constraint

1. If more than half of the total power of the validator price have unsupported price status, it returns a `PRICE_STATUS_UNKNOWN_SIGNAL_ID` price status with price 0.
2. If the total power of all validator prices reported is than price quorum percentage, it returns an `PRICE_STATUS_NOT_READY` price status with price 0.
3. If less than half of total power of prices reported have available price status, it also returns an `PRICE_STATUS_NOT_READY` price status with price 0.

#### Procedure

1. Filter and order the List:

* Filter the object with `SignalPriceStatus` as `Available` only.
* Sort the list by `Timestamp` in descending order (latest timestamp first).
* For entries with the same `Timestamp`, sort by `Power` in descending order.

2. Apply Power Weights:

* Calculate the total power from the list.
* Assign weights to the powers in segments as follows:
    - The first 1/32 of the total power is multiplied by 6.
    - The next 1/16 of the total power is multiplied by 4.
    - The next 1/8 of the total power is multiplied by 2.
    - The next 1/4 of the total power is multiplied by 1.1.
* If ValidatorPriceInfo overlaps between segments, split it into parts corresponding to each segment and assign the respective multiplier.
* Any power that falls outside these segments will have a multiplier of 1.

3. Generate Points:

* For each ValidatorPriceInfo, generate a point (at the `Price` with the assigned `Weight`.)

4. Calculating Weight Median

* Compute the weighted median of the generated points to determine the final aggregated price.
* The weighted median price is the price at which the cumulative power (sorted by increasing price) crosses half of the total weighted power.

### Update current feeds

At every `BlocksPerFeedsUpdate` block(s), the current feeds will be re-calculated based on the parameters of the module (e.g. `MinInterval` , `MaxCurrentFeeds` ). 

## Events

The feeds module emits the following events:

### EndBlocker

| Type                  | Attribute Key         | Attribute Value |
| --------------------- | --------------------- | --------------- |
| update_price          | signal_id             | {signalID}      |
| update_price          | price_status          | {priceStatus}   |
| update_price          | price                 | {price}         |
| update_price          | timestamp             | {timestamp}     |
| updated_current_feeds | last_update_timestamp | {timestamp}     |
| updated_current_feeds | last_update_block     | {block_height}  |

### Handlers

#### MsgSubmitSignalPrices

| Type                | Attribute Key       | Attribute Value     |
| ------------------- | ------------------- | ------------------- |
| submit_signal_price | signal_price_status | {signalPriceStatus} |
| submit_signal_price | validator           | {validatorAddress}  |
| submit_signal_price | signal_id           | {signalID}          |
| submit_signal_price | price               | {price}             |
| submit_signal_price | timestamp           | {timestamp}         |

#### MsgUpdateReferenceSourceConfig

| Type                           | Attribute Key      | Attribute Value    |
| ------------------------------ | ------------------ | ------------------ |
| update_reference_source_config | registry_ipfs_hash | {registryIPFSHash} |
| update_reference_source_config | version            | {registryVersion}  |

#### MsgUpdateParams

| Type          | Attribute Key | Attribute Value |
| ------------- | ------------- | --------------- |
| update_params | params        | {params}        |

#### MsgVote

| Type                      | Attribute Key | Attribute Value |
| ------------------------- | ------------- | --------------- |
| update_signal_total_power | signal_id     | {signalID}      |
| update_signal_total_power | power         | {power}         |
| delete_signal_total_power | signal_id     | {signalID}      |
