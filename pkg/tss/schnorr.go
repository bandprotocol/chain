package tss

import (
	"errors"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/schnorr"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func Sign(
	privKeyBytes types.PrivateKey,
	commitment []byte,
	generatorBytes *[]byte,
	nonceBytes *[]byte,
) (types.Signature, error) {
	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	privKeyScalar := &privKey.Key

	var generator *secp256k1.JacobianPoint
	if generatorBytes != nil {
		var err error
		generator, err = parseJacobianPoint(*generatorBytes)
		if err != nil {
			return nil, err
		}
	}

	for iterator := uint64(0); ; iterator++ {
		var nonce secp256k1.ModNScalar
		if nonceBytes == nil {
			nonce = *secp256k1.NonceRFC6979(privKeyBytes, commitment, schnorr.RFC6979ExtraDataV0[:], nil, uint32(iterator))
		} else {
			overflow := nonce.SetByteSlice(*nonceBytes)
			if overflow {
				return nil, errors.New("nonce is overflow")
			}
		}

		sig, err := schnorr.Sign(privKeyScalar, &nonce, commitment, generator)
		nonce.Zero()
		if err != nil {
			if nonceBytes == nil {
				continue
			}
			return nil, err
		}

		return sig.Serialize(), nil
	}
}

func Verify(
	signatureBytes types.Signature,
	commitment []byte,
	pubKeyBytes types.PublicKey,
	generatorBytes *[]byte,
) (bool, error) {
	sig, err := schnorr.ParseSignature(signatureBytes)
	if err != nil {
		return false, err
	}

	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false, err
	}

	var generator *secp256k1.JacobianPoint
	if generatorBytes != nil {
		generator, err = parseJacobianPoint(*generatorBytes)
		if err != nil {
			return false, err
		}
	}

	err = schnorr.Verify(sig, commitment, pubKey, generator)
	return err == nil, nil
}
