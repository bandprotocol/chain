package types

import (
	"time"

	"github.com/cometbft/cometbft/libs/bytes"

	"github.com/bandprotocol/chain/v3/pkg/tss"
)

// ====================================
// Group
// ====================================

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

// Validate performs basic validation of group information.
func (g Group) Validate() error {
	// validate group id
	if g.ID == 0 {
		return ErrInvalidGroup.Wrap("group id is 0")
	}

	// validate group size
	if g.Threshold > g.Size_ {
		return ErrInvalidGroup.Wrapf("group threshold is more than the group size")
	}

	// validate group public key
	if g.PubKey == nil {
		return ErrInvalidGroup.Wrap("group public key must not be nil")
	}
	if err := g.PubKey.Validate(); err != nil {
		return ErrInvalidGroup.Wrapf("group public key is invalid: %v", err)
	}

	// validate group status
	if _, ok := GroupStatus_name[int32(g.Status)]; !ok {
		return ErrInvalidGroup.Wrapf("invalid group status: %d", g.Status)
	}
	if g.Status == GROUP_STATUS_UNSPECIFIED {
		return ErrInvalidGroup.Wrap("group status is unspecified")
	}

	// validate created height
	if g.CreatedHeight == 0 {
		return ErrInvalidGroup.Wrap("created height is 0")
	}

	// validate module owner
	if g.ModuleOwner == "" {
		return ErrInvalidGroup.Wrap("module owner is empty")
	}

	return nil
}

// ====================================
// Round1Info
// ====================================

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

// Validate performs basic validation of round-1 group creation information.
func (r Round1Info) Validate() error {
	// Validate member ID
	if r.MemberID == 0 {
		return ErrInvalidMember.Wrap("member id is 0")
	}

	// Validate coefficients commit
	for _, c := range r.CoefficientCommits {
		if err := c.Validate(); err != nil {
			return ErrInvalidCoefficientCommit.Wrapf("invalid coefficient commit: %v", err)
		}
	}

	// Validate one time pub key
	if err := r.OneTimePubKey.Validate(); err != nil {
		return ErrInvalidPublicKey.Wrapf("invalid one-time public key: %v", err)
	}

	// Validate a0 signature
	if err := r.A0Signature.Validate(); err != nil {
		return ErrInvalidSignature.Wrapf("invalid a0 signature: %v", err)
	}

	// Validate one time signature
	if err := r.OneTimeSignature.Validate(); err != nil {
		return ErrInvalidSignature.Wrapf("invalid one-time signature: %v", err)
	}

	return nil
}

// ====================================
// Round2Info
// ====================================

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

// Validate performs basic validation of round-2 group creation information.
func (r Round2Info) Validate() error {
	if r.MemberID == 0 {
		return ErrInvalidMember.Wrap("member id is 0")
	}

	for i, ess := range r.EncryptedSecretShares {
		if err := ess.Validate(); err != nil {
			return ErrInvalidSecretShare.Wrapf("encrypted secret shares at index %d: %v", i, err)
		}
	}

	return nil
}

// ====================================
// Confirm
// ====================================

// NewConfirm creates a new Confirm instance.
func NewConfirm(memberID tss.MemberID, ownPubKeySig tss.Signature) Confirm {
	return Confirm{
		MemberID:     memberID,
		OwnPubKeySig: ownPubKeySig,
	}
}

// ====================================
// Complaint
// ====================================

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

// Validate performs basic validation of complaint information.
func (c Complaint) Validate() error {
	if c.Complainant == 0 {
		return ErrInvalidComplaint.Wrap("complainant is 0")
	}

	if c.Respondent == 0 {
		return ErrInvalidComplaint.Wrap("respondent is 0")
	}

	if c.Complainant == c.Respondent {
		return ErrInvalidComplaint.Wrap("complainant and respondent are the same")
	}

	if err := c.KeySym.Validate(); err != nil {
		return ErrInvalidComplaint.Wrapf("invalid symmetric key: %v", err)
	}

	if err := c.Signature.Validate(); err != nil {
		return ErrInvalidComplaint.Wrapf("invalid signature: %v", err)
	}

	return nil
}

// ====================================
// ComplaintWithStatus
// ====================================

// NewComplaintWithStatus creates a new Complaint instance with its status.
func NewComplaintWithStatus(complaint Complaint, status ComplaintStatus) ComplaintWithStatus {
	return ComplaintWithStatus{
		Complaint:       complaint,
		ComplaintStatus: status,
	}
}

// ====================================
// DE
// ====================================

// NewDE creates a new DE instance.
func NewDE(pubD tss.Point, pubE tss.Point) DE {
	return DE{
		PubD: pubD,
		PubE: pubE,
	}
}

// Validate performs basic validation of de information.
func (d DE) Validate() error {
	if err := d.PubD.Validate(); err != nil {
		return ErrInvalidDE.Wrap("invalid pub d")
	}

	if err := d.PubE.Validate(); err != nil {
		return ErrInvalidDE.Wrap("invalid pub e")
	}

	return nil
}

// ====================================
// DE Queue
// ====================================

// NewDEQueue creates a new DEQueue instance.
func NewDEQueue(head uint64, tail uint64) DEQueue {
	return DEQueue{
		Head: head,
		Tail: tail,
	}
}

// ====================================
// Signing
// ====================================

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

// ====================================
// SigningAttempt
// ====================================

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

// ====================================
// AssignedMember
// ====================================

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

// ====================================
// SigningExpiration
// ====================================

// NewSigningExpiration creates a new SigningExpiration instance.
func NewSigningExpiration(signingID tss.SigningID, attempt uint64) SigningExpiration {
	return SigningExpiration{
		SigningID:      signingID,
		SigningAttempt: attempt,
	}
}

// ====================================
// PendingProcessGroups
// ====================================

// NewPendingProcessGroups creates a new PendingProcessGroups instance.
func NewPendingProcessGroups(groupIDs []tss.GroupID) PendingProcessGroups {
	return PendingProcessGroups{GroupIDs: groupIDs}
}

// ====================================
// PendingProcessSignings
// ====================================

// NewPendingProcessSignings creates a new PendingProcessSignings instance.
func NewPendingProcessSignings(signingIDs []tss.SigningID) PendingProcessSignings {
	return PendingProcessSignings{SigningIDs: signingIDs}
}

// ====================================
// SigningExpirations
// ====================================

// NewSigningExpirations creates a new SigningExpirations instance.
func NewSigningExpirations(signingExpirations []SigningExpiration) SigningExpirations {
	return SigningExpirations{SigningExpirations: signingExpirations}
}

// ====================================
// PartialSignature
// ====================================

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

// ====================================
// EVMSignature
// ====================================

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

// ====================================
// SigningResult
// ====================================

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
