package tss

import (
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ContextString = "BAND-TSS-secp256k1-v0"
)

// Hash calculates the Keccak-256 hash of the given data.
// It returns the hash value as a byte slice.
func Hash(data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}

// HashRound1A0 computes a hash of the provided data for Round1A0 and returns it as a scalar (H1(m)).
func HashRound1A0(pubNonce Point, mid MemberID, dkgContext []byte, a0Pub Point) (Scalar, error) {
	scalar, err := NewScalar(
		Hash(
			[]byte(ContextString),
			[]byte("round1A0"),
			pubNonce,
			sdk.Uint64ToBigEndian(uint64(mid)),
			dkgContext,
			a0Pub,
		),
	)
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash round1A0")
	}

	return scalar, nil
}

// HashRound1OneTime computes a hash of the provided data for Round1OneTime and returns it as a scalar (H2(m)).
func HashRound1OneTime(pubNonce Point, mid MemberID, dkgContext []byte, oneTimePub Point) (Scalar, error) {
	scalar, err := NewScalar(
		Hash(
			[]byte(ContextString),
			[]byte("round1OneTime"),
			pubNonce,
			sdk.Uint64ToBigEndian(uint64(mid)),
			dkgContext,
			oneTimePub,
		),
	)
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash round1OneTime")
	}

	return scalar, nil
}

// HashRound3Complain computes a hash of the provided data for Round3Complain and returns it as a scalar (H3(m)).
func HashRound3Complain(
	pubNonce Point,
	nonceSym Point,
	oneTimePubI Point,
	oneTimePubJ Point,
	keySym Point,
) (Scalar, error) {
	scalar, err := NewScalar(
		Hash(
			[]byte(ContextString),
			[]byte("round3Complain"),
			pubNonce,
			nonceSym,
			oneTimePubI,
			oneTimePubJ,
			keySym,
		),
	)
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash round3Complain")
	}

	return scalar, nil
}

// HashRound3OwnPubKey computes a hash of the provided data for Round3OwnPubKey and returns it as a scalar (H4(m)).
func HashRound3OwnPubKey(pubNonce Point, mid MemberID, dkgContext []byte, ownPub Point) (Scalar, error) {
	scalar, err := NewScalar(
		Hash(
			[]byte(ContextString),
			[]byte("round3OwnPubKey"),
			pubNonce,
			sdk.Uint64ToBigEndian(uint64(mid)),
			dkgContext,
			ownPub,
		),
	)
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash round3OwnPubKey")
	}

	return scalar, nil
}

// HashSignMsg computes a hash of the message for signing purposes and returns the hash as a byte slice (H5(m)).
func HashSignMsg(data []byte) []byte {
	return Hash([]byte(ContextString), []byte("signMsg"), data)
}

// HashSignCommitment computes a hash of commitment and returns the hash as a byte slice (H6(m)).
func HashSignCommitment(data []byte) []byte {
	return Hash([]byte(ContextString), []byte("signCommitment"), data)
}

// HashBindingFactor computes a hash to generate binding factor and returns it as a scalar (H7(m)).
func HashBindingFactor(mid MemberID, data []byte, commitment []byte) (Scalar, error) {
	scalar, err := NewScalar(
		Hash(
			[]byte(ContextString),
			[]byte("bindingFactor"),
			sdk.Uint64ToBigEndian(uint64(mid)),
			HashSignMsg(data),
			HashSignCommitment(commitment),
		),
	)
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash bindingFactor")
	}

	return scalar, nil
}

// HashChallenge computes a hash to generate challenge of signing a signature and returns it as a scalar (H8(m)).
func HashChallenge(rawGroupPubNonce, rawGroupPubKey Point, data []byte) (Scalar, error) {
	rAddress, err := rawGroupPubNonce.Address()
	if err != nil {
		return nil, NewError(err, "parse group public nonce to address")
	}

	groupPubKey, err := rawGroupPubKey.publicKey()
	if err != nil {
		return nil, NewError(err, "parse group pubic key")
	}

	paddedPx := PaddingBytes(groupPubKey.X().Bytes(), 32)
	scalar, err := NewScalar(Hash(
		[]byte(ContextString),
		[]byte{0},
		[]byte("challenge"),
		[]byte{0},
		rAddress,
		[]byte{rawGroupPubKey[0] + 25},
		paddedPx,
		Hash(data),
	))
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash challenge")
	}

	return scalar, nil
}

// HashNonce computes a hash of the provided data for the nonce and returns it as a scalar (H9(m)).
func HashNonce(random []byte, secretKey Scalar) (Scalar, error) {
	scalar, err := NewScalar(Hash([]byte(ContextString), []byte("nonce"), random, secretKey))
	if err != nil {
		return nil, NewError(ErrNotInOrder, "hash nonce")
	}

	return scalar, nil
}
