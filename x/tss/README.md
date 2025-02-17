# `x/tss`

## Abstract

The TSS module's main purpose is to manage the threshold signature scheme (TSS) signing process, allowing other system modules to utilize this method for cryptographic signing.

To handle a signing process, the module has to create a group with selected members. These selected members then submit encrypted secret shares to create a public shared secret of the group, which is subsequently formed and owned by the caller module.

Once the group is established, the group's owner can request specific signatures. The resulting group signature, which can be verified using the group's public key, proves useful in various situations, rendering the TSS module quite valuable. This method of creating signatures not only ensures trust among all participants but also adds an extra layer of security to the system.

## Contents

- [`x/tss`](#xtss)
  - [Abstract](#abstract)
  - [Contents](#contents)
  - [Concepts](#concepts)
    - [Group](#group)
    - [Member](#member)
    - [Group Creation](#group-creation)
    - [Signing](#signing)
    - [DE](#de)
  - [State](#state)
    - [Group](#group-1)
    - [Member](#member-1)
    - [GroupCreation](#group-creation-1)
    - [Signing](#signing-1)
    - [DE](#de-1)
    - [Params](#params)
  - [Msg Service](#msg-service)
    - [Msg/SubmitDKGRound1](#msgsubmitdkground1)
    - [Msg/SubmitDKGRound2](#msgsubmitdkground2)
    - [Msg/Complain](#msgcomplain)
    - [Msg/Confirm](#msgconfirm)
    - [Msg/SubmitDEs](#msgsubmitdes)
    - [Msg/SubmitSignature](#msgsubmitsignature)
    - [Msg/UpdateParams](#msgupdateparams)
  - [Callbacks](#callbacks)
    - [OnGroupCreationCompleted](#ongroupcreationcompleted)
    - [OnGroupCreationFailed](#ongroupcreationfailed)
    - [OnGroupCreationExpired](#ongroupcreationexpired)
    - [OnSigningCompleted](#onsigningcompleted)
    - [OnSigningFailed](#onsigningfailed)
    - [OnSigningTimeout](#onsigningtimeout)
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
    - [EventTypeSubmitSignature](#eventtypesubmitsignature)
    - [EventTypeSigningFailed](#eventtypesigningfailed)
  - [Client](#client)
    - [CLI](#cli)
    - [gRPC](#grpc)
    - [REST](#rest)

## Concepts

### Group

The `x/tss` module defines a `Group` type to which a user requests a signature on a specific message. A group contains a group public key and a threshold, which specifies the number of members being required for creating a group signature.

A group is formed through a call by external module with a set of selected members. In a group creation process, each assigned member has to go through a key generation process. During the process, they receive their private key that will be used to generate part of the signature of the group.

```go
type Group struct {
	ID github_com_bandprotocol_chain_v2_pkg_tss.GroupID
	Size_ uint64
	Threshold uint64
	PubKey github_com_bandprotocol_chain_v2_pkg_tss.Point
	Status GroupStatus
	CreatedHeight uint64
	ModuleOwner string
}
```

### Member

The `x/tss` module defines a `Member` type to represent each participant's status and public key within a specific group. Member's public key is obtained during the group creation process. The member's status `IsActive` indicates if the member should be selected during the signing process.

```go
type Member struct {
	ID github_com_bandprotocol_chain_v2_pkg_tss.MemberID
	GroupID github_com_bandprotocol_chain_v2_pkg_tss.GroupID
	Address string
	PubKey github_com_bandprotocol_chain_v2_pkg_tss.Point
	IsMalicious bool
	IsActive bool
}
```

### Group Creation

When other module call the `x/tss` module to create a group, the `x/tss` module will emit an event and wait for the selected members to submit their information to generate a group public key. The steps required for members to generate a group public keys are the following.

First, members submit a coefficient commit and temporary public key for decrypting a message with their signature on the commit and their public key, which the `x/tss` module will check against if their input is valid. If every member submits their information (round-1 information), the `x/tss` module will aggregate members' coefficient commit, emit an event, and waiting members to submit other information.

After the members acknowledge that other members submit their coefficient commit, they generate a secret, called encrypted secret share, to each member and submit into a chain as a round-2 information. Once, every member submits round-2 information, the `x/tss` module emits an event to notify members.

Once the members are notified, they validate those public encrypted secrets that stored in the chain and send a confirm message if those secrets are valid, or a complaint message on specific members if not. If every member submits their confirm message, the group will be successfully created and, if any, trigger a callback to a requester module to process further on that module.

### Signing

The `x/tss` module defines a `Signing` type that stores all information about the signing request, including message, and assigned nonces of each member. When a user requests a signing from the group, each member must use the key of the group to sign on the message which will then be combined to generate the final signature of the group.

```go
type Signing struct {
  ID github_com_bandprotocol_chain_v2_pkg_tss.SigningID
  CurrentAttempt uint64
  GroupID github_com_bandprotocol_chain_v2_pkg_tss.GroupID
  GroupPubKey github_com_bandprotocol_chain_v2_pkg_tss.Point
  Originator github_com_cometbft_cometbft_libs_bytes.HexBytes
  Message github_com_cometbft_cometbft_libs_bytes.HexBytes
  GroupPubNonce github_com_bandprotocol_chain_v2_pkg_tss.Point
  Signature github_com_bandprotocol_chain_v2_pkg_tss.Signature
  Status SigningStatus
  CreatedHeight uint64
  CreatedTimestamp time.Time
}
```

### DE

In generating partial signature, the `x/tss` module uses DE submitted from members as a nonce for generating group public nonce for forming a group signature. Members must maintain their DEs for being selected as a signer in signing process.

```go
type DE struct {
	PubD github_com_bandprotocol_chain_v2_pkg_tss.Point
	PubE github_com_bandprotocol_chain_v2_pkg_tss.Point
}
```

## State

### Group

The `x/tss` module stores group information and the number of group existing on chain.

- GroupCount: `0x00 -> BigEndian(#group)`. Store the number of group existing on chain.
- Group: `0x10 | GroupID -> Group`. Store the information of the group.

### Member

The `x/tss` module stores member information for checking their status during the group and signing request creation process; users can be in multiple groups.

- Member: `0x11 | GroupID | MemberID -> Member`. Store a member information of the specific group.

### Group Creation

During the group creation process, the `x/tss` module stores information for generating group public key and they will be removed after the group creation process is expired.

- PendingProcessGroups: `0x02 -> []GroupID`. Store the list of groupID whose status and information should be updated at the EndBlock.
- DKGContext: `0x12 | GroupID -> []byte`. Store a nonce that being used in a group creation process.
- Round1Info: `0x13 | GroupID | MemberID -> Round1Info`. Store an information that member submits during the 1st round group creation message.
- Round1InfoCount: `0x14 | GroupID -> BigEndian(#Round1Info)`. Store the number of round1 information message.
- AccumulatedCommit: `0x15 | GroupID | index -> []byte`. Store accumulated commit point for generating a group public key
- Round2Info: `0x16 | GroupID | MemberID -> Round2Info`. Store an information that member submits during the 1st round group creation message.
- Round2InfoCount: `0x17 | GroupID -> BigEndian(#Round2Info)`. Store the number of round2 information message.
- ComplaintWithStatus: `0x18 | GroupID | MemberID -> ComplaintWithStatus`. Store a complaint information that member submits during the 3rd round group creation message with its status.
- ConfirmComplaintCount: `0x19 | GroupID -> BigEndian(#Confirm + #Complaint)`. Store the number of round3 information message.
- Confirm: `0x1a | GroupID | MemberID -> Confirm`. Store a confirm information that member submits during the 3rd round group creation message.

### Signing

In signing process, the `x/tss` module stores partial signature submitted from assigned members and aggregates them once every member submits it. The aggregated signature (group signature) will be stored in the signing object and those partial signatures will be removed.

- SigningCount: `0x01 -> BigEndian(#signing)`. Store the number of signings existing on chain.
- PendingProcessSignings: `0x03 -> []SigningID`. Store the list of signingID whose status and information should be updated at the EndBlock.
- SigningExpirations: `0x05 -> []SigningExpiration`. Store the list of signing expiration information. The order of expiration time should be increasing (from beginning of the list to the end).
- Signing `0x1d | SigningID -> Signing`. Store the information of the signing request.
- PartialSignatureCount: `0x1e | SigningID | Attempt -> BigEndian(#PartialSigning)`. Store the number of partial signature of the given signing ID.
- PartialSignature: `0x1f | SigningID | Attempt | MemberID -> PartialSignature`. Store the partial signature of the member of the given signing ID.
- SigningAttempt: `0x20 | SigningID | Attempt -> SigningAttempt`. Store the signing attempt object of the given signing ID and specific attempt. The SigningAttempt object store assigned members and expiration height of that attempt.

### DE

In generating partial signature, the `x/tss` module uses DE submitted from members as a nonce for generating group public nonce for forming a group signature. Members must maintain their DEs for being selected as a signer in signing process.

- DE `0x1b | address | index -> DE`. Store the DE object
- DEQueue `0x1c | address -> DEQueue`. Store the DEQueue object to identify an index of valid DE.

### Params

The `x/tss` module stores its params in state with the prefix of `0x20`, it can be updated with governance proposal or the address with authority.

- Params: `0x90 -> Params`

The `x/tss` module contains the following parameters

```protobuf
message Params {
  // max_group_size is the maximum of the member capacity of the group.
  uint64 max_group_size = 1;
  // max_d_e_size is the maximum of the de capacity of the member.
  uint64 max_d_e_size = 2;
  // creation_period is the number of blocks allowed to creating tss group.
  uint64 creation_period = 3;
  // signing_period is the number of blocks allowed to sign.
  uint64 signing_period = 4;
  // max_signing_attempt is the maximum number of signing retry process per signingID.
  uint64 max_signing_attempt = 5;
  // max_memo_length is the maximum length of the memo in the direct originator.
  uint64 max_memo_length = 6;
  // max_message_length is the maximum length of the message in the TextSignatureOrder.
  uint64 max_message_length = 7;
}
```

## Msg Service

### Msg/SubmitDKGRound1

This message is used to send round 1 information in the DKG process of the group.

When a group is created, all members of the group are required to send this message to the chain. So, the chain can proceed to the next step of the DKG process.

```protobuf
message MsgSubmitDKGRound1 {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "tss/MsgSubmitDKGRound1";

  // group_id is ID of the group.
  uint64 group_id = 1
      [(gogoproto.customname) = "GroupID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.GroupID"];
  // round1_info is all data that require to handle round 1.
  Round1Info round1_info = 2 [(gogoproto.nullable) = false];
  // sender is the user address that submits the group creation information;
  // must be a member of this group.
  string sender = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/SubmitDKGRound2

This message is used to send round 2 information in the DKG process of the group.

When a group is passed round 1, all members of the group are required to send this message to the chain. So, the chain can proceed to the next step of the DKG process.

```protobuf
message MsgSubmitDKGRound2 {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "tss/MsgSubmitDKGRound2";

  // group_id is ID of the group.
  uint64 group_id = 1
      [(gogoproto.customname) = "GroupID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.GroupID"];
  // round2_info is all data that is required to handle round 2.
  Round2Info round2_info = 2 [(gogoproto.nullable) = false];
  // sender is the user address that submits the group creation information;
  // must be a member of this group.
  string sender = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/Complain

This message is used to complain to any malicious member of the group if their shared secret data doesn't align with public information.

A member can send this message when the group is in round 3 of the DKG process. If there is one valid `MsgComplain` in this round, the group creation process will fail and the malicious member will be punished.

```protobuf
message MsgComplain {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "tss/MsgComplaint";

  // group_id is ID of the group.
  uint64 group_id = 1
      [(gogoproto.customname) = "GroupID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.GroupID"];
  // complaints is a list of complaints.
  repeated Complaint complaints = 2 [(gogoproto.nullable) = false];
  // sender is the user address that submits the group creation information;
  // must be a member of this group.
  string sender = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/Confirm

This message is used to confirm that all information from other members is correct.

A member can send this message when the group is in round 3 of the DKG process. They are required to send `MsgConfirm` or `MsgComplain` in this process. Otherwise, they will be deactivated from the TSS system.

```protobuf
message MsgConfirm {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "tss/MsgConfirm";

  // group_id is ID of the group.
  uint64 group_id = 1
      [(gogoproto.customname) = "GroupID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.GroupID"];
  // member_id is ID of the sender.
  uint64 member_id = 2
      [(gogoproto.customname) = "MemberID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.MemberID"];
  // own_pub_key_sig is a signature of the member_i on its own PubKey to confirm
  // that the address is able to derive the PubKey.
  bytes own_pub_key_sig = 3 [(gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.Signature"];
  // sender is the user address that submits the group creation information;
  // must be a member of this group.
  string sender = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/SubmitDEs

In the signing process, each member is required to have their nonces (D and E values). `MsgSubmitDEs` is the message for a member to send their public nonce to the chain. So, the chain can assign their nonce in the signing process.

```protobuf
message MsgSubmitDEs {
  option (cosmos.msg.v1.signer) = "sender";
  option (amino.name)           = "tss/MsgSubmitDEs";

  // des is a list of DE objects.
  repeated DE des = 1 [(gogoproto.customname) = "DEs", (gogoproto.nullable) = false];
  // sender is the user address that submits DE objects.
  string sender = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

It's expected to fail if:

- The number of remaining DEs exceeds the maximum size (`MaxDESize`) per user.

### Msg/SubmitSignature

When a user requests a signature from the group, the assigned member of the group is required to send `MsgSubmitSignature` to the chain. It contains `signing_id`, `member_id`, `address`, and `signature`.

Once all assigned member sends their signature to the chain, the chain will aggregate those signatures to be the final signature of the group for that request.

```protobuf
message MsgSubmitSignature {
  option (cosmos.msg.v1.signer) = "signer";
  option (amino.name)           = "tss/MsgSubmitSignature";

  // signing_id is the unique identifier of the signing process.
  uint64 signing_id = 1 [
    (gogoproto.customname) = "SigningID",
    (gogoproto.casttype)   = "github.com/bandprotocol/chain/v3/pkg/tss.SigningID"
  ];
  // member_id is the unique identifier of the signer in the group.
  uint64 member_id = 2
      [(gogoproto.customname) = "MemberID", (gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.MemberID"];
  // signature is the signature produced by the signer.
  bytes signature = 3 [(gogoproto.casttype) = "github.com/bandprotocol/chain/v3/pkg/tss.Signature"];
  // signer is the address who signs a message; must be a member of the group.
  string signer = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### Msg/UpdateParams

When anyone wants to update the parameters of the TSS module, they will have to open a governance proposal by using the `MsgUpdateParams` of the TSS module to update those parameters.

```protobuf
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";
  option (amino.name)           = "tss/MsgUpdateParams";

  // params defines the x/tss parameters to update.
  Params params = 1 [(gogoproto.nullable) = false];
  // authority is the address of the governance account.
  string authority = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

##

## Callbacks

A module can register its callback object within the x/tss module. This callback will be triggered when the group associated with the module acts on a specific process. The callback must provide the following methods:

### OnGroupCreationCompleted

- Trigger-by: Endblock of the group creation process. It will be triggered when every members in the group successfully submit their confirm message.

### OnGroupCreationFailed

- Trigger-by: Endblock of the group creation process. It will be triggered when some members in the group submit complain message due to invalid encrypted secret shares.

### OnGroupCreationExpired

- Trigger-by: Endblock of the group-creation expiration process. It will be triggered when the process takes too long.

### OnSigningCompleted

- Trigger-by: Endblock of the signing process. It will be triggered after every members in the group submit their partial signature and their signatures are combined to form a group signature.

### OnSigningFailed

- Trigger-by: A retry signing process, It will be triggered when the signing process is retried over the limit, which is defined by the module's parameter.

### OnSigningTimeout

- Trigger-by: Endblock of the signing expiration process. It will be triggered when the process waits assign members to submit their partial signature for too long.

## Events

The TSS module emits the following events:

### EventTypeCreateGroup

This event ( `create_group` ) is emitted when the group is created.

| Attribute Key | Attribute Value   |
| ------------- | ----------------- |
| group_id      | {groupID}         |
| size          | {groupSize}       |
| threshold     | {groupThreshold}  |
| pub_key       | ""                |
| status        | {groupStatus}     |
| dkg_context   | {groupDKGContext} |
| module_owner  | {moduleName}      |

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
| attempt          | {signing.CurrentAttempt}       |
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

## Client

### CLI

A user can query and interact with the `TSS` module using the CLI.

#### Query

The `query` commands allow users to query the `group` state.

```bash
bandd query tss --help
```

##### Counts

The `Counts` command allows users to query the number of existing group and signing in the `x/tss` module

```bash
bandd query tss counts [flags]
```

##### DELists

The `DELists` command allows users to query the existing DE of the specific address on chain.

```bash
bandd query tss de-list [address]
```

Example:

```bash
bandd query tss de-list band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q
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

##### IsGrantee

The `IsGrantee` command allows users to query if a given address is a grantee of the specific granter

```bash
bandd query tss is-grantee [granter_address] [grantee_address] [flags]
```

Example:

```bash
bandd query tss de-list band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q band1w8yurh6naeqg4mjx4zcs7hsu3fppwu0f4q4l7f
```

##### Members

The `Members` command allows users to query the member information of the specific groupID

```bash
bandd query tss members [group-id] [flags]
```

Example:

```bash
bandd query tss members 1
```

##### Pending Signings

The `Pending Signings` command allows users to query list of signings that waiting the given address to be signed

```bash
bandd query tss pending-signings [address] [flags]
```

Example:

```bash
bandd queue tss pending-signings band1nx0xkpnytk35wvflsg7gszf9smw3vaeauk248q
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
-d '{"signing_id":"1"}' localhost:9090 tss.v1beta1.Query/Signing
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
curl localhost:1317/tss/v1beta1/signings/1
```
