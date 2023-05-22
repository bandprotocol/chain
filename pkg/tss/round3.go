package tss

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeOwnPublicKey computes the own public key for a given set of sum commits and x.
// The formula used is: Yi = Σ(k=0 to t-1) (i^k * Σ(j=1 to n) (Commit_jk))
func ComputeOwnPublicKey(rawSumCommits Points, rawX uint32) (Point, error) {
	sumCommits, err := rawSumCommits.Parse()
	if err != nil {
		return nil, err
	}

	x := new(secp256k1.ModNScalar).SetInt(rawX)
	result := solvePointPolynomial(sumCommits, x)

	return ParsePoint(result), nil
}

// ComputeGroupPublicKey computes the group public key from a set of A0 commits.
// The formula used is: Y = Σ(i=1 to n) (Commit_j0)
// TODO: Remove this function after the chain itself move to use accumulated commits instead
func ComputeGroupPublicKey(rawA0Commits Points) (PublicKey, error) {
	a0Commits, err := rawA0Commits.Parse()
	if err != nil {
		return nil, err
	}

	pubKey := sumPoints(a0Commits...)
	return ParsePublicKey(pubKey), nil
}

// ComputeOwnPrivateKey computes the own private key from a set of secret shares.
// The formula used is: si = Σ(j=1 to n) (f_j(i))
func ComputeOwnPrivateKey(rawSecretShares Scalars) PrivateKey {
	privKey := sumScalars(rawSecretShares.Parse()...)
	return ParsePrivateKey(privKey)
}

// VerifySecretShare verifies the validity of a secret share for a given member.
// It compares the computed yG from the secret share with the yG computed from the commits.
func VerifySecretShare(mid MemberID, rawSecretShare Scalar, rawCommits Points) error {
	// Compute yG from the secret share.
	yG := new(secp256k1.JacobianPoint)
	secp256k1.ScalarBaseMultNonConst(rawSecretShare.Parse(), yG)

	// Compute yG from the commits.
	ssc, err := ComputeSecretShareCommit(rawCommits, uint32(mid))
	if err != nil {
		return err
	}

	// Compare the two yG values to check validity.
	if !bytes.Equal(ssc, ParsePoint(yG)) {
		return errors.New("invalid secret share")
	}

	return nil
}

// ComputeSecretShareCommit computes the secret share commit for a given set of commits and x.
// The formula used is: y * G = f_ij(x) * G = c_0 + c_1 * x^1 + ... + c_n-1 * x^(n-1) + c_n * x^n
// rawCommits represents the commits c_0, c_1, ..., c_n-1, c_n = a_0 * G, a_1 * G, ..., a_n-1 * G, a_n * G
// rawX represents x, the index of the shared secret commit.
func ComputeSecretShareCommit(rawCommits Points, rawX uint32) (Point, error) {
	commits, err := rawCommits.Parse()
	if err != nil {
		return nil, err
	}

	x := new(secp256k1.ModNScalar).SetInt(rawX)
	result := solvePointPolynomial(commits, x)

	return ParsePoint(result), nil
}

// DecryptSecretShares decrypts a set of encrypted secret shares using the corresponding key syms.
func DecryptSecretShares(
	encSecretShares Scalars,
	keySyms PublicKeys,
) (Scalars, error) {
	if len(encSecretShares) != len(keySyms) {
		return nil, errors.New("the length of encrypted secret shares and key syms is not equal")
	}

	var secretShares Scalars
	for i := 0; i < len(encSecretShares); i++ {
		secretShare := Decrypt(encSecretShares[i], keySyms[i])
		secretShares = append(secretShares, secretShare)
	}

	return secretShares, nil
}

// SignOwnPublicKey signs the own public key using the given DKG context, own public key, and own private key.
func SignOwnPublickey(
	mid MemberID,
	dkgContext []byte,
	ownPub PublicKey,
	ownPriv PrivateKey,
) (Signature, error) {
	challenge := GenerateChallengeOwnPublicKey(mid, dkgContext, ownPub)
	return Sign(ownPriv, challenge, nil)
}

// SignComplain generates a signature and related parameters for complaining against a misbehaving member.
func SignComplain(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	oneTimePrivI PrivateKey,
) (Signature, PublicKey, PublicKey, error) {
	keySym, err := ComputeKeySym(oneTimePrivI, oneTimePubJ)
	if err != nil {
		return nil, nil, nil, err
	}

	for iterator := uint32(0); ; iterator++ {
		nonce := GenerateNonce(
			oneTimePrivI,
			Hash(oneTimePubI, oneTimePubJ, keySym),
			iterator,
		)

		nonceSym, err := ComputeNonceSym(nonce, oneTimePubJ)
		if err != nil {
			return nil, nil, nil, err
		}

		challenge := GenerateChallengeComplain(oneTimePubI, oneTimePubJ, keySym, nonceSym)

		sig, err := Sign(oneTimePrivI, challenge, nonce)
		if err != nil {
			continue
		}

		return sig, keySym, nonceSym, nil
	}
}

// VerifyOwnPubKeySig verifies the signature of an own public key using the given DKG context, own public key, and signature.
func VerifyOwnPubKeySig(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	ownPub PublicKey,
) error {
	challenge := GenerateChallengeOwnPublicKey(mid, dkgContext, ownPub)
	return Verify(signature, challenge, ownPub, nil, nil)
}

// VerifyComplainSig verifies the signature of a complaint using the given parameters.
func VerifyComplainSig(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	keySym PublicKey,
	nonceSym PublicKey,
	signature Signature,
) error {
	challenge := GenerateChallengeComplain(oneTimePubI, oneTimePubJ, keySym, nonceSym)
	err := Verify(signature, challenge, oneTimePubI, nil, nil)
	if err != nil {
		return err
	}

	err = Verify(signature, challenge, keySym, Point(oneTimePubJ), nonceSym)
	if err != nil {
		return err
	}

	return nil
}

// generateChallengeOwnPublicKey generates the challenge for verifying an own public key signature.
func GenerateChallengeOwnPublicKey(mid MemberID, dkgContext []byte, ownPub PublicKey) []byte {
	return ConcatBytes([]byte("round3OwnPubKey"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, ownPub)
}

// generateChallengeComplain generates the challenge for verifying a complaint signature.
func GenerateChallengeComplain(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	keySym PublicKey,
	nonceSym PublicKey,
) []byte {
	return ConcatBytes([]byte("round3Complain"), oneTimePubJ, oneTimePubJ, keySym, nonceSym)
}
