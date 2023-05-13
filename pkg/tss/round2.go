package tss

import (
	"errors"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func ComputeEncryptedSecretShares(
	mid MemberID,
	rawPrivKey PrivateKey,
	rawPubKeys PublicKeys,
	rawCoeffcients Scalars,
) (Scalars, error) {
	// compute key sym for each person 1..n except mid
	var keySyms PublicKeys
	for i, rawPubKey := range rawPubKeys {
		idx := i + 1
		if MemberID(idx) == mid {
			continue
		}

		keySym, err := ComputeKeySym(rawPrivKey, rawPubKey)
		if err != nil {
			return nil, err
		}

		keySyms = append(keySyms, keySym)
	}

	// calculate secret share for each person 1..n except mid
	var secretShares Scalars
	for i := uint32(1); i <= uint32(len(rawPubKeys)); i++ {
		if MemberID(i) == mid {
			continue
		}

		secretShare := ComputeSecretShare(rawCoeffcients, i)
		secretShares = append(secretShares, secretShare)
	}

	// encrypt each secret share by its own key sym
	return EncryptSecretShares(secretShares, keySyms)
}

func EncryptSecretShares(
	secretShares Scalars,
	keySyms PublicKeys,
) (Scalars, error) {
	if len(secretShares) != len(keySyms) {
		return nil, errors.New("the length of secret shares and key syms is not equal")
	}

	var encSecretShares Scalars
	for i := 0; i < len(secretShares); i++ {
		encSecretShare := Encrypt(secretShares[i], keySyms[i])
		encSecretShares = append(encSecretShares, encSecretShare)
	}

	return encSecretShares, nil
}

func ComputeSecretShare(rawCoeffcients Scalars, rawX uint32) Scalar {
	x := new(secp256k1.ModNScalar).SetInt(rawX)
	result := solveScalarPolynomial(rawCoeffcients.Parse(), x)

	return ParseScalar(result)
}
