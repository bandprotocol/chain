package tss

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func ConcatBytes(data ...[]byte) []byte {
	var res []byte
	for _, b := range data {
		res = append(res, b...)
	}
	return res
}

func GenerateKeyPairs(n uint64) (types.KeyPairs, error) {
	var kps types.KeyPairs
	for i := uint64(0); i < n; i++ {
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		kps = append(kps, kp)
	}

	return kps, nil
}

func GenerateKeyPair() (types.KeyPair, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return types.KeyPair{}, err
	}

	return types.KeyPair{
		PrivateKey: key.Serialize(),
		PublicKey:  key.PubKey().SerializeCompressed(),
	}, nil
}

func GenerateKeySymIJ(privKeyIBytes types.PrivateKey, pubKeyJBytes types.PublicKey) ([]byte, error) {
	privKeyI := secp256k1.PrivKeyFromBytes(privKeyIBytes)

	pubKeyJ, err := secp256k1.ParsePubKey(pubKeyJBytes)
	if err != nil {
		return nil, err
	}

	var pubKeyJPoint secp256k1.JacobianPoint
	pubKeyJ.AsJacobian(&pubKeyJPoint)

	var keySymIJ secp256k1.JacobianPoint
	secp256k1.ScalarMultNonConst(&privKeyI.Key, &pubKeyJPoint, &keySymIJ)
	keySymIJ.ToAffine()

	return secp256k1.NewPublicKey(&keySymIJ.X, &keySymIJ.Y).SerializeCompressed(), nil
}

func parseJacobianPoint(bytes []byte) (*secp256k1.JacobianPoint, error) {
	pk, err := secp256k1.ParsePubKey(bytes)
	if err != nil {
		return nil, err
	}

	var point secp256k1.JacobianPoint
	pk.AsJacobian(&point)

	return &point, nil
}
