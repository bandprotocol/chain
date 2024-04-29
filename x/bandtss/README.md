# `x/bandtss`

## Abstract

Bandtss module facilitates secure message signing within a decentralized network.

Users can request the module to sign a message, which is then authenticated by members within the module. This ensures the integrity and authenticity of the message, enhancing trust and reliability within the network.

A fee is charged per request, as specified by the module configuration, and upon successful signing, the fee is transferred to the assigned members.

This module is used in the BandChain.

## Contents

- [`x/bandtss`](#xbandtss)
  - [Abstract](#abstract)
  - [Contents](#contents)
  - [Concepts](#concepts)
    - [Current Group](#current-group)
    - [Member](#member)
    - [Reward](#reward)
      - [Block rewards](#block-rewards)
      - [Request fee](#request-fee)
    - [Signing](#signing)
    - [Replacement](#replacement)
  - [State](#state)
  - [Msg Service](#msg-service)
    - [Msg/CreateGroup](#msgcreategroup)
    - [Msg/ReplaceGroup](#msgreplacegroup)
    - [Msg/RequestSignature](#msgrequestsignature)
    - [Msg/Activate](#msgactivate)
    - [Msg/HealthCheck](#msghealthcheck)
    - [Msg/UpdateParams](#msgupdateparams)
  - [Events](#events)
    - [EventTypeFirstGroupCreated](#eventtypefirstgroupcreated)
    - [EventTypeRequestSignature](#eventtyperequestsignature)
    - [EventTypeReplacement](#eventtypereplacement)
    - [EventTypeActivate](#eventtypeactivate)
    - [EventTypeHealthCheck](#eventtypehealthcheck)
    - [EventTypeInactiveStatus](#eventtypeinactivestatus)
  - [Parameters](#parameters)
  - [Client](#client)
    - [CLI](#cli)
      - [Query](#query)
    - [Group](#group-1)
    - [Signing](#signing-1)
    - [gRPC](#grpc)
      - [Group](#group-2)
      - [Signing](#signing-2)
    - [REST](#rest)
      - [Group](#group-3)
      - [Signing](#signing-3)

## Concepts

### Current Group

The signing process mainly happens in the tss module. To handle the signing in bandtss module, a proposal is made to create a tss group for the module and use it as a main group for a signing process. The proposal and bandtss's groups can be created multiple times, in that case, the first one is considered the main group of the module and being used for signing process until a proposal for group replacement is made and approved.

### Member

The members of the module, which are the members of the current group, are nominated in a proposal. Once assigned, these members are responsible for signing messages within the system. Failure to sign messages within the specified timeframe results in deactivation and forfeiture of some block rewards from the module.

Members who have been deactivated can reactivate themselves by calling [Msg/Activate](#msgactivate) once the penalty duration has been met.

Additionally, members are required to continuously notify the chain for their active status. This is achieved through the [Msg/HealthCheck](#msghealthcheck).

Changing the members of the module can be done by [group replacement process](#replacement)

### Reward

#### Block rewards

In each block, active validators being served as members on the bandtss system are rewarded with additional block rewards proportional to their validating power, as recognition for their service.

The `RewardPercentage` parameter determines the percentage of block rewards allocated to these validators. By default, this parameter is set to 50%. However, please note that this percentage is calculated based on the remaining rewards. For instance, if other modules claim 40% of the rewards, the bandtss module will receive only 30% (50% of the remaining 60%) of the total block rewards.

#### Request fee

Users requesting signatures from the bandtss system are required to pay a fee for the service. This fee price is configured in the module's params. Only assigned members of the request will receive this fee as a reward for their service to the group, in addition to block rewards.

### Signing

A signing request can be submitted to the module, and the signing process is then forwarded to the TSS module. Subsequently, the bandtss module charges a fee to the requester, which will be transferred later to the assigned members once the message is successfully signed (as a request fee).

### Replacement

The replacement process is utilized when updating the members of the module or modifying the module's shared key. The steps involved in the replacement process are as follows:

1. Initiate a proposal to create a new bandtss signing group.
2. Submit a replacement proposal with a newly created group and the replacement execution time.
3. After the proposal is approved, the current members are responsible for signing the replacement message.
4. At the replacement execution time, the new group becomes the current group, existing members are removed, and new members are activated.

This process provides users with ample time to update their key before the replacement takes effect. During this transition period, signing requests are sent to both the current group and the group scheduled for replacement. However, signing fees are allocated exclusively to the assigned members of the current group.

## State

The `x/bandtss` module keeps the state of the following primary objects:

1. CurrentGroupID stores main group ID of the module.
2. Members stores members information and their status.
3. Signings stores signing ID of the current group and replacing group, if any.
4. Replacement stores latest replacement information.

In addition, the `x/bandtss` module still keeps temporary information such as the mapping between signing ID from tss module to bandtss signing ID for using as a mapping when the hooks are called from the tss module.

Here are the prefixes for each object in the KVStore of the bandtss module.

```go
var (
	GlobalStoreKeyPrefix = []byte{0x00}
	ParamsKeyPrefix = []byte{0x01}
	MemberStoreKeyPrefix = []byte{0x02}
	SigningStoreKeyPrefix = []byte{0x03}

	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)
	CurrentGroupIDStoreKey = append(GlobalStoreKeyPrefix, []byte("CurrentGroupID")...)
	ReplacementStoreKey = append(GlobalStoreKeyPrefix, []byte("Replacement")...)

	SigningInfoStoreKeyPrefix = append(SigningStoreKeyPrefix, []byte{0x00}...)
	SigningIDMappingStoreKeyPrefix = append(SigningStoreKeyPrefix, []byte{0x01}...)
)
```

## Msg Service

### Msg/CreateGroup

A new group can be created with the `MsgCreateGroup` which needs to open through governance proposal.
This message contains the list of members, the threshold of the group.

It's expected to fail if:

- Members are not correct (e.g. wrong address format, duplicates).
- Threshold is more than the number of the members.

### Msg/ReplaceGroup

A replacement can be created with the `MsgReplaceGrouup` which needs to open through a governance proposal.
This message contains `new_group_id` , and `exec_time`.

It's expected to fail if:

- The status of groups is not active.
- Can't request signing `replacement message` from `current_group_id`

### Msg/RequestSignature

Anyone who wants to have a signature from the group can use `MsgRequestSignature` to send their message to the group to request a signature.

It contains `fee_limit`, and `Content`. `Content` is an interface that any module can implement to have its logic get the specific data from its module so that the module can produce a signature for that data.

### Msg/Activate

If members are deactivated due to one of the module's mechanisms, such as a health check or missing signature, they must send `MsgActivate` to rejoin the system. However, there is a punishment period for rejoining the process.

### Msg/HealthCheck

This message is used by members in the bandtss system. All active members have to regularly send `MsgHealthCheck` to the chain to show if they are still active.

The frequency of sending is determined by `ActiveDuration` parameters.

### Msg/UpdateParams

When anyone wants to update the parameters of the bandtss module, they will have to open a governance proposal by using the `MsgUpdateParams` of the bandtss module to update those parameters.

## Events

The bandtss module emits the following events:

### EventTypeFirstGroupCreated

This event ( `first_group_created` ) is emitted when the first bandtss group is created and is set as a current group.

| Attribute Key    | Attribute Value |
| ---------------- | --------------- |
| current_group_id | {groupID}       |

### EventTypeRequestSignature

This event ( `bandtss_signing_request_created` ) is emitted when the module is requested to sign the data.

| Attribute Key              | Attribute Value    |
| -------------------------- | ------------------ |
| bandtss_signing_id         | {bandtssSigningID} |
| current_group_id           | {groupID}          |
| current_group_signing_id   | {signingID}        |
| replacing_group_id         | {groupID}          |
| replacing_group_signing_id | {signingID}        |

### EventTypeReplacement

This event ( `replacement` ) is emitted when replacement is requested via an approved proposal or the replacement changes its status.

| Attribute Key      | Attribute Value     |
| ------------------ | ------------------- |
| signingID          | {signingID}         |
| current_group_id   | {groupID}           |
| replacing_group_id | {groupID}           |
| status             | {replacementStatus} |

### EventTypeActivate

This event ( `activate` ) is emitted when an account submitted `MsgActivate` to the chain

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| address       | {memberAddress} |

### EventTypeHealthCheck

This event ( `healthcheck` ) is emitted when an account submitted `MsgHealthCheck` to the chain

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| address       | {memberAddress} |

### EventTypeInactiveStatus

This event ( `inactive_status` ) is emitted when an account is deactivated

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| address       | {memberAddress} |

## Parameters

The module contains the following parameters

```protobuf
type Params struct {
	// active_duration is the duration where a member can be active without interaction.
	ActiveDuration time.Duration
	// reward_percentage is the percentage of block rewards allocated to active TSS validators after being allocated to
	// oracle rewards.
	RewardPercentage uint64
	// inactive_penalty_duration is the duration where a member cannot activate back after inactive.
	InactivePenaltyDuration time.Duration
	// fee is the tokens that will be paid per signing.
	Fee github_com_cosmos_cosmos_sdk_types.Coins
}
```

## Client

### CLI

A user can query and interact with the `bandtss` module using the CLI.

#### Query

The `query` commands allow users to know other possible queries of the `bandtss` module.

```bash
bandd query bandtss --help
```

##### CurrentGroup

The `current-group` command allows users to query for current group information.

```bash
bandd query bandtss current-group
```

##### Member

The `Member` command allows users to query for member information by giving a member address.

```bash
bandd query bandtss member [address] [flags]
```

Example:

```bash
bandd query bandtss member band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q
```

##### Replacement

The `Replacement` command allows users to query for replacement information.

```bash
bandd query bandtss replacement
```

##### Signing

The `Signing` command allows users to query for bandtss signing information by giving a signing id.

```bash
bandd query bandtss signing [id] [flags]
```

Example:

```bash
bandd query bandtss signing 1
```

##### Params

The `Params` command allows users to query for module's configuration.

```bash
bandd query bandtss params
```

### gRPC

A user can query the `bandtss` module using gRPC endpoints.

##### CurrentGroup

The `current-group` command allows users to query for current group information.

```bash
bandtss.v1beta1.Query/CurrentGroup
```

##### Member

The `Member` command allows users to query for member information by giving a member address.

```bash
bandtss.v1beta1.Query/Member
```

Example:

```bash
grpcurl -plaintext
-d '{"address":"band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q"}' localhost:9090 bandtss.v1beta1.Query/Member
```

##### Replacement

The `Replacement` command allows users to query for replacement information.

```bash
bandtss.v1beta1.Query/Replacement
```

##### Signing

The `Signing` command allows users to query for bandtss signing information by giving a signing id.

```bash
bandtss.v1beta1.Query/Signing
```

Example:

```bash
grpcurl -plaintext
-d '{"signing_id":1}' localhost:9090 bandtss.v1beta1.Query/QuerySigningRequest
```

##### Params

The `Params` command allows users to query for module's configuration.

```bash
bandtss.v1beta1.Query/Params
```

### REST

A user can query the `bandtss` module using REST endpoints.

##### CurrentGroup

The `current-group` command allows users to query for current group information.

```bash
/tss/v1beta1/current_group
```

##### Member

The `Member` command allows users to query for member information by giving a member address.

```bash
/tss/v1beta1/current_group
```

Example:

```bash
curl localhost:1317/tss/v1beta1/groups/band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q
```

##### Replacement

The `Replacement` command allows users to query for replacement information.

```bash
/tss/v1beta1/replacement
```

##### Signing

The `Signing` command allows users to query for bandtss signing information by giving a signing id.

```bash
/tss/v1beta1/signing
```

Example:

```bash
curl localhost:1317/tss/v1beta1/signing/1
```

##### Params

The `Params` command allows users to query for module's configuration.

```bash
/tss/v1beta1/params
```
