package tss

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func ConcatBytes(data ...[]byte) []byte {
	var res []byte
	for _, b := range data {
		res = append(res, b...)
	}

	return res
}

func GenerateKeyPairs(n uint64) (KeyPairs, error) {
	var kps KeyPairs
	for i := uint64(0); i < n; i++ {
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		kps = append(kps, kp)
	}

	return kps, nil
}

func GenerateKeyPair() (KeyPair, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKey: key.Serialize(),
		PublicKey:  key.PubKey().SerializeCompressed(),
	}, nil
}

func GenerateKeySymIJ(rawPrivKeyI PrivateKey, rawPubKeyJ PublicKey) (PublicKey, error) {
	privKeyI := rawPrivKeyI.Parse()

	pubKeyJ, err := rawPubKeyJ.Point()
	if err != nil {
		return nil, err
	}

	var keySymIJ secp256k1.JacobianPoint
	secp256k1.ScalarMultNonConst(&privKeyI.Key, pubKeyJ, &keySymIJ)

	return ParsePublicKey(keySymIJ), nil
}
