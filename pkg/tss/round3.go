package tss

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeOwnPublicKey computes the own public key for a given set of sum commits and x.
// The formula used is: Yi = Σ(k=0 to t-1) (i^k * Σ(j=1 to n) (Commit_jk))
func ComputeOwnPublicKey(rawSumCommits Points, mid MemberID) (PublicKey, error) {
	sumCommits, err := rawSumCommits.Parse()
	if err != nil {
		return nil, NewError(err, "parse sum commits")
	}

	x := new(secp256k1.ModNScalar).SetInt(uint32(mid))
	result := solvePointPolynomial(sumCommits, x)

	return ParsePublicKeyFromPoint(result), nil
}

// ComputeGroupPublicKey computes the group public key from a set of A0 commits.
// The formula used is: Y = Σ(i=1 to n) (Commit_j0)
func ComputeGroupPublicKey(rawA0Commits ...Point) (PublicKey, error) {
	a0Commits, err := Points(rawA0Commits).Parse()
	if err != nil {
		return nil, NewError(err, "parse a0 commits")
	}

	pubKey := sumPoints(a0Commits...)
	return ParsePublicKeyFromPoint(pubKey), nil
}

// ComputeOwnPrivateKey computes the own private key from a set of secret shares.
// The formula used is: si = Σ(j=1 to n) (f_j(i))
func ComputeOwnPrivateKey(rawSecretShares ...Scalar) (PrivateKey, error) {
	secretShares, err := Scalars(rawSecretShares).Parse()
	if err != nil {
		return nil, NewError(err, "parse secret shares")
	}

	privKey := sumScalars(secretShares...)

	return ParsePrivateKeyFromScalar(privKey), nil
}

// VerifySecretShare verifies the validity of a secret share for a given member.
// It compares the computed yG from the secret share with the yG computed from the commits.
func VerifySecretShare(mid MemberID, rawSecretShare Scalar, rawCommits Points) error {
	// Compute yG from the secret share.
	yG := new(secp256k1.JacobianPoint)
	secretShare, err := rawSecretShare.Parse()
	if err != nil {
		return NewError(err, "parse secret share")
	}
	secp256k1.ScalarBaseMultNonConst(secretShare, yG)

	// Compute yG from the commits.
	ssc, err := ComputeSecretShareCommit(rawCommits, mid)
	if err != nil {
		return NewError(err, "compute secret share commit")
	}

	// Compare the two yG values to check validity.
	if !bytes.Equal(ssc, ParsePoint(yG)) {
		return ErrInvalidSecretShare
	}

	return nil
}

// ComputeSecretShareCommit computes the secret share commit for a given set of commits and x.
// The formula used is: y * G = f_ij(x) * G = c_0 + c_1 * x^1 + ... + c_n-1 * x^(n-1) + c_n * x^n
// rawCommits represents the commits c_0, c_1, ..., c_n-1, c_n = a_0 * G, a_1 * G, ..., a_n-1 * G, a_n * G
// rawX represents x, the index of the shared secret commit.
func ComputeSecretShareCommit(rawCommits Points, mid MemberID) (Point, error) {
	commits, err := rawCommits.Parse()
	if err != nil {
		return nil, NewError(err, "parse commits")
	}

	x := new(secp256k1.ModNScalar).SetInt(uint32(mid))
	result := solvePointPolynomial(commits, x)

	return ParsePoint(result), nil
}

// DecryptSecretShares decrypts a set of encrypted secret shares using the corresponding key syms.
func DecryptSecretShares(
	encSecretShares Scalars,
	keySyms PublicKeys,
) (Scalars, error) {
	if len(encSecretShares) != len(keySyms) {
		return nil, NewError(
			ErrInvalidLength,
			"len(encrypted secret shares) != len(key sym): %d != %d",
			len(encSecretShares),
			len(keySyms),
		)
	}

	var secretShares Scalars
	for i := 0; i < len(encSecretShares); i++ {
		secretShare, err := DecryptSecretShare(encSecretShares[i], keySyms[i])
		if err != nil {
			return nil, NewError(err, "decrypt secret share")
		}

		secretShares = append(secretShares, secretShare)
	}

	return secretShares, nil
}

