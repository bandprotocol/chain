package types

import (
	"time"

	"github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

// NewGroup creates a new Group instance.
func NewGroup(
	id tss.GroupID,
	size uint64,
	threshold uint64,
	pubKey tss.Point,
	status GroupStatus,
	createdHeight uint64,
	moduleOwner string,
) Group {
	return Group{
		ID:            id,
		Size_:         size,
		Threshold:     threshold,
		PubKey:        pubKey,
		Status:        status,
		CreatedHeight: createdHeight,
		ModuleOwner:   moduleOwner,
	}
}

// NewMember creates a new Member instance.
func NewMember(
	id tss.MemberID,
	groupID tss.GroupID,
	addr sdk.AccAddress,
	pubKey tss.Point,
	isMalicious bool,
	isActive bool,
) Member {
	return Member{
		ID:          id,
		GroupID:     groupID,
		Address:     addr.String(),
		PubKey:      pubKey,
		IsMalicious: isMalicious,
		IsActive:    isActive,
	}
}

// NewRound1Info creates a new Round1Info instance.
func NewRound1Info(
	memberID tss.MemberID,
	coefficientCommits tss.Points,
	oneTimePubKey tss.Point,
	a0Sig tss.Signature,
	oneTimeSig tss.Signature,
) Round1Info {
	return Round1Info{
		MemberID:           memberID,
		CoefficientCommits: coefficientCommits,
		OneTimePubKey:      oneTimePubKey,
		A0Signature:        a0Sig,
		OneTimeSignature:   oneTimeSig,
	}
}

// NewRound2Info creates a new Round2Info instance.
func NewRound2Info(
	memberID tss.MemberID,
	encryptedSecretShares tss.EncSecretShares,
) Round2Info {
	return Round2Info{
		MemberID:              memberID,
		EncryptedSecretShares: encryptedSecretShares,
	}
}

// NewConfirm creates a new Confirm instance.
func NewConfirm(memberID tss.MemberID, ownPubKeySig tss.Signature) Confirm {
	return Confirm{
		MemberID:     memberID,
		OwnPubKeySig: ownPubKeySig,
	}
}

// NewComplaint creates a new Complaint instance.
func NewComplaint(
	complainant tss.MemberID,
	respondent tss.MemberID,
	keySym tss.Point,
	complaintSig tss.ComplaintSignature,
) Complaint {
	return Complaint{
		Complainant: complainant,
		Respondent:  respondent,
		KeySym:      keySym,
		Signature:   complaintSig,
	}
}

// NewComplaintWithStatus creates a new Complaint instance with its status.
func NewComplaintWithStatus(complaint Complaint, status ComplaintStatus) ComplaintWithStatus {
	return ComplaintWithStatus{
		Complaint:       complaint,
		ComplaintStatus: status,
	}
}

// NewDE creates a new DE instance.
func NewDE(pubD tss.Point, pubE tss.Point) DE {
	return DE{
		PubD: pubD,
		PubE: pubE,
	}
}

// NewDEQueue creates a new DEQueue instance.
func NewDEQueue(head uint64, tail uint64) DEQueue {
	return DEQueue{
		Head: head,
		Tail: tail,
	}
}

// NewSigning creates a new Signing instance with provided parameters.
func NewSigning(
	id tss.SigningID,
	currentAttempt uint64,
	gid tss.GroupID,
	groupPubKey tss.Point,
	originatorBz []byte,
	msg []byte,
	groupPubNonce tss.Point,
	signature tss.Signature,
	status SigningStatus,
	createdHeight uint64,
	createdTimestamp time.Time,
) Signing {
	return Signing{
		ID:               id,
		CurrentAttempt:   currentAttempt,
		GroupID:          gid,
		GroupPubKey:      groupPubKey,
		Originator:       originatorBz,
		Message:          msg,
		GroupPubNonce:    groupPubNonce,
		Signature:        signature,
		Status:           status,
		CreatedHeight:    createdHeight,
		CreatedTimestamp: createdTimestamp,
	}
}

// NewSigningAttempt creates a new signingAttempt instance.
func NewSigningAttempt(
	signingID tss.SigningID,
	attempt uint64,
	expiredHeight uint64,
	assignedMembers []AssignedMember,
) SigningAttempt {
	return SigningAttempt{
		SigningID:       signingID,
		Attempt:         attempt,
		ExpiredHeight:   expiredHeight,
		AssignedMembers: assignedMembers,
	}
}

// NewAssignedMember creates a new AssignedMember instance.
func NewAssignedMember(
	member Member,
	de DE,
	bindingFactor tss.Scalar,
	pubNonce tss.Point,
) AssignedMember {
	return AssignedMember{
		MemberID:      member.ID,
		Address:       member.Address,
		PubKey:        member.PubKey,
		PubD:          de.PubD,
		PubE:          de.PubE,
		BindingFactor: bindingFactor,
		PubNonce:      pubNonce,
	}
}

// NewSigningExpiration creates a new SigningExpiration instance.
func NewSigningExpiration(signingID tss.SigningID, attempt uint64) SigningExpiration {
	return SigningExpiration{
		SigningID:      signingID,
		SigningAttempt: attempt,
	}
}

// NewPendingProcessGroups creates a new PendingProcessGroups instance.
func NewPendingProcessGroups(groupIDs []tss.GroupID) PendingProcessGroups {
	return PendingProcessGroups{GroupIDs: groupIDs}
}

// NewPendingProcessSignings creates a new PendingProcessSignings instance.
func NewPendingProcessSignings(signingIDs []tss.SigningID) PendingProcessSignings {
	return PendingProcessSignings{SigningIDs: signingIDs}
}

// NewSigningExpirations creates a new SigningExpirations instance.
func NewSigningExpirations(signingExpirations []SigningExpiration) SigningExpirations {
	return SigningExpirations{SigningExpirations: signingExpirations}
}

// NewPartialSignature creates a new PartialSignature instance.
func NewPartialSignature(
	signingID tss.SigningID,
	attempt uint64,
	memberID tss.MemberID,
	signature tss.Signature,
) PartialSignature {
	return PartialSignature{
		SigningID:      signingID,
		SigningAttempt: attempt,
		MemberID:       memberID,
		Signature:      signature,
	}
}

// NewEVMSignature creates a new EVMSignature instance.
func NewEVMSignature(signature tss.Signature) (EVMSignature, error) {
	rAddr, err := signature.R().Address()
	if err != nil {
		return EVMSignature{}, err
	}

	return EVMSignature{
		RAddress:  rAddr,
		Signature: bytes.HexBytes(signature.S()),
	}, nil
}

// NewSigningResult creates a new SigningResult instance.
func NewSigningResult(
	signing Signing,
	signingAttempt *SigningAttempt,
	evmSignature *EVMSignature,
	receivedPartialSigs []PartialSignature,
) SigningResult {
	return SigningResult{
		Signing:                   signing,
		CurrentSigningAttempt:     signingAttempt,
		EVMSignature:              evmSignature,
		ReceivedPartialSignatures: receivedPartialSigs,
	}
}
