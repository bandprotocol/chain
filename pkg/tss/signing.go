package tss

import (
	"math/big"

	"github.com/bandprotocol/chain/v2/pkg/tss/internal/lagrange"
)

// Note: Currently, support maximum N at 20
func ComputeLagrangeCoefficient(mid MemberID, n uint64) *big.Int {
	return lagrange.ComputeCoefficient(int64(mid), int64(n))
}

// func ComputeOwnLo(I, message, B) {

// }
// func ComputeOwnPublicNonce(D, E ,lo) {

// }
// func ComputeOwnPrivateNonce(d, e, lo) {

// }

// func ComputeGroupPublicNonce(<D,E>, n, message, B) {

// }

// // SignSigning signs the signing using the given DKG context, own public key, and own private key.
// func SignSigning(
// 	mid MemberID,
// 	dkgContext []byte,
// 	ownPub PublicKey,
// 	ownPriv PrivateKey,
// ) (Signature, error) {
// 	challenge := GenerateChallengeSigning(mid, dkgContext, ownPub)
// 	return Sign(ownPriv, challenge, nil)
// }

// // VerifySigning verifies the signature of an own public key using the given DKG context, own public key, and signature.
// func VerifySigning(
// 	mid MemberID,
// 	dkgContext []byte,
// 	signature Signature,
// 	ownPub PublicKey,
// ) error {
// 	challenge := GenerateChallengeSigning(mid, dkgContext, ownPub)
// 	return Verify(signature, challenge, ownPub, nil, nil)
// }

// // GenerateChallengeSigning generates the challenge for verifying an own public key signature.
// func GenerateChallengeSigning(mid MemberID, dkgContext []byte, ownPub PublicKey) []byte {
// 	return ConcatBytes([]byte("signing"), sdk.Uint64ToBigEndian(uint64(mid)), dkgContext, ownPub)
// }
