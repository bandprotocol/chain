package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestVerifyComplain(t *testing.T) {
	// prepare
	kp1, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	kp2, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	// sign
	sig, keySym, nonceSym, err := tss.SignComplain(kp1.PublicKey, kp2.PublicKey, kp1.PrivateKey)
	assert.NoError(t, err)

	// success case
	err = tss.VerifyComplainSig(kp1.PublicKey, kp2.PublicKey, keySym, nonceSym, sig)
	assert.NoError(t, err)

	// wrong public key I case
	err = tss.VerifyComplainSig(fakeKp.PublicKey, kp2.PublicKey, keySym, nonceSym, sig)
	assert.Error(t, err)

	// wrong public key J case
	err = tss.VerifyComplainSig(kp1.PublicKey, fakeKp.PublicKey, keySym, nonceSym, sig)
	assert.Error(t, err)

	// wrong key sym case
	err = tss.VerifyComplainSig(kp1.PublicKey, kp2.PublicKey, fakeKp.PublicKey, nonceSym, sig)
	assert.Error(t, err)

	// wrong nonce sym case
	err = tss.VerifyComplainSig(kp1.PublicKey, kp2.PublicKey, keySym, fakeKp.PublicKey, sig)
	assert.Error(t, err)
}
