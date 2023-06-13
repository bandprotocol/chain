package tss

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeEncryptedSecretShares computes the encrypted secret shares for a member.
func ComputeEncryptedSecretShares(
	mid MemberID,
	rawPrivKey PrivateKey,
	rawPubKeys PublicKeys,
	rawCoeffcients Scalars,
) (Scalars, error) {
	// Compute the key sym for each member 1..n except mid.
	var keySyms PublicKeys
	for i, rawPubKey := range rawPubKeys {
		idx := i + 1
		if MemberID(idx) == mid {
			continue
		}

		keySym, err := ComputeKeySym(rawPrivKey, rawPubKey)
		if err != nil {
			return nil, NewError(err, "compute key sym")
		}

		keySyms = append(keySyms, keySym)
	}

	// Calculate the secret share for each member 1..n except mid.
	var secretShares Scalars
	for i := uint32(1); i <= uint32(len(rawPubKeys)); i++ {
		if MemberID(i) == mid {
			continue
		}

		secretShare, err := ComputeSecretShare(rawCoeffcients, MemberID(i))
		if err != nil {
			return nil, NewError(err, "compute secret share: member id: %d", i)
		}

		secretShares = append(secretShares, secretShare)
	}

	// Encrypt each secret share using its corresponding key sym.
	return EncryptSecretShares(secretShares, keySyms)
}

// EncryptSecretShares encrypts secret shares using key syms.
func EncryptSecretShares(
	secretShares Scalars,
	keySyms PublicKeys,
) (Scalars, error) {
	if len(secretShares) != len(keySyms) {
		return nil, NewError(
			ErrInvalidLength,
			"len(secret shares) != len(key sym): %d != %d",
			len(secretShares),
			len(keySyms),
		)
	}

	var encSecretShares Scalars
	for i := 0; i < len(secretShares); i++ {
		encSecretShare, err := Encrypt(secretShares[i], keySyms[i])
		if err != nil {
			return nil, NewError(err, "compute secret share: member id: %d", i)
		}

		encSecretShares = append(encSecretShares, encSecretShare)
	}

	return encSecretShares, nil
}

// ComputeSecretShare computes the secret share for a given set of coefficients and x.
func ComputeSecretShare(rawCoeffcients Scalars, mid MemberID) (Scalar, error) {
	x := new(secp256k1.ModNScalar).SetInt(uint32(mid))

	coeffcients, err := rawCoeffcients.Parse()
	if err != nil {
		return nil, NewError(err, "parse coefficients")
	}

	result := solveScalarPolynomial(coeffcients, x)
	return ParseScalar(result), nil
}
