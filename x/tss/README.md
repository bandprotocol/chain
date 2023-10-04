# `x/tss`

## Abstract

This document specifies the TSS module.

The TSS module is responsible for generating threshold signatures of groups from members in the group.
It also includes logic on the key generation process of the group between members to get one public key of the group and each one has their private key for generating the signature of the group.

In addition, the TSS module tracks and provides query support for requesting signatures and information of groups.

This module is used in the BandChain.

## Contents

* [Concepts](#concepts)
  * [Status](#status)
  * [Reward](#reward)
  * [Group](#group)
  * [Signing](#signing)
  * [Group replacement](#groupreplacement)
* [State](#state)
* [Msg Service](#msg-service)
  * [Msg/CreateGroup](#msgcreategroup)
  * [Msg/ReplaceGroup](#msgreplacegroup)
  * [Msg/UpdateGroupFee](#msgupdategroupfee)
  * [Msg/SubmitDKGRound1](#msgsubmitdkground1)
  * [Msg/SubmitDKGRound2](#msgsubmitdkground2)
  * [Msg/Complain](#msgcomplain)
  * [Msg/Confirm](#msgconfirm)
  * [Msg/SubmitDEs](#msgsubmitdes)
  * [Msg/RequestSignature](#msgrequestsignature)
  * [Msg/SubmitSignature](#msgsubmitsignature)
  * [Msg/Activate](#msgactivate)
  * [Msg/Active](#msgactive)
  * [Msg/UpdateParams](#msgupdateparams)
* [Events](#events)
  * [EventTypeCreateGroup](#eventtypecreategroup)
  * [EventTypeSubmitDKGRound1](#eventtypesubmitdkground1)
  * [EventTypeRound1Success](#eventtyperound1success)
  * [EventTypeSubmitDKGRound2](#eventtypesubmitdkground2)
  * [EventTypeRound2Success](#eventtyperound2success)
  * [EventTypeComplainSuccess](#eventtypecomplainsuccess)
  * [EventTypeComplainFailed](#eventtypecomplainfailed)
  * [EventTypeConfirmSuccess](#eventtypeconfirmsuccess)
  * [EventTypeRound3Success](#eventtyperound3success)
  * [EventTypeRound3Failed](#eventtyperound3failed)
  * [EventTypeRequestSignature](#eventtyperequestsignature)
  * [EventTypeSigningSuccess](#eventtypesigningsuccess)
  * [EventTypeReplaceSuccess](#eventtypereplacesuccess)
  * [EventTypeSubmitSignature](#eventtypesubmitsignature)
  * [EventTypeSigningFailed](#eventtypesigningfailed)
  * [EventTypeActivate](#eventtypeactivate)
  * [EventTypeActive](#eventtypeactive)
* [Parameters](#parameters)
* [Client](#client)
  * [CLI](#cli)
  * [gRPC](#grpc)
  * [REST](#rest)


## Concepts

### Status

A status in the TSS system is on the account level. An account must send a message to the chain to show if they want to be a participant in the TSS system. Once they have activated their status in the TSS module. They will have to send a health-check message to the chain every "ActiveDuration" (default is 1 day) to show if they are still active.

If an account didn't send a health-check message or failed to participate in any assigned actions such as creating a group or requesting signature. They will be deactivated for a specific period depending on the action. This mechanism will help to eliminate inactive accounts from the TSS system and improve the reliability of the system.


### Reward

#### Block rewards

In each block, all active accounts that are validators will receive more block rewards depending on their validating power as a reward for providing service on the TSS system.

The `RewardPercentage` parameter will be the percent of block rewards that will be assigned to those validators. The default value is 50%. However, this percentage is calculated from the remaning rewards. For example, if somehow other module took 40% as their rewards. TSS module will receive only 30% (50% of 60%) of the full block rewards.

#### Request fee

All users who request signatures on data from the TSS group will have to pay the fee for the TSS service. The fee will depend on the group. Only assigned accounts of the request will receive this fee as a reward for providing service to the group on top of block rewards.


### Group

A group contains multiple members. Each group has its public key that multiple members (at least the threshold of the group) will be able to generate signatures on the message of that public key.

A group will be created through a governance proposal at this phase. At first, when creating a group, each assigned member will have to go through a key generation process to generate a group key together. After that, they will receive their private key that will be used to generate part of the signature of the group.


### Signing

A signing is a request to sign some data from a user to the group. It contains all information of this request such as message, assigned members, and assigned nonce of each member. When a user requests a signing from the group, each member will have to use their key of the group to sign on the message that will combine to generate the final signature of the group.


### Group replacement

The process of group replacement is used when we need to change who is in a group and also update the group's key. We can't just swap out individual members because their keys are linked to the group's key. To replacement the group, we have to create a new group and then update the old group's information with the new group's details.

Here are the step of replcement process:
1. Create a new group through a proposal
2. Create a group replacement proposal with replacement time
3. After the proposal passed, the old group will be assigned to sign the `changing group` message
4. Once it reaches replacement time, all information from the old group will be replaced by information from the new group.

This process allows users to have spare time to update their key before it reaches replacement time. Also, users can choose to request from old and new group IDs.

## State

The `x/tss` module keeps the state of the following primary objects:

1. Groups
2. Signings
3. Statuses
4. Nonces (DEs)
5. Replacements

In addition, the `x/tss` module still keeps temporary information such as group count, round1Info, round2Info, queue of replacements, groups, and signings.

Here are the prefixes for each object in the KVStore of the TSS module.
```go
var (
	GlobalStoreKeyPrefix = []byte{0x00}
	GroupCountStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupCount")...)
	ReplacementCountStoreKey = append(GlobalStoreKeyPrefix, []byte("ReplacementCount")...)
	LastExpiredGroupIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredGroupID")...)
	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)
	LastExpiredSigningIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredSigningID")...)
	RollingSeedStoreKey = append(GlobalStoreKeyPrefix, []byte("RollingSeed")...)
	PendingProcessGroupsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessGroups")...)
	PendingSigningsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessSignings")...)
	PendingReplaceGroupsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingReplaceGroups")...)
	GroupStoreKeyPrefix = []byte{0x01}
	DKGContextStoreKeyPrefix = []byte{0x02}
	MemberStoreKeyPrefix = []byte{0x03}
	Round1InfoStoreKeyPrefix = []byte{0x04}
	Round1InfoCountStoreKeyPrefix = []byte{0x05}
	AccumulatedCommitStoreKeyPrefix = []byte{0x06}
	Round2InfoStoreKeyPrefix = []byte{0x07}
	Round2InfoCountStoreKeyPrefix = []byte{0x08}
	ComplainsWithStatusStoreKeyPrefix = []byte{0x09}
	ConfirmComplainCountStoreKeyPrefix = []byte{0x0a}
	ConfirmStoreKeyPrefix = []byte{0x0b}
	DEStoreKeyPrefix = []byte{0x0c}
	DEQueueStoreKeyPrefix = []byte{0x0d}
	SigningStoreKeyPrefix = []byte{0x0e}
	SigCountStoreKeyPrefix = []byte{0x0f}
	PartialSigStoreKeyPrefix = []byte{0x10}
	StatusStoreKeyPrefix = []byte{0x11}
	ParamsKeyPrefix = []byte{0x12}
	ReplacementKeyPrefix = []byte{0x13}
	ReplacementQueuePrefix = []byte{0x14}
)
```

## Msg Service

### Msg/CreateGroup

A new group can be created with the `MsgCreateGroup` which needs to open through governance proposal.
This message contains the list of members, the threshold of the group, and the fee for requesting.

It's expected to fail if:

* The number of members is greater than the `MaxGroupSize` parameters.
* One of the members has inactive TSS status.
* Members are not correct (e.g. wrong address format, duplicates).


### Msg/ReplaceGroup

A replacement can be created with the `MsgReplaceGrouup` which needs to open through a governance proposal.
This message contains `from_group_id`, `to_group_id`, and `exec_time`.

It's expected to fail if:

* The status of groups is not active.
* The `to_group_id` is in the replacement process.
* Can't request signing `changing group` message from `to_group_id`


### Msg/UpdateGroupFee

A changing fee of the group can be created with the `MsgUpdateGroupFee` which needs to open through the governance proposal. This message contains the ID of the group and the new fee.

It's expected to fail if:

* The group doesn't exist.


### Msg/SubmitDKGRound1

This message is used to send round 1 information in the DKG process of the group.

When a group is created, all members of the group are required to send this message to the chain. So, the chain can proceed to the next step of the DKG process.


### Msg/SubmitDKGRound2

This message is used to send round 2 information in the DKG process of the group.

When a group is passed round 1, all members of the group are required to send this message to the chain. So, the chain can proceed to the next step of the DKG process.


### Msg/Complain

This message is used to complain to any malicious member of the group if their shared secret data doesn't align with public information.

A member can send this message when the group is in round 3 of the DKG process. If there is one valid `MsgComplain` in this round, the group creation process will fail and the malicious member will be punished.


### Msg/Confirm

This message is used to confirm that all information from other members is correct.

A member can send this message when the group is in round 3 of the DKG process. They are required to send `MsgConfirm` or `MsgComplain` in this process. Otherwise, they will be deactivated from the TSS system.


### Msg/SubmitDEs

In the signing process, each member is required to have their nonces (D and E values). `MsgSubmitDEs` is the message for a member to send their public nonce to the chain. So, the chain can assign their nonce in the signing process.


It's expected to fail if:

* The number of remaining DEs exceeds the maximum size (`MaxDESize`) per user.


### Msg/RequestSignature

Anyone who wants to have a signature from the group can use `MsgRequestSignature` to send their message to the group to request a signature.

It contains `group_id`, `fee_limit`, and `content`. `content` is an interface that any module can implement to have its logic to get the specific data from its module so that the TSS module can get a signature for that data.


### Msg/SubmitSignature

When a user requests a signature from the group, the assigned member of the group is required to send `MsgSubmitSignature` to the chain. It contains `signing_id`, `member_id`, `address`, and `signature`.

Once all assigned member sends their signature to the chain, the chain will aggregate those signatures to be the final signature of the group for that request.


### Msg/Activate

An account that wants to participate as a TSS provider (signature provider) has to activate its TSS status through `MsgActivate`.

If the account is deactivated by one of the TSS mechanisms (such as a health check, or missing signature), they will have to send `MsgActivate` again to rejoin the system. However, there is a punishment period for rejoining depending on the action that the account got deactivated.


### Msg/Active

This message is used by participators in the TSS system. All active TSS accounts have to regularly send `MsgActive` to the chain to show if they are still active.

The frequency of sending is determined by `ActiveDuration` parameters.

### Msg/UpdateParams

When anyone wants to update the parameters of the TSS module, they will have to open a governance proposal by using the `MsgUpdateParams` of the TSS module to update those parameters.


## Events

The TSS module emits the following events:

### EventTypeCreateGroup

This event (`create_group`) is emitted when the group is created.

| Attribute Key | Attribute Value   |
| ------------- | ----------------- |
| group_id      | {groupID}         |
| size          | {groupSize}       |
| thredhold     | {groupThreshold}  |
| fee           | {groupFee}        |
| pub_key       | ""                |
| status        | {groupStatus}     |
| dkg_context   | {groupDKGContext} |


### EventTypeSubmitDKGRound1

This event (`submit_dkg_round1`) is emitted when a member submits round 1 information of the DKG process.

| Attribute Key | Attribute Value  |
| ------------- | ---------------- |
| group_id      | {groupID}        |
| member_id     | {groupSize}      |
| member        | {groupThreshold} |
| round1_info   | {round1Info}     |


### EventTypeRound1Success

This event (`round1_success`) is emitted at the end block when all members of the group submit round 1 information.

| Attribute Key | Attribute Value        |
| ------------- | ---------------------- |
| group_id      | {groupID}              |
| status        | "GROUP_STATUS_ROUND_2" |


### EventTypeSubmitDKGRound2

This event (`submit_dkg_round2`) is emitted when a member submits round 2 information of the DKG process.

| Attribute Key | Attribute Value  |
| ------------- | ---------------- |
| group_id      | {groupID}        |
| member_id     | {groupSize}      |
| member        | {groupThreshold} |
| round2_info   | {round2Info}     |


### EventTypeRound2Success

This event (`round2_success`) is emitted at the end block when all members of the group submit round 2 information.

| Attribute Key | Attribute Value        |
| ------------- | ---------------------- |
| group_id      | {groupID}              |
| status        | "GROUP_STATUS_ROUND_3" |


### EventTypeComplainSuccess

This event (`complain_success`) is emitted when a member submits `MsgComplain` and the complaint is successful.

| Attribute Key  | Attribute Value |
| -------------- | --------------- |
| group_id       | {groupID}       |
| complainant_id | {complianantID} |
| respondent_id  | {respondentID}  |
| key_sym        | {keySym}        |
| signature      | {signature}     |
| member         | {memberAddress} |


### EventTypeComplainFailed

This event (`complain_failed`) is emitted when a member submits `MsgComplain` and the complaint fails

| Attribute Key  | Attribute Value |
| -------------- | --------------- |
| group_id       | {groupID}       |
| complainant_id | {complianantID} |
| respondent_id  | {respondentID}  |
| key_sym        | {keySym}        |
| signature      | {signature}     |
| member         | {memberAddress} |


### EventTypeConfirmSuccess

This event (`confirm_success`) is emitted when a member submits `MsgComfirm` and the confirmation is successful.

| Attribute Key   | Attribute Value         |
| --------------- | ----------------------- |
| group_id        | {groupID}               |
| member_id       | {memberID}              |
| own_pub_key_sig | {ownPublicKeySignature} |
| member          | {memberAddress}         |


### EventTypeRound3Success

This event (`round3_success`) is emitted at the end block when all members of the group submit round 3 information (`MsgConfirm` / `MsgComplain`) and the process is successful.

| Attribute Key | Attribute Value       |
| ------------- | --------------------- |
| group_id      | {groupID}             |
| status        | "GROUP_STATUS_ACTIVE" |


### EventTypeRound3Failed

This event (`round3_failed`) is emitted at the end block when all members of the group submit round 3 information (`MsgConfirm` / `MsgComplain`) and the process fails.

| Attribute Key | Attribute Value       |
| ------------- | --------------------- |
| group_id      | {groupID}             |
| status        | "GROUP_STATUS_FALLEN" |


### EventTypeRequestSignature

This event (`request_signature`) is emitted when the group is requested to sign the data.

| Attribute Key    | Attribute Value                |
| ---------------- | ------------------------------ |
| group_id         | {groupID}                      |
| signing_id       | {signingID}                    |
| message          | {message}                      |
| group_pub_nonce  | {groupPublicNonce}             |
| member_id[]      | {assignedMemberIDs}            |
| member[]         | {assignedMemberAddresses}      |
| binding_factor[] | {assignedMemberBindingFactors} |
| pub_nonce[]      | {assignedMemberPublicNonces}   |
| pub_d[]          | {assignedMemberPublicDs}       |
| pub_e[]          | {assignedMemberPublicEs}       |


### EventTypeSigningSuccess

This event (`signing_success`) is emitted at the end block when all assigned members submit their signatures and the aggregation process is successful.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| signing_id    | {signingID}     |
| group_id      | {groupID}       |
| signature     | {signature}     |



### EventTypeReplaceSuccess

This event (`replace_success`) is emitted at the end block when it reaches replacement time and replacement is successful.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| signing_id    | {signingID}     |
| from_group_id | {fromGroupID}   |
| to_group_id   | {toGroupID}     |



### EventTypeSubmitSignature

This event (`submit_signature`) is emitted when an assigned member submits his or her signature of the signing request.

| Attribute Key | Attribute Value           |
| ------------- | ------------------------- |
| signing_id    | {signingID}               |
| group_id      | {groupID}                 |
| member_id     | {assignedMemberID}        |
| member        | {assignedMemberAddress}   |
| pub_d         | {assignedMemberPublicD}   |
| pub_e         | {assignedMemberPublicE}   |
| signature     | {assignedMemberSignature} |



### EventTypeSigningFailed

This event (`signing_failed`) is emitted at the end block when all assigned members submit their signature and the aggregation process fails.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| signing_id    | {signingID}     |
| group_id      | {groupID}       |
| reason        | {failedReason}  |


### EventTypeActivate

This event (`activate`) is emitted when an account submitted `MsgActivate` to the chain

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| member        | {memberAddress} |


### EventTypeActive

This event (`active`) is emitted when an account submitted `MsgActive` to the chain

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| member        | {memberAddress} |


## Parameters

The TSS module contains the following parameters

```protobuf
type Params struct {
	// MaxGroupSize is the maximum of the member capacity of the group.
	MaxGroupSize uint64
	// MaxDESize is the maximum of the de capacity of the member.
	MaxDESize uint64
	// CreatingPeriod is the number of blocks allowed to creating group signature.
	CreatingPeriod uint64
	// SigningPeriod is the number of blocks allowed to sign.
	SigningPeriod uint64
	// ActiveDuration is the duration where a member can be active without interaction.
	ActiveDuration time.Duration
	// InactivePenaltyDuration is the duration where a member cannot activate back after inactive.
	InactivePenaltyDuration time.Duration
	// JailPenaltyDuration is the duration where a member cannot activate back after jail.
	JailPenaltyDuration time.Duration
	// RewardPercentage is the percentage of block rewards allocated to active TSS validators after being allocated to oracle rewards.
	RewardPercentage uint64
}
```


## Client

### CLI

A user can query and interact with the `TSS` module using the CLI.

#### Query

The `query` commands allow users to query `group` state.

```bash
bandd query tss --help
```

##### Group

The `Group` command allows users to query for group information by given group id.

```bash
bandd query tss group [id] [flags]
```

Example:

```bash
bandd query tss group 1
```

##### Signing

The `Signing` command allows users to query for signing information by given signing id.

```bash
bandd query tss signing [id] [flags]
```

Example:

```bash
bandd query tss signing 1
```

### gRPC

A user can query the `TSS` module using gRPC endpoints.

#### Group

The `Group` endpoint allows users to query for group information by given group id.

```bash
tss.v1beta1.Query/Group
```

Example:

```bash
grpcurl -plaintext \
-d '{"group_id":1}' localhost:9090 tss.v1beta1.Query/Group
```


#### Signing

The `Signing` endpoint allows users to query for signing information by given signing id.

```bash
tss.v1beta1.Query/Signing
```

Example:

```bash
grpcurl -plaintext \
-d '{"address":"cosmos1.."}' localhost:9090 tss.v1beta1.Query/Signing
```


### REST

A user can query the `TSS` module using REST endpoints.

#### Group

The `Group` endpoint allows users to query for group information by given group id.

```bash
/tss/v1beta1/groups/{group_id}
```

Example:

```bash
curl localhost:1317/tss/v1beta1/groups/1
```

#### Signing

The `Signing` endpoint allows users to query for signing information by given signing id.

```bash
/tss/v1beta1/signings/{signing_id}
```

Example:

```bash
curl localhost:1317/tss/v1beta1/signings/{signing_id}
```
