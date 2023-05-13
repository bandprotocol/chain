package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestCalculateEncryptedSecretShares(t *testing.T) {
	// n = 4
	// t = 2
	// 0-3 for people in groups, 4-5 for coefficients
	kps, err := tss.GenerateKeyPairs(6)
	assert.NoError(t, err)

	scalars, err := tss.ComputeEncryptedSecretShares(
		1,
		kps[0].PrivateKey,
		tss.PublicKeys{kps[0].PublicKey, kps[1].PublicKey, kps[2].PublicKey, kps[3].PublicKey},
		tss.Scalars{tss.Scalar(kps[4].PrivateKey), tss.Scalar(kps[5].PrivateKey)},
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(scalars))
}
