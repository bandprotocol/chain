package tss_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestSignAndVerify() {
	// Sign
	sig, err := tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, nil)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(sig.R(), sig.S(), suite.data, suite.member1.OneTimePubKey, nil, nil)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(sig.R(), sig.S(), suite.fakeData, suite.member1.OneTimePubKey, nil, nil)
	suite.Require().Error(err)

	// Wrong public key case
	err = tss.Verify(sig.R(), sig.S(), suite.data, suite.fakeKey.PublicKey, nil, nil)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestSignAndVerifyWithCustomGenerator() {
	// Prepare
	generator := []byte(suite.member2.OneTimePubKey)
	fakeGenerator := []byte(suite.fakeKey.PublicKey)

	// Sign
	sig, err := tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, nil)
	suite.Require().NoError(err)

	keySym, err := tss.ComputeKeySym(suite.member1.OneTimePrivKey, generator)
	suite.Require().NoError(err)

	nonceSym, err := tss.ComputeNonceSym(suite.nonce, generator)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.data, keySym, generator, nil)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.fakeData, keySym, generator, nil)
	suite.Require().Error(err)

	// Wrong key sym case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.data, suite.fakeKey.PublicKey, generator, nil)
	suite.Require().Error(err)

	// Wrong generator case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.data, keySym, fakeGenerator, nil)
	suite.Require().Error(err)

	// Wrong nonce sym case
	err = tss.Verify(tss.Point(suite.fakeKey.PublicKey), sig.S(), suite.data, keySym, fakeGenerator, nil)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestSignAndVerifyWithCustomLagrange() {
	lagrange := tss.Scalar(hexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8d"))
	fakeLagrange := tss.Scalar(hexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8e"))

	// Sign
	sig, err := tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, lagrange)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(sig.R(), sig.S(), suite.data, suite.member1.OneTimePubKey, nil, lagrange)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(sig.R(), sig.S(), suite.fakeData, suite.member1.OneTimePubKey, nil, lagrange)
	suite.Require().Error(err)

	// Wrong public key case
	err = tss.Verify(sig.R(), sig.S(), suite.data, suite.fakeKey.PublicKey, nil, lagrange)
	suite.Require().Error(err)

	// Wrong lagrange case
	err = tss.Verify(sig.R(), sig.S(), suite.data, suite.member1.OneTimePubKey, nil, fakeLagrange)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestSignAndVerifyRandomly() {
	// Use a unique random seed each test instance and log it if the tests fail.
	seed := time.Now().Unix()
	rng := rand.New(rand.NewSource(seed))
	defer func(t *testing.T, seed int64) {
		if t.Failed() {
			t.Logf("random seed: %d", seed)
		}
	}(suite.T(), seed)

	for i := 0; i < 1000; i++ {
		// Generate a random private key.
		var privKey [32]byte
		if _, err := rng.Read(privKey[:]); err != nil {
			suite.T().Fatalf("failed to read random private key: %v", err)
		}

		// Generate a random nonce.
		var nonce [32]byte
		if _, err := rng.Read(nonce[:]); err != nil {
			suite.T().Fatalf("failed to read random nonce: %v", err)
		}

		// Generate a random hash to sign.
		var msg [1000]byte
		if _, err := rng.Read(msg[:]); err != nil {
			suite.T().Fatalf("failed to read random hash: %v", err)
		}

		// Sign the hash with the private key and then ensure the produced
		// signature is valid for the hash and public key associated with the
		// private key.
		sig, err := tss.Sign(privKey[:], msg[:], nonce[:], nil)
		if err != nil {
			suite.T().Fatalf("failed to sign\nprivate key: %x\nhash: %x", privKey, msg)
		}

		pubKey, _ := tss.PrivateKey(privKey[:]).PublicKey()
		if err := tss.Verify(sig.R(), sig.S(), msg[:], pubKey, nil, nil); err != nil {
			suite.T().
				Fatalf("failed to verify signature\nsig: %x\nhash: %x\n"+"private key: %x\npublic key: %x", sig, msg, privKey, pubKey)
		}

		// Change a random bit in the data and ensure
		// the original good signature fails to verify the new bad data.
		badMsg := make([]byte, len(msg))
		copy(badMsg, msg[:])
		randByte := rng.Intn(len(badMsg))
		randBit := rng.Intn(7)
		badMsg[randByte] ^= 1 << randBit
		if err := tss.Verify(sig.R(), sig.S(), badMsg[:], pubKey, nil, nil); err == nil {
			suite.T().Fatalf("verified signature for bad hash\nsig: %x\nhash: %x\n"+"pubkey: %x", sig, badMsg, pubKey)
		}
	}
}
