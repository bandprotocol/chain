package testutil

import "github.com/bandprotocol/chain/v2/pkg/tss"

type AssignedMember struct {
	ID            tss.MemberID
	PrivD         tss.Scalar
	PrivE         tss.Scalar
	BindingFactor tss.Scalar
	PrivNonce     tss.Scalar
	Lagrange      tss.Scalar
	Signature     tss.Signature
}

func (am AssignedMember) PubD() tss.Point {
	return Point(am.PrivD)
}

func (am AssignedMember) PubE() tss.Point {
	return Point(am.PrivE)
}

func (am AssignedMember) PubNonce() tss.Point {
	return Point(am.PrivNonce)
}

func CopyAssignedMember(src AssignedMember) AssignedMember {
	return AssignedMember{
		ID:            src.ID,
		PrivD:         Copy(src.PrivD),
		PrivE:         Copy(src.PrivE),
		BindingFactor: Copy(src.BindingFactor),
		PrivNonce:     Copy(src.PrivNonce),
		Lagrange:      Copy(src.Lagrange),
		Signature:     Copy(src.Signature),
	}
}

func CopyAssignedMembers(src []AssignedMember) []AssignedMember {
	var dst []AssignedMember
	for _, m := range src {
		dst = append(dst, CopyAssignedMember(m))
	}

	return dst
}

type Signing struct {
	ID              tss.SigningID
	Data            []byte
	Commitment      []byte
	PubNonce        tss.Point
	Signature       tss.Signature
	AssignedMembers []AssignedMember
}

func (s Signing) GetAllIDs() (res []tss.MemberID) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, assignedMember.ID)
	}

	return res
}

func (s Signing) GetAllPubDs() (res []tss.Point) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, Point(assignedMember.PrivD))
	}

	return res
}

func (s Signing) GetAllPubEs() (res []tss.Point) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, Point(assignedMember.PrivE))
	}

	return res
}

func (s Signing) GetAllOwnPubNonces() (res []tss.Point) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, Point(assignedMember.PrivNonce))
	}

	return res
}

func (s Signing) GetAllSigs() (res []tss.Signature) {
	for _, assignedMember := range s.AssignedMembers {
		res = append(res, assignedMember.Signature)
	}

	return res
}

func CopySigning(src Signing) Signing {
	return Signing{
		ID:              src.ID,
		Data:            Copy(src.Data),
		Commitment:      Copy(src.Commitment),
		PubNonce:        Copy(src.PubNonce),
		Signature:       Copy(src.Signature),
		AssignedMembers: CopyAssignedMembers(src.AssignedMembers),
	}
}

func CopySignings(src []Signing) []Signing {
	var dst []Signing
	for _, m := range src {
		dst = append(dst, CopySigning(m))
	}

	return dst
}
