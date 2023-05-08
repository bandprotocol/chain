package tss

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Round1Data struct {
	OneTimePrivKey     types.PrivateKey
	OneTimePubKey      types.PublicKey
	OneTimeSig         types.Signature
	A0PrivKey          types.PrivateKey
	A0PubKey           types.PublicKey
	A0Sig              types.Signature
	Coefficients       types.Coefficients
	CoefficientsCommit types.Points
}

func GenerateRound1Data(
	gid types.GroupID,
	mid types.MemberID,
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
	oneTimeCommitment := Hash1(sdk.Uint64ToBigEndian(uint64(gid)), dkgContext, oneTimePubKey)
	oneTimeSig, err := Sign(oneTimePrivKey, oneTimeCommitment)
	if err != nil {
		return nil, err
	}

	// get a0 information
	a0PrivKey := kps[1].PrivateKey
	a0PubKey := kps[1].PublicKey
	a0Commitment := Hash2(sdk.Uint64ToBigEndian(uint64(gid)), dkgContext, a0PubKey)
	a0Sig, err := Sign(a0PrivKey, a0Commitment)
	if err != nil {
		return nil, err
	}

	// get coeffcients
	var coefficientsCommit types.Points
	var coefficients types.Coefficients
	for i := 1; i < len(kps); i++ {
		coefficientsCommit = append(coefficientsCommit, types.Point(kps[i].PublicKey))
		coefficients = append(coefficients, types.Coefficient(kps[i].PrivateKey))
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

func VerifyOneTimeSig(
	gid types.GroupID,
	dkgContext []byte,
	signature types.Signature,
	pubKey types.PublicKey,
) (bool, error) {
	commitment := Hash1(sdk.Uint64ToBigEndian(uint64(gid)), dkgContext, pubKey)
	return Verify(signature, commitment, pubKey)
}

func VerifyA0Sig(
	gid types.GroupID,
	dkgContext []byte,
	signature types.Signature,
	pubKey types.PublicKey,
) (bool, error) {
	commitment := Hash2(sdk.Uint64ToBigEndian(uint64(gid)), dkgContext, pubKey)
	return Verify(signature, commitment, pubKey)
}
