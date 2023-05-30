package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/lagrange"
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeLagrangeCoefficient calculates the Lagrange coefficient for a given member ID and total number of members.
// Note: Currently, supports a maximum mid at 20.
func ComputeLagrangeCoefficient(mid MemberID, memberList []MemberID) Scalar {
	var mids []int64
	for _, member := range memberList {
		mids = append(mids, int64(member))
	}

	coeff := lagrange.ComputeCoefficient(int64(mid), mids).Bytes()

	scalarValue := new(secp256k1.ModNScalar)
	scalarValue.SetByteSlice(coeff)

	return ParseScalar(scalarValue)
}

// ComputeOwnLo calculates the own Lo value for a given member ID, data, and bytes.
// Lo = Hash(i, data , B)
// B = <<i,Di,Ei>,...>
func ComputeOwnLo(mid MemberID, data []byte, bytes []byte) Scalar {
	bz := Hash([]byte("signingLo"), sdk.Uint64ToBigEndian(uint64(mid)), Hash(data), Hash(bytes))

	var lo secp256k1.ModNScalar
	lo.SetByteSlice(bz)

	return ParseScalar(&lo)
}

// ComputeOwnPublicNonce calculates the own public nonce for a given public D, public E, and Lo.
// Formula: D + Lo * E
func ComputeOwnPublicNonce(rawPubD PublicKey, rawPubE PublicKey, rawLo Scalar) (PublicKey, error) {
	lo, err := rawLo.Parse()
	if err != nil {
		return nil, err
	}

	pubD, err := rawPubD.Point()
	if err != nil {
		return nil, err
	}

	pubE, err := rawPubE.Point()
	if err != nil {
		return nil, err
	}

	var loE secp256k1.JacobianPoint
	secp256k1.ScalarMultNonConst(lo, pubE, &loE)

	var ownPubNonce secp256k1.JacobianPoint
	secp256k1.AddNonConst(pubD, &loE, &ownPubNonce)

	return ParsePublicKey(&ownPubNonce), nil
}

// ComputeOwnPrivateNonce calculates the own private nonce for a given private d, private e, and Lo.
// Formula: d + Lo * e
func ComputeOwnPrivateNonce(rawPrivD PrivateKey, rawPrivE PrivateKey, rawLo Scalar) (PrivateKey, error) {
	lo, err := rawLo.Parse()
	if err != nil {
		return nil, err
	}

	privD, err := rawPrivD.Scalar()
	if err != nil {
		return nil, err
	}

	privE, err := rawPrivE.Scalar()
	if err != nil {
		return nil, err
	}

	lo.Mul(privE)
	privD.Add(lo)

	return ParsePrivateKey(privD), nil
}

// ComputeGroupPublicNonce calculates the group public nonce for a given slice of own public nonces.
// Formula: Sum(PubNonce1, PubNonce2, ..., PubNonceN)
func ComputeGroupPublicNonce(rawOwnPubNonces ...PublicKey) (PublicKey, error) {
	pubNonces, err := PublicKeys(rawOwnPubNonces).Points()
	if err != nil {
		return nil, err
	}

	return ParsePublicKey(sumPoints(pubNonces...)), nil
}

// CombineSignatures performs combining all signatures by sum up R and sum up S.
func CombineSignatures(rawSigs ...Signature) (Signature, error) {
	var allR []*secp256k1.JacobianPoint
	var allS []*secp256k1.ModNScalar
	for _, rawSig := range rawSigs {
		sig, err := rawSig.Parse()
		if err != nil {
			return nil, err
		}

		allR = append(allR, &sig.R)
		allS = append(allS, &sig.S)
	}

	return ParseSignature(schnorr.NewSignature(sumPoints(allR...), sumScalars(allS...))), nil
}

// SignSigning performs signing using the group public nonce, group public key, data, Lagrange coefficient,
// own private nonce, and own private key.
func SignSigning(
	groupPubNonce PublicKey,
	groupPubKey PublicKey,
	data []byte,
	rawLagrange Scalar,
	ownPrivNonce PrivateKey,
	ownPrivKey PrivateKey,
) (Signature, error) {
	msg := ConcatBytes(groupPubNonce, GenerateMessageGroupSigning(groupPubKey, data))
	return Sign(ownPrivKey, msg, Scalar(ownPrivNonce), rawLagrange)
}

// VerifySigning verifies the signing using the group public nonce, group public key, data, Lagrange coefficient,
// signature, and own public key.
func VerifySigningSig(
	groupPubNonce PublicKey,
	groupPubKey PublicKey,
	data []byte,
	rawLagrange Scalar,
	sig Signature,
	ownPubKey PublicKey,
) error {
	msg := ConcatBytes(groupPubNonce, GenerateMessageGroupSigning(groupPubKey, data))
	return Verify(sig.R(), sig.S(), msg, ownPubKey, nil, rawLagrange)
}

// VerifyGroupSigning verifies the group signing using the group public key, data, and signature.
func VerifyGroupSigningSig(
	groupPubKey PublicKey,
	data []byte,
	sig Signature,
) error {
	msg := ConcatBytes(sig.R(), GenerateMessageGroupSigning(groupPubKey, data))
	return Verify(sig.R(), sig.S(), msg, groupPubKey, nil, nil)
}

// GenerateMessageGroupSigning generates the message for group signing using the group public key and data.
func GenerateMessageGroupSigning(rawGroupPubKey PublicKey, data []byte) []byte {
	return ConcatBytes([]byte("signing"), rawGroupPubKey, data)
}
