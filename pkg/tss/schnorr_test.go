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
	pass, err := tss.Verify(sig, commitment, kp.PublicKey, nil)
	assert.NoError(t, err)
	assert.True(t, pass)

	// wrong commitment case
	pass, err = tss.Verify(sig, append(commitment, []byte("a")...), kp.PublicKey, nil)
	assert.NoError(t, err)
	assert.False(t, pass)

	// wrong public key case
	pass, err = tss.Verify(sig, commitment, fakeKp.PublicKey, nil)
	assert.NoError(t, err)
	assert.False(t, pass)
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
	sig, err := tss.Sign(kp1.PrivateKey, commitment, &generator, nil)
	assert.NoError(t, err)

	// generate key sym
	keySym, err := tss.GenerateKeySymIJ(kp1.PrivateKey, generator)
	assert.NoError(t, err)

	// success case
	pass, err := tss.Verify(sig, commitment, keySym, &generator)
	assert.NoError(t, err)
	assert.True(t, pass)

	// wrong commitment case
	pass, err = tss.Verify(sig, append(commitment, []byte("a")...), keySym, &generator)
	assert.NoError(t, err)
	assert.False(t, pass)

	// wrong key sym case
	pass, err = tss.Verify(sig, commitment, fakeKp.PublicKey, &generator)
	assert.NoError(t, err)
	assert.False(t, pass)

	// wrong generator case
	pass, err = tss.Verify(sig, commitment, keySym, &fakeGenerator)
	assert.NoError(t, err)
	assert.False(t, pass)
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
	sig, err := tss.Sign(kp.PrivateKey, commitment, nil, &nonce)
	assert.NoError(t, err)

	// success case
	pass, err := tss.Verify(sig, commitment, kp.PublicKey, nil)
	assert.NoError(t, err)
	assert.True(t, pass)

	// wrong commitment case
	pass, err = tss.Verify(sig, append(commitment, []byte("a")...), kp.PublicKey, nil)
	assert.NoError(t, err)
	assert.False(t, pass)

	// wrong public key case
	pass, err = tss.Verify(sig, commitment, fakeKp.PublicKey, nil)
	assert.NoError(t, err)
	assert.False(t, pass)
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
	sig, err := tss.Sign(kp1.PrivateKey, commitment, &generator, &nonce)
	assert.NoError(t, err)

	keySym, err := tss.GenerateKeySymIJ(kp1.PrivateKey, generator)
	assert.NoError(t, err)

	// success case
	pass, err := tss.Verify(sig, commitment, keySym, &generator)
	assert.NoError(t, err)
	assert.True(t, pass)

	// wrong commitment case
	pass, err = tss.Verify(sig, append(commitment, []byte("a")...), keySym, &generator)
	assert.NoError(t, err)
	assert.False(t, pass)

	// wrong key sym case
	pass, err = tss.Verify(sig, commitment, fakeKp.PublicKey, &generator)
	assert.NoError(t, err)
	assert.False(t, pass)

	// wrong generator case
	pass, err = tss.Verify(sig, commitment, keySym, &fakeGenerator)
	assert.NoError(t, err)
	assert.False(t, pass)
}
