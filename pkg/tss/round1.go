package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Round1Data contains the data for round 1 of the DKG process of TSS
type Round1Data struct {
	OneTimePrivKey     PrivateKey
	OneTimePubKey      PublicKey
	OneTimeSig         Signature
	A0PrivKey          PrivateKey
	A0PubKey           PublicKey
	A0Sig              Signature
	Coefficients       Scalars
	CoefficientsCommit Points
}

// GenerateRound1Data generates the data of round1 for a member in the DKG process of TSS
func GenerateRound1Data(
	mid MemberID,
	threshold uint64,
	dkgContext []byte,
) (*Round1Data, error) {
	// Generate threshold + 1 key pairs (commitments, onetime).
	kps, err := GenerateKeyPairs(threshold + 1)
	if err != nil {
		return nil, err
	}

	// Get one-time information.
	oneTimePrivKey := kps[0].PrivateKey
	oneTimePubKey := kps[0].PublicKey
	oneTimeSig, err := SignOneTime(mid, dkgContext, oneTimePubKey, oneTimePrivKey)
	if err != nil {
		return nil, err
	}

	// Get a0 information.
	a0PrivKey := kps[1].PrivateKey
	a0PubKey := kps[1].PublicKey
	a0Sig, err := SignA0(mid, dkgContext, a0PubKey, a0PrivKey)
	if err != nil {
		return nil, err
	}

	// Get coefficients.
	var coefficientsCommit Points
	var coefficients Scalars
	for i := 1; i < len(kps); i++ {
		coefficientsCommit = append(coefficientsCommit, Point(kps[i].PublicKey))
		coefficients = append(coefficients, Scalar(kps[i].PrivateKey))
	}

	return &Round1Data{
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
	challenge := GenerateChallengeA0(mid, dkgContext, a0Pub)
	return Sign(a0Priv, challenge, nil)
}

// VerifyA0Sig verifies the signature for the A0 in round 1.
func VerifyA0Sig(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	a0Pub PublicKey,
) error {
	challenge := GenerateChallengeA0(mid, dkgContext, a0Pub)
	return Verify(signature, challenge, a0Pub, nil, nil)
}

// SignOneTime generates a signature for the one-time in round 1.
func SignOneTime(
	mid MemberID,
	dkgContext []byte,
	oneTimePub PublicKey,
	onetimePriv PrivateKey,
) (Signature, error) {
	challenge := GenerateChallengeOneTime(mid, dkgContext, oneTimePub)
	return Sign(onetimePriv, challenge, nil)
}

// VerifyOneTimeSig verifies the signature for the one-time in round 1.
func VerifyOneTimeSig(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	oneTimePub PublicKey,
) error {
	challenge := GenerateChallengeOneTime(mid, dkgContext, oneTimePub)
	return Verify(signature, challenge, oneTimePub, nil, nil)
}

// Generate the challenge for the A0 signature using the member ID, DKG context, and A0 public key.
func GenerateChallengeA0(mid MemberID, dkgContext []byte, a0Pub PublicKey) []byte {
	return ConcatBytes([]byte("round1A0"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, a0Pub)
}

// Generate the challenge for the one-time signature using the member ID, DKG context, and one-time public key.
func GenerateChallengeOneTime(mid MemberID, dkgContext []byte, oneTimePub PublicKey) []byte {
	return ConcatBytes([]byte("round1OneTime"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, oneTimePub)
}
