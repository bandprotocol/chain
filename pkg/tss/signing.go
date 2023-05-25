package tss

import (
	"github.com/bandprotocol/chain/v2/pkg/tss/internal/lagrange"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Note: Currently, support maximum N at 20
func ComputeLagrangeCoefficient(mid MemberID, n uint64) Scalar {
	coeff := lagrange.ComputeCoefficient(int64(mid), int64(n)).Bytes()
	scalarValue := new(secp256k1.ModNScalar)
	scalarValue.SetByteSlice(coeff)

	return ParseScalar(scalarValue)
}

func ComputeOwnLo(mid MemberID, msg []byte, bytes []byte) Scalar {
	bz := Hash([]byte("signingLo"), sdk.Uint64ToBigEndian(uint64(mid)), msg, bytes)

	var lo secp256k1.ModNScalar
	lo.SetByteSlice(bz)

	return ParseScalar(&lo)
}

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

func ComputeGroupPublicNonce(rawOwnPubNonces PublicKeys) (PublicKey, error) {
	points, err := rawOwnPubNonces.Points()
	if err != nil {
		return nil, err
	}

	return ParsePublicKey(sumPoints(points...)), nil
}

func SignSigning(
	rawGroupPubNonce PublicKey,
	rawGroupPubKey PublicKey,
	msg []byte,
	rawLagrange Scalar,
	ownPrivNonce PrivateKey,
	ownPrivKey PrivateKey,
) (Signature, error) {
	challenge := GenerateChallengeSigning(rawGroupPubNonce, rawGroupPubKey, msg)

	// TODO-TSS: use lagrange
	// TODO-TSS: remove inserting public nonce in to challenge
	return Sign(ownPrivKey, challenge, Scalar(ownPrivNonce))
}

func VerifySigning(
	rawGroupPubNonce PublicKey,
	rawGroupPubKey PublicKey,
	msg []byte,
	rawLagrange Scalar,
	rawSig Signature,
	ownPubKey PublicKey,
) error {
	challenge := GenerateChallengeSigning(rawGroupPubNonce, rawGroupPubKey, msg)

	// TODO-TSS: use lagrange
	// TODO-TSS: remove inserting public nonce in to challenge
	return Verify(rawSig, challenge, ownPubKey, nil, nil)
}

func VerifyGroupSigning(
	rawGroupPubKey PublicKey,
	msg []byte,
	rawSig Signature,
) error {
	challenge := GenerateChallengeGroupSigning(rawGroupPubKey, msg)
	return Verify(rawSig, challenge, rawGroupPubKey, nil, nil)
}

func GenerateChallengeSigning(rawGroupPubNonce PublicKey, rawGroupPubKey PublicKey, msg []byte) []byte {
	return ConcatBytes(rawGroupPubNonce, []byte("signing"), rawGroupPubKey, msg)
}

func GenerateChallengeGroupSigning(rawGroupPubKey PublicKey, msg []byte) []byte {
	return ConcatBytes([]byte("signing"), rawGroupPubKey, msg)
}
