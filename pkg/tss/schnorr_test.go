package tss_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestSignAndVerify() {
	// Sign
	sig, err := tss.Sign(suite.kpI.PrivateKey, suite.message, suite.nonce, nil)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(sig.R(), sig.S(), suite.message, suite.kpI.PublicKey, nil, nil)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(sig.R(), sig.S(), suite.fakeMessage, suite.kpI.PublicKey, nil, nil)
	suite.Require().Error(err)

	// Wrong public key case
	err = tss.Verify(sig.R(), sig.S(), suite.message, suite.fakeKp.PublicKey, nil, nil)
	suite.Require().Error(err)
}

func (suite *TSSTestSuite) TestSignAndVerifyWithCustomGenerator() {
	// Prepare
	generator := []byte(suite.kpJ.PublicKey)
	fakeGenerator := []byte(suite.fakeKp.PublicKey)

	// Sign
	sig, err := tss.Sign(suite.kpI.PrivateKey, suite.message, suite.nonce, nil)
	suite.Require().NoError(err)

	keySym, err := tss.ComputeKeySym(suite.kpI.PrivateKey, generator)
	suite.Require().NoError(err)

	nonceSym, err := tss.ComputeNonceSym(suite.nonce, generator)
	suite.Require().NoError(err)

	// Success case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.message, keySym, generator, nil)
	suite.Require().NoError(err)

	// Wrong challenge case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.fakeMessage, keySym, generator, nil)
	suite.Require().Error(err)

	// Wrong key sym case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.message, suite.fakeKp.PublicKey, generator, nil)
	suite.Require().Error(err)

	// Wrong generator case
	err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.message, keySym, fakeGenerator, nil)
	suite.Require().Error(err)

	// Wrong nonce sym case
	err = tss.Verify(tss.Point(suite.fakeKp.PublicKey), sig.S(), suite.message, keySym, fakeGenerator, nil)
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

		// Change a random bit in the message and ensure
		// the original good signature fails to verify the new bad message.
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
