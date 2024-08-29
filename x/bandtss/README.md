# `x/bandtss`

## Abstract

The Bandtss module serves as a critical component for ensuring secure message signing within a decentralized network, playing a pivotal role in maintaining the integrity and authenticity of communications across the system.

When a user requests the module to sign a message, it triggers a process where the message is authenticated by designated members within the module. This rigorous authentication process is designed to guarantee that the message has not been tampered with, thereby reinforcing trust and reliability within the network.

The module is configured to charge a fee for each signing request, a cost that is predefined in the module's settings. Upon the successful completion of the signing process, this fee is automatically transferred to the assigned members who participated in the authentication, rewarding them for their contribution.

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
    - [Transition](#transition)
  - [State](#state)
    - [Group & Member](#group--member)
    - [Group Transition](#group-transition)
    - [Signing](#signing-1)
    - [Params](#params)
  - [Msg Service](#msg-service)
    - [Msg/TransitionGroup](#msgtransitiongroup)
    - [Msg/ForceReplaceGroup](#msgforcereplacegroup)
    - [Msg/RequestSignature](#msgrequestsignature)
    - [Msg/Activate](#msgactivate)
    - [Msg/Heartbeat](#msgheartbeat)
    - [Msg/UpdateParams](#msgupdateparams)
  - [Events](#events)
    - [EventTypeSigningRequestCreated](#eventtypesigningrequestcreated)
    - [EventTypeGroupTransition](#eventtypegrouptransition)
    - [EventTypeGroupTransitionFailed](#eventtypegrouptransitionfailed)
    - [EventTypeGroupTransitionSuccess](#eventtypegrouptransitionsuccess)
    - [EventTypeActivate](#eventtypeactivate)
    - [EventTypeHeartbeat](#eventtypeheartbeat)
    - [EventTypeInactiveStatus](#eventtypeinactivestatus)
  - [Parameters](#parameters)
  - [Client](#client)
    - [CLI](#cli)
    - [gRPC](#grpc)
    - [REST](#rest)

## Concepts

### Current Group

The signing process mainly happens in the tss module. To handle the signing in bandtss module, a proposal is made to create a tss group for the module and use it as a main group for a signing process. The proposal and bandtss's groups can be created multiple times, in that case, the first one is considered the main group of the module and being used for signing process until a proposal for group transition is made and approved.

### Member

Members of the module are nominated through a proposal. Once assigned, they are responsible for signing messages within the system. If a member fails to sign messages within the specified timeframe, they face deactivation and forfeit a portion of their block rewards from the module.

Deactivated members can reactivate themselves after the penalty duration has elapsed. Additionally, members must continuously notify the chain to maintain their active status.

Changing the members of the module can be done by [group transition process](#transition)

```go
type Member struct {
	Address string
	GroupID github_com_bandprotocol_chain_v2_pkg_tss.GroupID
	IsActive bool
	Since time.Time
	LastActive time.Time
}
```

### Reward

#### Block rewards

In each block, active members in the current active group are rewarded with additional block rewards equally as recognition for their service.

The `RewardPercentage` parameter determines the percentage of block rewards allocated to these members in the current group. By default, this parameter is set to 50%. However, please note that this percentage is calculated based on the remaining rewards. For instance, if other modules claim 40% of the rewards, the bandtss module will receive only 30% (50% of the remaining 60%) of the total block rewards.

#### Request fee

Users requesting signatures from the bandtss system are required to pay a fee for the service. This fee price is configured in the module's params. Only assigned members of the request will receive this fee as a reward for their service to the group, in addition to block rewards.

### Signing

When a signing request is submitted to the module, the request is forwarded to the TSS module for processing. Following this, the bandtss module imposes a fee on the requester. This fee, referred to as the request fee, will be transferred to the assigned members after the message has been successfully signed.

If there is an incoming group during the transition process, the assigned members of this group are required to sign a given message without receiving any reward. Signer only are eligible for rewards if they are in the current active group.

```go
type Signing struct {
	ID SigningID
	FeePerSigner github_com_cosmos_cosmos_sdk_types.Coins
	Requester string
	CurrentGroupSigningID github_com_bandprotocol_chain_v2_pkg_tss.SigningID
	IncomingGroupSigningID github_com_bandprotocol_chain_v2_pkg_tss.SigningID
}
```

### Transition

The transition process is employed when updating the members of a module or modifying the module's shared key. This process ensures a smooth handover between the current and incoming groups, allowing users ample time to update their keys before the transition is finalized.

During the transition period, signing requests are sent to both the current group and the incoming group, although signing fees are allocated exclusively to the members of the current group.

The steps involved in the transition process are as follows:

1. Initiate Proposal: Start by proposing a transition from the current group members to those in the incoming group.
2. Group Creation in TSS Module: Once the proposal is approved, the TSS module triggers the group creation process, including the members listed in the proposal.
3. Sign Transition Message: After the group is successfully created, a signing request is sent to the current group to sign a transition message.
4. Group Transition: If the assigned members of the current group sign the message, the incoming group is prepared to replace the current group. If the execution time elapses, the incoming group automatically becomes the current group, existing members are removed, and new members are activated.

```go
type GroupTransition struct {
	SigningID github_com_bandprotocol_chain_v2_pkg_tss.SigningID
	CurrentGroupID github_com_bandprotocol_chain_v2_pkg_tss.GroupID
	CurrentGroupPubKey github_com_bandprotocol_chain_v2_pkg_tss.Point
	IncomingGroupID github_com_bandprotocol_chain_v2_pkg_tss.GroupID
	IncomingGroupPubKey github_com_bandprotocol_chain_v2_pkg_tss.Point
	Status TransitionStatus
	ExecTime time.Time
}
```

## State

### Group & Member

The `x/bandtss` module stores group and member information including their active status on the module.

- CurrentGroupID : `0x00 | "CurrentGroupID" -> BigEndian(groupID)`
- Member: `0x02 | GroupID | MemberAddress -> Member`

### Group Transition

- GroupTransition : `0x00 | "GroupTransition" -> GroupTransition`

### Signing

The `x/bandtss` module stores signing information and mapping between tss SigningID to bandtss SigningID.

- SigningCount : `0x00 | "SigningCount" -> BidEndian(#signing)`
- Signing : `0x03 | BandtssSigningID -> Signing`
- SigningMappingID `0x04 | TssSigningID -> BidEndian(BandtssSigningID)`

### Params

The `x/bandtss` module stores its params in state with the prefix of `0x01`, it can be updated with governance proposal or the address with authority.

- Params: `0x01 -> Params`

## Msg Service

### Msg/TransitionGroup

This message is used to initiate the transition process, which subsequently triggers the creation of a new group and the signing of a transition message. Before the process can proceed, the message requires approval through the submission and approval of a proposal.

It's expected to fail if:

- The members are incorrect (e.g., wrong address format, duplicates).
- The threshold exceeds the number of members.
- The execution time is before the current time or beyond the maximum transition duration.

```protobuf
message MsgTransitionGroup {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "bandtss/MsgTransitionGroup";

  // members is a list of members in this group.
  repeated string members = 1;
  // threshold is a minimum number of members required to produce a signature.
  uint64 threshold = 2;
  // exec_time is the time that will be substituted in place of the group.
  google.protobuf.Timestamp exec_time = 3 [(gogoproto.stdtime) = true, (gogoproto.nullable) = false];
  // authority is the address that controls the module (defaults to x/gov unless overwritten).
  string authority = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/ForceReplaceGroup

A current group can be replaced by an incoming group without needing a signing request (transition message) from the current group.

It's expected to fail if:

- The status of groups is not active.
- Can't request signing `transition message` from `current_group_id`.
- The execution time is before the current time or beyond the maximum transition duration.

```protobuf
message MsgForceReplaceGroup {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "bandtss/ForceReplaceGroup";

  // incoming_group_id is the ID of the group that want to replace.
  uint64 incoming_group_id = 1 [
    (gogoproto.customname) = "IncomingGroupID",
    (gogoproto.casttype)   = "github.com/bandprotocol/chain/v2/pkg/tss.GroupID"
  ];
  // exec_time is the time that will be substituted in place of the group.
  google.protobuf.Timestamp exec_time = 2 [(gogoproto.stdtime) = true, (gogoproto.nullable) = false];
  // authority is the address that controls the module (defaults to x/gov unless overwritten).
  string authority = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/RequestSignature

This message is used to send a signature request to the BandTSS group. It includes parameters such as `fee_limit` and `Content`.

The `Content` is an interface that can be implemented by any module, allowing it to define the logic needed to extract specific data from that module. This enables the module to generate a signature for the provided data, ensuring that the signature request aligns with the module's unique requirements and operations.

The `fee_limit` specifies the maximum amount of fees that can be charged for processing the signature request, ensuring that the costs remain within an acceptable range for the requester.

```protobuf
message MsgRequestSignature {
  option (cosmos.msg.v1.signer)      = "sender";
  option (amino.name)                = "bandtss/MsgRequestSignature";
  option (gogoproto.goproto_getters) = false;

  // content is the signature order of this request signature message.
  google.protobuf.Any content = 1 [(cosmos_proto.accepts_interface) = "Content"];
  // memo is the additional note of the message.
  string memo = 2;
  // fee_limit is the maximum tokens that will be paid for this request.
  repeated cosmos.base.v1beta1.Coin fee_limit = 3
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
  // sender is the requester of the signing process.
  string sender = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/Activate

If members are deactivated due to one of the module's mechanisms, such as a health check or missing signature, they must send `MsgActivate` to rejoin the system. However, there is a punishment period for rejoining the process.

```protobuf
message MsgActivate {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "bandtss/MsgActivate";

  // address is the signer of this message, who must be a member of the group.
  string sender = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // group_id is the group id of the member.
  uint64 group_id = 2
      [(gogoproto.customname) = "GroupID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v2/pkg/tss.GroupID"];
}
```

### Msg/Heartbeat

This message is used by members in the bandtss system. All active members have to regularly send `MsgHeartbeat` to the chain to show if they are still active.

The frequency of sending is determined by `ActiveDuration` parameters.

```protobuf
message MsgHeartbeat {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "bandtss/MsgHeartbeat";

  // address is the signer of this message, who must be a member of the group.
  string sender = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // group_id is the group id of the member.
  uint64 group_id = 2
      [(gogoproto.customname) = "GroupID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v2/pkg/tss.GroupID"];
}
```

### Msg/UpdateParams

When anyone wants to update the parameters of the bandtss module, they will have to open a governance proposal by using the `MsgUpdateParams` of the bandtss module to update those parameters.

## Events

The bandtss module emits the following events:

### EventTypeSigningRequestCreated

This event ( `bandtss_signing_request_created` ) is emitted when the module is requested to sign the data.

| Attribute Key             | Attribute Value    |
| ------------------------- | ------------------ |
| bandtss_signing_id        | {bandtssSigningID} |
| current_group_id          | {groupID}          |
| current_group_signing_id  | {signingID}        |
| incoming_group_id         | {groupID}          |
| incoming_group_signing_id | {signingID}        |

### EventTypeGroupTransition

This event ( `group_transition` ) is emitted when transition is requested via an approved proposal or the transition changes its status.

| Attribute Key          | Attribute Value    |
| ---------------------- | ------------------ |
| signing_id             | {signingID}        |
| current_group_id       | {groupID}          |
| current_group_pub_key  | {groupPubKey}      |
| incoming_group_id      | {groupID}          |
| incoming_group_pub_key | {groupPubKey}      |
| status                 | {transitionStatus} |
| exec_time              | {execute_time}     |

### EventTypeGroupTransitionFailed

This event ( `group_transition_failed` ) is emitted when fail to execute a transition process.

| Attribute Key     | Attribute Value |
| ----------------- | --------------- |
| signing_id        | {signingID}     |
| current_group_id  | {groupID}       |
| incoming_group_id | {groupID}       |

### EventTypeGroupTransitionSuccess

This event ( `group_transition_success` ) is emitted when successfully execute a transition process.

| Attribute Key     | Attribute Value |
| ----------------- | --------------- |
| signing_id        | {signingID}     |
| current_group_id  | {groupID}       |
| incoming_group_id | {groupID}       |

### EventTypeActivate

This event ( `activate` ) is emitted when an account submitted `MsgActivate` to the chain

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| address       | {memberAddress} |
| group_id      | {groupID}       |

### EventTypeHeartbeat

This event ( `heartbeat` ) is emitted when an account submitted `MsgHeartbeat` to the chain

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| address       | {memberAddress} |
| group_id      | {groupID}       |

### EventTypeInactiveStatus

This event ( `inactive_status` ) is emitted when an account is deactivated

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| address       | {memberAddress} |
| group_id      | {groupID}       |

## Parameters

The module contains the following parameters

```protobuf
message Params {
  // active_duration is the duration where a member is active without interaction.
  google.protobuf.Duration active_duration = 1 [(gogoproto.stdduration) = true, (gogoproto.nullable) = false];
  // reward_percentage is the percentage of block rewards allocated to active TSS members.
  // The reward proportion is calculated after being allocated to oracle rewards.
  uint64 reward_percentage = 2 [(gogoproto.customname) = "RewardPercentage"];
  // inactive_penalty_duration is the duration where a member cannot activate back after being set to inactive.
  google.protobuf.Duration inactive_penalty_duration = 3 [(gogoproto.stdduration) = true, (gogoproto.nullable) = false];
  // max_transition_duration is the maximum duration where the transition process waits
  // since the start of the process until an incoming group replaces a current group.
  google.protobuf.Duration max_transition_duration = 4 [(gogoproto.stdduration) = true, (gogoproto.nullable) = false];
  // fee is the tokens that will be paid per signer.
  repeated cosmos.base.v1beta1.Coin fee = 5
      [(gogoproto.nullable) = false, (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.Coins"];
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

##### IncomingGroup

The `incoming-group` command allows users to query for incoming group information.

```bash
bandd query bandtss incoming-group
```

##### Count Signing

The `counts` command allows users to query a number of bandtss signing in the chain.

```bash
bandd query bandtss counts
```

##### Member

The `Member` command allows users to query for member information by giving a member address, both information in the current group and incoming group.

```bash
bandd query bandtss member [address] [flags]
```

Example:

```bash
bandd query bandtss member band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q
```

##### GroupTransition

The `GroupTransition` command allows users to query for group transition information.

```bash
bandd query bandtss group-transition
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

##### IncomingGroup

The `incoming-group` command allows users to query for incoming group information.

```bash
bandtss.v1beta1.Query/IncomingGroup
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

##### GroupTransition

The `GroupTransition` command allows users to query for group transition information.

```bash
bandtss.v1beta1.Query/GroupTransition
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
/bandtss/v1beta1/current_group
```

##### IncomingGroup

The `incoming-group` command allows users to query for incoming group information.

```bash
/bandtss/v1beta1/incoming_group
```

##### Member

The `Member` command allows users to query for member information by giving a member address.

```bash
/bandtss/v1beta1/members/{address}
```

Example:

```bash
curl localhost:1317/bandtss/v1beta1/members/band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q
```

##### GroupTransition

The `GroupTransition` command allows users to query for group transition information.

```bash
/bandtss/v1beta1/group_transition
```

##### Signing

The `Signing` command allows users to query for bandtss signing information by giving a signing id.

```bash
/bandtss/v1beta1/signing/{id}
```

Example:

```bash
curl localhost:1317/bandtss/v1beta1/signing/1
```

##### Params

The `Params` command allows users to query for module's configuration.

```bash
/bandtss/v1beta1/params
```
