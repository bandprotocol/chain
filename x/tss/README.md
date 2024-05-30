# `x/tss`

## Abstract

The TSS module's main purpose is to manage the threshold signature scheme (TSS) signing process, allowing other system modules to utilize this method for cryptographic signing.

To handle a signing process, the module has to create a group with selected members. These selected members then submit encrypted secret shares to create a public shared secret of the group, which is subsequently formed and owned by the caller module.

Once the group is established, the group's owner can request specific signatures. The resulting group signature, which can be verified using the group's public key, proves useful in various situations, rendering the TSS module quite valuable. This method of creating signatures not only ensures trust among all participants but also adds an extra layer of security to the system.

This module is used in bandtss module in BandChain.

## Contents

- [`x/tss`](#xtss)
  - [Abstract](#abstract)
  - [Contents](#contents)
  - [Concepts](#concepts)
    - [Group](#group)
    - [Signing](#signing)
  - [State](#state)
  - [Msg Service](#msg-service)
    - [Msg/SubmitDKGRound1](#msgsubmitdkground1)
    - [Msg/SubmitDKGRound2](#msgsubmitdkground2)
    - [Msg/Complain](#msgcomplain)
    - [Msg/Confirm](#msgconfirm)
    - [Msg/SubmitDEs](#msgsubmitdes)
    - [Msg/SubmitSignature](#msgsubmitsignature)
    - [Msg/UpdateParams](#msgupdateparams)
  - [Events](#events)
    - [EventTypeSubmitDKGRound1](#eventtypesubmitdkground1)
    - [EventTypeRound1Success](#eventtyperound1success)
    - [EventTypeSubmitDKGRound2](#eventtypesubmitdkground2)
    - [EventTypeRound2Success](#eventtyperound2success)
    - [EventTypeComplainSuccess](#eventtypecomplainsuccess)
    - [EventTypeComplainFailed](#eventtypecomplainfailed)
    - [EventTypeConfirmSuccess](#eventtypeconfirmsuccess)
    - [EventTypeRound3Success](#eventtyperound3success)
    - [EventTypeRound3Failed](#eventtyperound3failed)
    - [EventTypeRequestSignature](#eventtyperequestsignature)
    - [EventTypeSigningSuccess](#eventtypesigningsuccess)
    - [EventTypeReplaceSuccess](#eventtypereplacesuccess)
    - [EventTypeSubmitSignature](#eventtypesubmitsignature)
    - [EventTypeSigningFailed](#eventtypesigningfailed)
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

### Group

A group contains multiple members. Each group has its public key that multiple members (at least the threshold of the group) will be able to generate signatures on the message of that public key.

A group is created through a call by external module with a set of selected members. At first, when creating a group, each assigned member will have to go through a key generation process to generate a group key together. After that, they will receive their private key that will be used to generate part of the signature of the group.

### Signing

A module creates a signing request to the group that the module owns. It contains all information of this request such as message, assigned members, and assigned nonce of each member. When a user requests a signing from the group, each member will have to use the key of the group to sign on the message that will combine to generate the final signature of the group.

## State

The `x/tss` module keeps the state of the following primary objects:

1. Groups
2. Signings
3. DEs (Nonces being used in a signing process)

In addition, the `x/tss` module still keeps temporary information such as group count, round1Info, round2Info, queue of replacements, groups, and partial signings information.

Here are the prefixes for each object in the KVStore of the TSS module.

```go
var (
	GlobalStoreKeyPrefix = []byte{0x00}
	GroupCountStoreKey = append(GlobalStoreKeyPrefix, []byte("GroupCount")...)
	LastExpiredGroupIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredGroupID")...)
	SigningCountStoreKey = append(GlobalStoreKeyPrefix, []byte("SigningCount")...)
	LastExpiredSigningIDStoreKey = append(GlobalStoreKeyPrefix, []byte("LastExpiredSigningID")...)
	PendingProcessGroupsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessGroups")...)
	PendingSigningsStoreKey = append(GlobalStoreKeyPrefix, []byte("PendingProcessSignings")...)
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
	DECountStoreKeyPrefix = []byte{0x0d}
	SigningStoreKeyPrefix = []byte{0x0e}
	PartialSignatureCountStoreKeyPrefix = []byte{0x0f}
	PartialSignatureStoreKeyPrefix = []byte{0x10}
	ParamsKeyPrefix = []byte{0x11}
)
```

## Msg Service

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

- The number of remaining DEs exceeds the maximum size (`MaxDESize`) per user.

### Msg/SubmitSignature

When a user requests a signature from the group, the assigned member of the group is required to send `MsgSubmitSignature` to the chain. It contains `signing_id`, `member_id`, `address`, and `signature`.

Once all assigned member sends their signature to the chain, the chain will aggregate those signatures to be the final signature of the group for that request.

### Msg/UpdateParams

When anyone wants to update the parameters of the TSS module, they will have to open a governance proposal by using the `MsgUpdateParams` of the TSS module to update those parameters.

## Events

The TSS module emits the following events:

### EventTypeCreateGroup

This event ( `create_group` ) is emitted when the group is created.

| Attribute Key | Attribute Value   |
| ------------- | ----------------- |
| group_id      | {groupID}         |
| size          | {groupSize}       |
| thredhold     | {groupThreshold}  |
| pub_key       | ""                |
| status        | {groupStatus}     |
| dkg_context   | {groupDKGContext} |
| module_owner  | ""                |

### EventTypeSubmitDKGRound1

This event ( `submit_dkg_round1` ) is emitted when a member submits round 1 information of the DKG process.

| Attribute Key | Attribute Value  |
| ------------- | ---------------- |
| group_id      | {groupID}        |
| member_id     | {groupSize}      |
| threshold     | {groupThreshold} |
| round1_info   | {round1Info}     |

### EventTypeRound1Success

This event ( `round1_success` ) is emitted at the end block when all members of the group submit round 1 information.

| Attribute Key | Attribute Value        |
| ------------- | ---------------------- |
| group_id      | {groupID}              |
| status        | "GROUP_STATUS_ROUND_2" |

### EventTypeSubmitDKGRound2

This event ( `submit_dkg_round2` ) is emitted when a member submits information about round 2 in the DKG process.

| Attribute Key | Attribute Value  |
| ------------- | ---------------- |
| group_id      | {groupID}        |
| member_id     | {groupSize}      |
| threshold     | {groupThreshold} |
| round2_info   | {round2Info}     |

### EventTypeRound2Success

This event ( `round2_success` ) is emitted at the end block when all members of the group submit round 2 information.

| Attribute Key | Attribute Value        |
| ------------- | ---------------------- |
| group_id      | {groupID}              |
| status        | "GROUP_STATUS_ROUND_3" |

### EventTypeComplainSuccess

This event ( `complain_success` ) is emitted when a member submits `MsgComplain` and the complaint is successful.

| Attribute Key  | Attribute Value |
| -------------- | --------------- |
| group_id       | {groupID}       |
| complainant_id | {complianantID} |
| respondent_id  | {respondentID}  |
| key_sym        | {keySym}        |
| signature      | {signature}     |
| address        | {memberAddress} |

### EventTypeComplainFailed

This event ( `complain_failed` ) is emitted when a member submits `MsgComplain` and the complaint fails

| Attribute Key  | Attribute Value |
| -------------- | --------------- |
| group_id       | {groupID}       |
| complainant_id | {complianantID} |
| respondent_id  | {respondentID}  |
| key_sym        | {keySym}        |
| signature      | {signature}     |
| address        | {memberAddress} |

### EventTypeConfirmSuccess

This event ( `confirm_success` ) is emitted when a member submits `MsgComfirm` and the confirmation is successful.

| Attribute Key   | Attribute Value         |
| --------------- | ----------------------- |
| group_id        | {groupID}               |
| member_id       | {memberID}              |
| own_pub_key_sig | {ownPublicKeySignature} |
| address         | {memberAddress}         |

### EventTypeRound3Success

This event ( `round3_success` ) is emitted at the end block when all members of the group submit round 3 information ( `MsgConfirm` / `MsgComplain` ) and the process is successful.

| Attribute Key | Attribute Value       |
| ------------- | --------------------- |
| group_id      | {groupID}             |
| status        | "GROUP_STATUS_ACTIVE" |

### EventTypeRound3Failed

This event ( `round3_failed` ) is emitted at the end block when all members of the group submit round 3 information ( `MsgConfirm` / `MsgComplain` ) and the process fails.

| Attribute Key | Attribute Value       |
| ------------- | --------------------- |
| group_id      | {groupID}             |
| status        | "GROUP_STATUS_FALLEN" |

### EventTypeRequestSignature

This event ( `request_signature` ) is emitted when the group is requested to sign the data.

| Attribute Key    | Attribute Value                |
| ---------------- | ------------------------------ |
| group_id         | {groupID}                      |
| signing_id       | {signingID}                    |
| message          | {message}                      |
| group_pub_nonce  | {groupPublicNonce}             |
| member_id[]      | {assignedMemberIDs}            |
| address[]        | {assignedMemberAddresses}      |
| binding_factor[] | {assignedMemberBindingFactors} |
| pub_nonce[]      | {assignedMemberPublicNonces}   |
| pub_d[]          | {assignedMemberPublicDs}       |
| pub_e[]          | {assignedMemberPublicEs}       |

### EventTypeSigningSuccess

This event ( `signing_success` ) is emitted at the end block when all assigned members submit their signatures and the aggregation process is successful.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| signing_id    | {signingID}     |
| group_id      | {groupID}       |
| signature     | {signature}     |

### EventTypeSubmitSignature

This event ( `submit_signature` ) is emitted when an assigned member submits his or her signature on the signing request.

| Attribute Key | Attribute Value           |
| ------------- | ------------------------- |
| signing_id    | {signingID}               |
| group_id      | {groupID}                 |
| member_id     | {assignedMemberID}        |
| address       | {assignedMemberAddress}   |
| pub_d         | {assignedMemberPublicD}   |
| pub_e         | {assignedMemberPublicE}   |
| signature     | {assignedMemberSignature} |

### EventTypeSigningFailed

This event ( `signing_failed` ) is emitted at the end block when all assigned members submit their signatures and the aggregation process fails.

| Attribute Key | Attribute Value |
| ------------- | --------------- |
| signing_id    | {signingID}     |
| group_id      | {groupID}       |
| reason        | {failedReason}  |

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
}
```

## Client

### CLI

A user can query and interact with the `TSS` module using the CLI.

#### Query

The `query` commands allow users to query the `group` state.

```bash
bandd query tss --help
```

##### Group

The `Group` command allows users to query for group information by given group ID.

```bash
bandd query tss group [id] [flags]
```

Example:

```bash
bandd query tss group 1
```

##### Signing

The `Signing` command allows users to query for signing information by giving a signing ID.

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

The `Group` endpoint allows users to query for group information by given group ID.

```bash
tss.v1beta1.Query/Group
```

Example:

```bash
grpcurl -plaintext \
-d '{"group_id":1}' localhost:9090 tss.v1beta1.Query/Group
```

#### Signing

The `Signing` endpoint allows users to query for signing information by giving a signing ID.

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

The `Group` endpoint allows users to query for group information by given group ID.

```bash
/tss/v1beta1/groups/{group_id}
```

Example:

```bash
curl localhost:1317/tss/v1beta1/groups/1
```

#### Signing

The `Signing` endpoint allows users to query for signing information by giving a signing ID.

```bash
/tss/v1beta1/signings/{signing_id}
```

Example:

```bash
curl localhost:1317/tss/v1beta1/signings/{signing_id}
```
