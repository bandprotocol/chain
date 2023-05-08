package tss

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/schnorr"
)

// TODO-CYLINDER: USING KECCAK INSTEAD OF BLAKE FOR HASH INTERNALLY
// NOTE: r,s in SIGNATURE are private variables.
func Sign(privKey types.PrivateKey, hash []byte) (types.Signature, error) {
	pk := secp256k1.PrivKeyFromBytes(privKey)
	sig, err := schnorr.Sign(pk, hash)
	return sig.Serialize(), err
}

func Verify(signature types.Signature, hash []byte, pubKey types.PublicKey) (bool, error) {
	sig, err := schnorr.ParseSignature(signature)
	if err != nil {
		return false, err
	}

	pk, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		return false, err
	}

	return sig.Verify(hash, pk), nil
}
