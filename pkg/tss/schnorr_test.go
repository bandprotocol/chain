package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func TestVerify(t *testing.T) {
	// prepare
	challenge := []byte("TestSignAndVerify")

	kp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	// sign
	sig, err := tss.Sign(kp.PrivateKey, challenge, nil, nil)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, challenge, kp.PublicKey, nil)
	assert.NoError(t, err)

	// wrong challenge case
	err = tss.Verify(sig, append(challenge, []byte("a")...), kp.PublicKey, nil)
	assert.Error(t, err)

	// wrong public key case
	err = tss.Verify(sig, challenge, fakeKp.PublicKey, nil)
	assert.Error(t, err)
}

func TestVerifyWithCustomGenerator(t *testing.T) {
	// prepare
	challenge := []byte("TestSignAndVerifyWithCustomGenerator")

	kp1, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	kp2, err := tss.GenerateKeyPair()
	assert.NoError(t, err)
	generator := []byte(kp2.PublicKey)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)
	fakeGenerator := []byte(fakeKp.PublicKey)

	// sign
	sig, err := tss.Sign(kp1.PrivateKey, challenge, generator, nil)
	assert.NoError(t, err)

	// generate key sym
	keySym, err := tss.GenerateKeySymIJ(kp1.PrivateKey, generator)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, challenge, keySym, generator)
	assert.NoError(t, err)

	// wrong challenge case
	err = tss.Verify(sig, append(challenge, []byte("a")...), keySym, generator)
	assert.Error(t, err)

	// wrong key sym case
	err = tss.Verify(sig, challenge, fakeKp.PublicKey, generator)
	assert.Error(t, err)

	// wrong generator case
	err = tss.Verify(sig, challenge, keySym, fakeGenerator)
	assert.Error(t, err)
}

func TestVerifyWithCustomNonce(t *testing.T) {
	// prepare
	challenge := []byte("TestSignWithCustomNonce")
	nonce := []byte("nonce")

	kp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	fakeKp, err := tss.GenerateKeyPair()
	assert.NoError(t, err)

	// sign
	sig, err := tss.Sign(kp.PrivateKey, challenge, nil, nonce)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, challenge, kp.PublicKey, nil)
	assert.NoError(t, err)

	// wrong challenge case
	err = tss.Verify(sig, append(challenge, []byte("a")...), kp.PublicKey, nil)
	assert.Error(t, err)

	// wrong public key case
	err = tss.Verify(sig, challenge, fakeKp.PublicKey, nil)
	assert.Error(t, err)
}

func TestVerifyWithCustomNonceAndGenerator(t *testing.T) {
	// prepare
	challenge := []byte("TestSignAndVerifyWithCustomNonceAndGenerator")
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
	sig, err := tss.Sign(kp1.PrivateKey, challenge, generator, nonce)
	assert.NoError(t, err)

	keySym, err := tss.GenerateKeySymIJ(kp1.PrivateKey, generator)
	assert.NoError(t, err)

	// success case
	err = tss.Verify(sig, challenge, keySym, generator)
	assert.NoError(t, err)

	// wrong challenge case
	err = tss.Verify(sig, append(challenge, []byte("a")...), keySym, generator)
	assert.Error(t, err)

	// wrong key sym case
	err = tss.Verify(sig, challenge, fakeKp.PublicKey, generator)
	assert.Error(t, err)

	// wrong generator case
	err = tss.Verify(sig, challenge, keySym, fakeGenerator)
	assert.Error(t, err)
}