// DecryptSecretShare decrypts a encrypted secret share using the key sym.
func DecryptSecretShare(
	encSecretShare Scalar,
	keySym PublicKey,
) (Scalar, error) {
	return Decrypt(encSecretShare, keySym)
}

// SignOwnPubkey signs the own public key using the given DKG context, own public key, and own private key.
func SignOwnPubkey(
	mid MemberID,
	dkgContext []byte,
	ownPub PublicKey,
	ownPriv PrivateKey,
) (Signature, error) {
	msg := generateMessageOwnPublicKey(mid, dkgContext, ownPub)
	nonce, pubNonce := GenerateNonce(ownPriv, Hash(msg))
	return Sign(ownPriv, ConcatBytes(pubNonce, msg), nonce, nil)
}

// VerifyOwnPubKeySig verifies the signature of an own public key using the given DKG context, own public key, and signature.
func VerifyOwnPubKeySig(
	mid MemberID,
	dkgContext []byte,
	sig Signature,
	ownPub PublicKey,
) error {
	msg := ConcatBytes(sig.R(), generateMessageOwnPublicKey(mid, dkgContext, ownPub))
	return Verify(sig.R(), sig.S(), msg, ownPub, nil, nil)
}

// generateMessageOwnPublicKey generates the message for verifying an own public key signature.
func generateMessageOwnPublicKey(mid MemberID, dkgContext []byte, ownPub PublicKey) []byte {
	return ConcatBytes([]byte("round3OwnPubKey"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, ownPub)
}

// SignComplaint generates a signature and related parameters for complaining against a misbehaving member.
func SignComplaint(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	oneTimePrivI PrivateKey,
) (ComplaintSignature, PublicKey, error) {
	keySym, err := ComputeKeySym(oneTimePrivI, oneTimePubJ)
	if err != nil {
		return nil, nil, NewError(err, "compute key sym")
	}

	msg := generateMessageComplaint(oneTimePubI, oneTimePubJ, keySym)
	nonce, pubNonce := GenerateNonce(oneTimePrivI, Hash(msg))

	nonceSym, err := ComputeNonceSym(nonce, oneTimePubJ)
	if err != nil {
		return nil, nil, NewError(err, "compute nonce sym")
	}

	sig, err := Sign(oneTimePrivI, ConcatBytes(pubNonce, nonceSym, msg), nonce, nil)
	if err != nil {
		return nil, nil, NewError(err, "sign")
	}

	complaintSig, err := NewComplaintSignature(sig.R(), Point(nonceSym), sig.S())
	if err != nil {
		return nil, nil, NewError(err, "create complaint signature")
	}

	return complaintSig, keySym, nil
}

// VerifyComplaint verifies the complaint using the complaint signature and encrypted secret share.
func VerifyComplaint(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	keySym PublicKey,
	complaintSig ComplaintSignature,
	encSecretShare Scalar,
	midI MemberID,
	commits Points,
) error {
	err := VerifyComplaintSig(oneTimePubI, oneTimePubJ, keySym, complaintSig)
	if err != nil {
		return NewError(err, "verify complaint signature")
	}

	secretShare, err := DecryptSecretShare(encSecretShare, keySym)
	if err != nil {
		return NewError(err, "decrypt secret share")
	}

	err = VerifySecretShare(midI, secretShare, commits)
	if err == nil {
		return ErrValidSecretShare
	}

	return nil
}

// VerifyComplaintSig verifies the signature of a complaint using the given parameters.
func VerifyComplaintSig(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	keySym PublicKey,
	complaintSig ComplaintSignature,
) error {
	msg := ConcatBytes(
		complaintSig.A1(),
		complaintSig.A2(),
		generateMessageComplaint(oneTimePubI, oneTimePubJ, keySym),
	)

	err := Verify(complaintSig.A1(), complaintSig.Z(), msg, oneTimePubI, nil, nil)
	if err != nil {
		return NewError(err, "verify")
	}

	return Verify(complaintSig.A2(), complaintSig.Z(), msg, keySym, Point(oneTimePubJ), nil)
}

// generateMessageComplaint generates the message for verifying a complaint signature.
func generateMessageComplaint(oneTimePubI PublicKey, oneTimePubJ PublicKey, keySym PublicKey) []byte {
	return ConcatBytes([]byte("round3Complain"), oneTimePubJ, oneTimePubJ, keySym)
}
