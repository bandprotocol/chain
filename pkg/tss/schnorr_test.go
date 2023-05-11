package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {
	// prepare
	commitment := []byte("TestSignAndVerify")

	kp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	// sign
	sig, err := tss.Sign(kp.PrivateKey, commitment, nil, nil)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, commitment, kp.PublicKey, nil)
	assert.NoError(t, err)

	// wrong commitment case
	err = tss.Verify(sig, append(commitment, []byte("a")...), kp.PublicKey, nil)
	assert.Error(t, err)

	// wrong public key case
	err = tss.Verify(sig, commitment, fakeKp.PublicKey, nil)
	assert.Error(t, err)
}

func TestVerifyWithCustomGenerator(t *testing.T) {
	// prepare
	commitment := []byte("TestSignAndVerifyWithCustomGenerator")

	kp1, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	kp2, err := tss.GenerateKeyPair()
	assert.NoError(t, err)
	generator := []byte(kp2.PublicKey)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)
	fakeGenerator := []byte(fakeKp.PublicKey)

	// sign
	sig, err := tss.Sign(kp1.PrivateKey, commitment, generator, nil)
	assert.NoError(t, err)

	// generate key sym
	keySym, err := tss.GenerateKeySymIJ(kp1.PrivateKey, generator)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, commitment, keySym, generator)
	assert.NoError(t, err)

	// wrong commitment case
	err = tss.Verify(sig, append(commitment, []byte("a")...), keySym, generator)
	assert.Error(t, err)

	// wrong key sym case
	err = tss.Verify(sig, commitment, fakeKp.PublicKey, generator)
	assert.Error(t, err)

	// wrong generator case
	err = tss.Verify(sig, commitment, keySym, fakeGenerator)
	assert.Error(t, err)
}

func TestVerifyWithCustomNonce(t *testing.T) {
	// prepare
	commitment := []byte("TestSignWithCustomNonce")
	nonce := []byte("nonce")

	kp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	// sign
	sig, err := tss.Sign(kp.PrivateKey, commitment, nil, nonce)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, commitment, kp.PublicKey, nil)
	assert.NoError(t, err)

	// wrong commitment case
	err = tss.Verify(sig, append(commitment, []byte("a")...), kp.PublicKey, nil)
	assert.Error(t, err)

	// wrong public key case
	err = tss.Verify(sig, commitment, fakeKp.PublicKey, nil)
	assert.Error(t, err)
}

func TestVerifyWithCustomNonceAndGenerator(t *testing.T) {
	// prepare
	commitment := []byte("TestSignAndVerifyWithCustomNonceAndGenerator")
	nonce := []byte("nonce")

	kp1, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	kp2, err := tss.GenerateKeyPair()
	assert.NoError(t, err)
	generator := []byte(kp2.PublicKey)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)
	fakeGenerator := []byte(fakeKp.PublicKey)

	// sign
	sig, err := tss.Sign(kp1.PrivateKey, commitment, generator, nonce)
	assert.NoError(t, err)

	keySym, err := tss.GenerateKeySymIJ(kp1.PrivateKey, generator)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, commitment, keySym, generator)
	assert.NoError(t, err)

	// wrong commitment case
	err = tss.Verify(sig, append(commitment, []byte("a")...), keySym, generator)
	assert.Error(t, err)

	// wrong key sym case
	err = tss.Verify(sig, commitment, fakeKp.PublicKey, generator)
	assert.Error(t, err)

	// wrong generator case
	err = tss.Verify(sig, commitment, keySym, fakeGenerator)
	assert.Error(t, err)
}
