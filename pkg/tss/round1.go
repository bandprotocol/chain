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
	gid GroupID,
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
	oneTimeSig, err := SignOneTime(gid, dkgContext, oneTimePubKey, oneTimePrivKey)
	if err != nil {
		return nil, err
	}

	// get a0 information
	a0PrivKey := kps[1].PrivateKey
	a0PubKey := kps[1].PublicKey
	a0Sig, err := SignA0(gid, dkgContext, a0PubKey, a0PrivKey)
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
	gid GroupID,
	dkgContext []byte,
	a0Pub PublicKey,
	a0Priv PrivateKey,
) ([]byte, error) {
	commitment := generateCommitmentA0(gid, dkgContext, a0Pub)
	return Sign(a0Priv, commitment, nil, nil)
}

func VerifyA0Sig(
	gid GroupID,
	dkgContext []byte,
	signature Signature,
	a0Pub PublicKey,
) (bool, error) {
	commitment := generateCommitmentA0(gid, dkgContext, a0Pub)
	return Verify(signature, commitment, a0Pub, nil)
}

func SignOneTime(
	gid GroupID,
	dkgContext []byte,
	oneTimePub PublicKey,
	onetimePriv PrivateKey,
) ([]byte, error) {
	commitment := generateCommitmentOneTime(gid, dkgContext, oneTimePub)
	return Sign(onetimePriv, commitment, nil, nil)
}

func VerifyOneTimeSig(
	gid GroupID,
	dkgContext []byte,
	signature Signature,
	oneTimePub PublicKey,
) (bool, error) {
	commitment := generateCommitmentOneTime(gid, dkgContext, oneTimePub)
	return Verify(signature, commitment, oneTimePub, nil)
}

func generateCommitmentA0(gid GroupID, dkgContext []byte, a0Pub PublicKey) []byte {
	return ConcatBytes([]byte("round1A0"), sdk.Uint64ToBigEndian(uint64(gid)), dkgContext, a0Pub)
}

func generateCommitmentOneTime(gid GroupID, dkgContext []byte, oneTimePub PublicKey) []byte {
	return ConcatBytes([]byte("round1OneTime"), sdk.Uint64ToBigEndian(uint64(gid)), dkgContext, oneTimePub)
}
