package tss

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeEncryptedSecretShares computes the encrypted secret shares for a member.
func ComputeEncryptedSecretShares(
	mid MemberID,
	rawPrivKey Scalar,
	rawPubKeys Points,
	rawCoeffcients Scalars,
	n16g INonce16Generator,
) (EncSecretShares, error) {
	// Compute the key sym for each member 1..n except mid.
	var keySyms Points
	for i, rawPubKey := range rawPubKeys {
		idx := i + 1
		if NewMemberID(idx) == mid {
			continue
		}

		keySym, err := ComputeSecretSym(rawPrivKey, rawPubKey)
		if err != nil {
			return nil, NewError(err, "compute key sym")
		}

		keySyms = append(keySyms, keySym)
	}

	// Calculate the secret share for each member 1..n except mid.
	var secretShares Scalars
	for i := uint32(1); i <= uint32(len(rawPubKeys)); i++ {
		if NewMemberID(i) == mid {
			continue
		}

		secretShare, err := ComputeSecretShare(rawCoeffcients, NewMemberID(i))
		if err != nil {
			return nil, NewError(err, "compute secret share: member id: %d", i)
		}

		secretShares = append(secretShares, secretShare)
	}

	// Encrypt each secret share using its corresponding key sym.
	return EncryptSecretShares(secretShares, keySyms, n16g)
}

// EncryptSecretShares encrypts secret shares using key syms.
func EncryptSecretShares(
	secretShares Scalars,
	keySyms Points,
	n16g INonce16Generator,
) (EncSecretShares, error) {
	if len(secretShares) != len(keySyms) {
		return nil, NewError(
			ErrInvalidLength,
			"len(secret shares) != len(key sym): %d != %d",
			len(secretShares),
			len(keySyms),
		)
	}

	var encSecretShares EncSecretShares
	for i := 0; i < len(secretShares); i++ {
		enc, err := Encrypt(secretShares[i], keySyms[i], n16g)
		if err != nil {
			return nil, NewError(err, "compute secret share: member id: %d", i)
		}

		encSecretShares = append(encSecretShares, enc)
	}

	return encSecretShares, nil
}

// ComputeSecretShare computes the secret share for a given set of coefficients and x.
func ComputeSecretShare(rawCoeffcients Scalars, mid MemberID) (Scalar, error) {
	x := new(secp256k1.ModNScalar).SetInt(uint32(mid))

	coeffcients := rawCoeffcients.modNScalars()
	result := solveScalarPolynomial(coeffcients, x)

	return NewScalarFromModNScalar(result), nil
}
