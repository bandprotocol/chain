package testutil

import "github.com/bandprotocol/chain/v2/pkg/tss"

type AssignedMember struct {
	ID            tss.MemberID
	PrivD         tss.PrivateKey
	PrivE         tss.PrivateKey
	BindingFactor tss.Scalar
	PrivNonce     tss.PrivateKey
	Lagrange      tss.Scalar
	Sig           tss.Signature
}

func (am AssignedMember) PubD() tss.PublicKey {
	return PublicKey(am.PrivD)
}

func (am AssignedMember) PubE() tss.PublicKey {
	return PublicKey(am.PrivE)
}

func (am AssignedMember) PubNonce() tss.PublicKey {
	return PublicKey(am.PrivNonce)
}

type Signing struct {
	ID              tss.SigningID
	Data            []byte
	Commitment      []byte
	PubNonce        tss.PublicKey
	Sig             tss.Signature
	AssignedMembers []AssignedMember
}

func (s Signing) GetAllIDs() (res []tss.MemberID) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, assignedMember.ID)
	}

	return res
}

func (s Signing) GetAllPubDs() (res []tss.PublicKey) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, PublicKey(assignedMember.PrivD))
	}

	return res
}

func (s Signing) GetAllPubEs() (res []tss.PublicKey) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, PublicKey(assignedMember.PrivE))
	}

	return res
}

func (s Signing) GetAllOwnPubNonces() (res []tss.PublicKey) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, PublicKey(assignedMember.PrivNonce))
	}

	return res
}

func (s Signing) GetAllSigs() (res []tss.Signature) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, assignedMember.Sig)
	}

	return res
}
