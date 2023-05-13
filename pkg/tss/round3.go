package tss

import (
	"bytes"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Yi = Sigma_k=0_t-1 (i^k * (Sigma_j=1_n (Commit_jk)))
// Yi = Sigma_k=0_t-1 (i^k * SubCommit_k)
func ComputeOwnPublicKey(rawSumCommits Points, rawX uint32) (Point, error) {
	sumCommits, err := rawSumCommits.Parse()
	if err != nil {
		return nil, err
	}

	x := new(secp256k1.ModNScalar).SetInt(rawX)
	result := solvePointEquation(sumCommits, x)

	return ParsePoint(result), nil
}

// TODO: Remove this function. Use precompute accumulated commits during each message to optimize the cost instead
// Y = Sigma_i=1_n (Commit_j0)
func ComputeGroupPublicKey(rawA0Commits Points) (PublicKey, error) {
	a0Commits, err := rawA0Commits.Parse()
	if err != nil {
		return nil, err
	}

	pubKey := sumPoints(a0Commits...)
	return ParsePublicKey(pubKey), nil
}

// si = Sigma_j=1_n (f_j(i))
func ComputeOwnPrivateKey(rawSecretShares Scalars) PrivateKey {
	privKey := sumScalars(rawSecretShares.Parse()...)
	return ParsePrivateKey(privKey)
}

func VerifySecretShare(mid MemberID, rawSecretShare Scalar, rawCommits Points) error {
	// compute yG from secert share
	yG := new(secp256k1.JacobianPoint)
	secp256k1.ScalarBaseMultNonConst(rawSecretShare.Parse(), yG)

	// compute yG from commits
	ssc, err := ComputeSecretShareCommit(rawCommits, uint32(mid))
	if err != nil {
		return err
	}

	// compare two YG to check validity
	if !bytes.Equal(ssc, ParsePoint(yG)) {
		return errors.New("invalid secret share")
	}

	return nil
}

// y * G = f_ij(x) * G = c_0 + c_1 * x^1 + ... + c_n-1 * x^(n-1) + c_n * x^n
// rawCommits = c_0, c_1, ..., c_n-1, c_n = a_0 * G, a_1 * G, ..., a_n-1 * G, a_n * G
// rawX = x -> index of shared secret commit
func ComputeSecretShareCommit(rawCommits Points, rawX uint32) (Point, error) {
	commits, err := rawCommits.Parse()
	if err != nil {
		return nil, err
	}

	x := new(secp256k1.ModNScalar).SetInt(rawX)
	result := solvePointEquation(commits, x)

	return ParsePoint(result), nil
}

func DecryptSecretShares(
	encSecretShares Scalars,
	keySyms PublicKeys,
) (Scalars, error) {
	if len(encSecretShares) != len(keySyms) {
		return nil, errors.New("the length of encrypted secret shares and key syms is not equal")
	}

	var secretShares Scalars
	for i := 0; i < len(secretShares); i++ {
		secretShare := Decrypt(secretShares[i], keySyms[i])
		secretShares = append(secretShares, secretShare)
	}

	return secretShares, nil
}

func SignOwnPublickey(
	mid MemberID,
	dkgContext []byte,
	ownPub PublicKey,
	ownPriv PrivateKey,
) (Signature, error) {
	challenge := generateChallengeOwnPublicKey(mid, dkgContext, ownPub)
	return Sign(ownPriv, challenge, nil)
}

func SignComplain(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	oneTimePrivI PrivateKey,
) (Signature, PublicKey, PublicKey, error) {
	keySym, err := GenerateKeySym(oneTimePrivI, oneTimePubJ)
	if err != nil {
		return nil, nil, nil, err
	}

	for iterator := uint32(0); ; iterator++ {
		nonce := GenerateNonce(
			oneTimePrivI,
			Hash(oneTimePubI, oneTimePubJ, keySym),
			iterator,
		)

		nonceSym, err := GenerateNonceSym(nonce, oneTimePubJ)
		if err != nil {
			return nil, nil, nil, err
		}

		challenge := generateChallengeComplain(oneTimePubI, oneTimePubJ, keySym, nonceSym)

		sig, err := Sign(oneTimePrivI, challenge, nonce)
		if err != nil {
			continue
		}

		return sig, keySym, nonceSym, nil
	}
}

func VerifyOwnPubKeySig(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	ownPub PublicKey,
) error {
	challenge := generateChallengeOwnPublicKey(mid, dkgContext, ownPub)
	return Verify(signature, challenge, ownPub, nil, nil)
}

func VerifyComplainSig(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	keySym PublicKey,
	nonceSym PublicKey,
	signature Signature,
) error {
	challenge := generateChallengeComplain(oneTimePubI, oneTimePubJ, keySym, nonceSym)
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

func generateChallengeOwnPublicKey(mid MemberID, dkgContext []byte, ownPub PublicKey) []byte {
	return ConcatBytes([]byte("round3OwnPubKey"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, ownPub)
}

func generateChallengeComplain(
	oneTimePubI PublicKey,
	oneTimePubJ PublicKey,
	keySym PublicKey,
	nonceSym PublicKey,
) []byte {
	return ConcatBytes([]byte("round3Complain"), oneTimePubJ, oneTimePubJ, keySym, nonceSym)
}
