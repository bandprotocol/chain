package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func GenerateRound1Data(
	mid MemberID,
	threshold uint64,
	dkgContext []byte,
) (*Round1Data, error) {
	// generate threshold + 1 key pairs (commiments, onetime)
	kps, err := GenerateKeyPairs(threshold + 1)
	if err != nil {
		return nil, err
	}

	// get one time information
	oneTimePrivKey := kps[0].PrivateKey
	oneTimePubKey := kps[0].PublicKey
	oneTimeSig, err := SignOneTime(mid, dkgContext, oneTimePubKey, oneTimePrivKey)
	if err != nil {
		return nil, err
	}

	// get a0 information
	a0PrivKey := kps[1].PrivateKey
	a0PubKey := kps[1].PublicKey
	a0Sig, err := SignA0(mid, dkgContext, a0PubKey, a0PrivKey)
	if err != nil {
		return nil, err
	}

	// get coeffcients
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

func SignA0(
	mid MemberID,
	dkgContext []byte,
	a0Pub PublicKey,
	a0Priv PrivateKey,
) (Signature, error) {
	challenge := generateChallengeA0(mid, dkgContext, a0Pub)
	return Sign(a0Priv, challenge, nil, nil)
}

func VerifyA0Sig(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	a0Pub PublicKey,
) error {
	challenge := generateChallengeA0(mid, dkgContext, a0Pub)
	return Verify(signature, challenge, a0Pub, nil)
}

func SignOneTime(
	mid MemberID,
	dkgContext []byte,
	oneTimePub PublicKey,
	onetimePriv PrivateKey,
) (Signature, error) {
	challenge := generateChallengeOneTime(mid, dkgContext, oneTimePub)
	return Sign(onetimePriv, challenge, nil, nil)
}

func VerifyOneTimeSig(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	oneTimePub PublicKey,
) error {
	challenge := generateChallengeOneTime(mid, dkgContext, oneTimePub)
	return Verify(signature, challenge, oneTimePub, nil)
}

func generateChallengeA0(mid MemberID, dkgContext []byte, a0Pub PublicKey) []byte {
	return ConcatBytes([]byte("round1A0"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, a0Pub)
}

func generateChallengeOneTime(mid MemberID, dkgContext []byte, oneTimePub PublicKey) []byte {
	return ConcatBytes([]byte("round1OneTime"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, oneTimePub)
}
