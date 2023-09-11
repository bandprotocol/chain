package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/lagrange"
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeLagrangeCoefficientOp calculates the Lagrange coefficient with optimization for a given member ID and total number of members.
// Note: Currently, supports a maximum mid at 20.
func ComputeLagrangeCoefficientOp(mid MemberID, memberList []MemberID) Scalar {
	if len(memberList) > 20 {
		panic("compute lagrange coefficient optimization supports a maximum mid at 20.")
	}
	var mids []int64
	for _, member := range memberList {
		mids = append(mids, int64(member))
	}

	coeff := lagrange.ComputeCoefficientPreCompute(int64(mid), mids).Bytes()

	scalarValue := new(secp256k1.ModNScalar)
	scalarValue.SetByteSlice(coeff)

	return NewScalarFromModNScalar(scalarValue)
}

// ComputeLagrangeCoefficient calculates the Lagrange coefficient for a given member ID and total number of members.
func ComputeLagrangeCoefficient(mid MemberID, memberList []MemberID) Scalar {
	var mids []int64
	for _, member := range memberList {
		mids = append(mids, int64(member))
	}

	coeff := lagrange.ComputeCoefficient(int64(mid), mids).Bytes()

	scalarValue := new(secp256k1.ModNScalar)
	scalarValue.SetByteSlice(coeff)

	return NewScalarFromModNScalar(scalarValue)
}

// ComputeCommitment calculates the bytes that consists of memberID, public D, and public E.
func ComputeCommitment(mids []MemberID, pubDs Points, pubEs Points) ([]byte, error) {
	if len(mids) != len(pubDs) {
		return nil, NewError(ErrInvalidLength, "len(mids) != len(pubDs): %d != %d", len(mids), len(pubDs))
	}

	if len(mids) != len(pubEs) {
		return nil, NewError(ErrInvalidLength, "len(mids) != len(pubEs): %d != %d", len(mids), len(pubEs))
	}

	var commitment []byte
	prevMid := MemberIDZero()
	for i, mid := range mids {
		if prevMid >= mid {
			return nil, NewError(ErrInvalidOrder, "prevMid >= mid: %d != %d", prevMid, mid)
		}
		commitment = append(commitment, sdk.Uint64ToBigEndian(uint64(mid))...)
		commitment = append(commitment, pubDs[i]...)
		commitment = append(commitment, pubEs[i]...)
	}

	return commitment, nil
}

// ComputeOwnBindingFactor calculates the own binding factor (Lo) value for a given member ID, data, and commitment.
// bindingFactor = HashBindingFactor(i, data , B)
// B = <<i,Di,Ei>,...>
func ComputeOwnBindingFactor(mid MemberID, data []byte, commitment []byte) (Scalar, error) {
	scalar, err := HashBindingFactor(mid, data, commitment)
	if err != nil {
		return nil, err
	}

	return scalar, nil
}

// ComputeOwnPubNonce calculates the own public nonce for a given public D, public E, and binding factor.
// Formula: D + bindingFactor * E
func ComputeOwnPubNonce(rawPubD Point, rawPubE Point, rawBindingFactor Scalar) (Point, error) {
	bindingFactor := rawBindingFactor.modNScalar()

	pubD, err := rawPubD.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse public D")
	}

	pubE, err := rawPubE.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse public E")
	}

	var mulE secp256k1.JacobianPoint
	secp256k1.ScalarMultNonConst(bindingFactor, pubE, &mulE)

	var ownPubNonce secp256k1.JacobianPoint
	secp256k1.AddNonConst(pubD, &mulE, &ownPubNonce)

	return NewPointFromJacobianPoint(&ownPubNonce), nil
}

// ComputeOwnPrivNonce calculates the own private nonce for a given private d, private e, and binding factor.
// Formula: d + bindingFactor * e
func ComputeOwnPrivNonce(rawPrivD Scalar, rawPrivE Scalar, rawBindingFactor Scalar) (Scalar, error) {
	bindingFactor := rawBindingFactor.modNScalar()
	privD := rawPrivD.modNScalar()
	privE := rawPrivE.modNScalar()

	bindingFactor.Mul(privE)
	privD.Add(bindingFactor)

	return NewScalarFromModNScalar(privD), nil
}

// ComputeGroupPublicNonce calculates the group public nonce for a given slice of own public nonces.
// Formula: Sum(PubNonce1, PubNonce2, ..., PubNonceN)
func ComputeGroupPublicNonce(rawOwnPubNonces ...Point) (Point, error) {
	pubNonces, err := Points(rawOwnPubNonces).jacobianPoints()
	if err != nil {
		return nil, NewError(err, "parse own public nonces")
	}

	return NewPointFromJacobianPoint(sumPoints(pubNonces...)), nil
}

// CombineSignatures performs combining all signatures by sum up R and sum up S.
func CombineSignatures(rawSignatures ...Signature) (Signature, error) {
	var allR []*secp256k1.JacobianPoint
	var allS []*secp256k1.ModNScalar
	for idx, rawSignature := range rawSignatures {
		signature, err := rawSignature.signature()
		if err != nil {
			return nil, NewError(err, "parse signature: index: %d", idx)
		}

		allR = append(allR, &signature.R)
		allS = append(allS, &signature.S)
	}

	return NewSignatureFromType(schnorr.NewSignature(sumPoints(allR...), sumScalars(allS...))), nil
}

// SignSigning performs signing using the group public nonce, group public key, data, Lagrange coefficient,
// own private nonce, and own private key.
func SignSigning(
	groupPubNonce Point,
	groupPubKey Point,
	data []byte,
	rawLagrange Scalar,
	ownPrivNonce Scalar,
	ownPrivKey Scalar,
) (Signature, error) {
	challenge, err := HashChallenge(groupPubNonce, groupPubKey, data)
	if err != nil {
		return nil, err
	}

	return Sign(ownPrivKey, challenge, Scalar(ownPrivNonce), rawLagrange)
}

// VerifySigning verifies the signing using the group public nonce, group public key, data, Lagrange coefficient,
// signature, and own public key.
func VerifySigningSignature(
	groupPubNonce Point,
	groupPubKey Point,
	data []byte,
	rawLagrange Scalar,
	signature Signature,
	ownPubKey Point,
) error {
	challenge, err := HashChallenge(groupPubNonce, groupPubKey, data)
	if err != nil {
		return err
	}

	return Verify(signature.R(), signature.S(), challenge, ownPubKey, nil, rawLagrange)
}

// VerifyGroupSigning verifies the group signing using the group public key, data, and signature.
func VerifyGroupSigningSignature(
	groupPubKey Point,
	data []byte,
	signature Signature,
) error {
	challenge, err := HashChallenge(signature.R(), groupPubKey, data)
	if err != nil {
		return err
	}

	return Verify(signature.R(), signature.S(), challenge, groupPubKey, nil, nil)
}
