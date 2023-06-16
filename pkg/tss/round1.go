package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// round1Info contains the data for round 1 of the DKG process of TSS
type round1Info struct {
	OneTimePrivKey     PrivateKey
	OneTimePubKey      PublicKey
	OneTimeSig         Signature
	A0PrivKey          PrivateKey
	A0PubKey           PublicKey
	A0Sig              Signature
	Coefficients       Scalars
	CoefficientsCommit Points
}

// Generateround1Info generates the data of round 1 for a member in the DKG process of TSS
func Generateround1Info(
	mid MemberID,
	threshold uint64,
	dkgContext []byte,
) (*round1Info, error) {
	// Generate threshold + 1 key pairs (commits, onetime).
	kps, err := GenerateKeyPairs(threshold + 1)
	if err != nil {
		return nil, NewError(err, "generate key pairs")
	}

	// Get one-time information.
	oneTimePrivKey := kps[0].PrivKey
	oneTimePubKey := kps[0].PubKey
	oneTimeSig, err := SignOneTime(mid, dkgContext, oneTimePubKey, oneTimePrivKey)
	if err != nil {
		return nil, NewError(err, "sign one time")
	}

	// Get a0 information.
	a0PrivKey := kps[1].PrivKey
	a0PubKey := kps[1].PubKey
	a0Sig, err := SignA0(mid, dkgContext, a0PubKey, a0PrivKey)
	if err != nil {
		return nil, NewError(err, "sign A0")
	}

	// Get coefficients.
	var coefficientsCommit Points
	var coefficients Scalars
	for i := 1; i < len(kps); i++ {
		coefficientsCommit = append(coefficientsCommit, Point(kps[i].PubKey))
		coefficients = append(coefficients, Scalar(kps[i].PrivKey))
	}

	return &round1Info{
		OneTimePrivKey:     oneTimePrivKey,
		OneTimePubKey:      oneTimePubKey,
		OneTimeSig:         oneTimeSig,
		A0PrivKey:          a0PrivKey,
		A0PubKey:           a0PubKey,
		A0Sig:              a0Sig,
		Coefficients:       coefficients,
		CoefficientsCommit: coefficientsCommit,
	}, nil
}

// SignA0 generates a signature for the A0 in round 1.
func SignA0(
	mid MemberID,
	dkgContext []byte,
	a0Pub PublicKey,
	a0Priv PrivateKey,
) (Signature, error) {
	msg := generateMessageA0(mid, dkgContext, a0Pub)
	nonce, pubNonce := GenerateNonce(a0Priv, Hash(msg))
	return Sign(a0Priv, ConcatBytes(pubNonce, msg), nonce, nil)
}

// VerifyA0Sig verifies the signature for the A0 in round 1.
func VerifyA0Sig(
	mid MemberID,
	dkgContext []byte,
	sig Signature,
	a0Pub PublicKey,
) error {
	msg := ConcatBytes(sig.R(), generateMessageA0(mid, dkgContext, a0Pub))
	return Verify(sig.R(), sig.S(), msg, a0Pub, nil, nil)
}

// generateMessageA0 generates the message for the A0 signature.
func generateMessageA0(mid MemberID, dkgContext []byte, a0Pub PublicKey) []byte {
	return ConcatBytes([]byte("round1A0"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, a0Pub)
}

// SignOneTime generates a signature for the one-time in round 1.
func SignOneTime(
	mid MemberID,
	dkgContext []byte,
	oneTimePub PublicKey,
	onetimePriv PrivateKey,
) (Signature, error) {
	msg := generateMessageOneTime(mid, dkgContext, oneTimePub)
	nonce, pubNonce := GenerateNonce(onetimePriv, Hash(msg))
	return Sign(onetimePriv, ConcatBytes(pubNonce, msg), nonce, nil)
}

// VerifyOneTimeSig verifies the signature for the one-time in round 1.
func VerifyOneTimeSig(
	mid MemberID,
	dkgContext []byte,
	sig Signature,
	oneTimePub PublicKey,
) error {
	msg := ConcatBytes(sig.R(), generateMessageOneTime(mid, dkgContext, oneTimePub))
	return Verify(sig.R(), sig.S(), msg, oneTimePub, nil, nil)
}

// generateMessageOneTime generates the message for the one-time signature.
func generateMessageOneTime(mid MemberID, dkgContext []byte, oneTimePub PublicKey) []byte {
	return ConcatBytes([]byte("round1OneTime"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, oneTimePub)
}
