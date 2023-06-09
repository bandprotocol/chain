package tss_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestSignAndVerify() {
	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Sign
		sig, err := tss.Sign(member.OneTimePrivKey, suite.data, suite.nonce, nil)
		suite.Require().NoError(err)

		// Success case
		err = tss.Verify(sig.R(), sig.S(), suite.data, member.OneTimePubKey(), nil, nil)
		suite.Require().NoError(err)

		// Wrong msg case
		err = tss.Verify(sig.R(), sig.S(), []byte("fake data"), member.OneTimePubKey(), nil, nil)
		suite.Require().Error(err)

		// Wrong public key case
		err = tss.Verify(sig.R(), sig.S(), suite.data, testutil.FakePubKey, nil, nil)
		suite.Require().Error(err)
	})
}

func (suite *TSSTestSuite) TestSignAndVerifyWithCustomGenerator() {
	suite.RunOnPairMembers(
		suite.testCases,
		func(tc testutil.TestCase, memberI testutil.Member, memberJ testutil.Member) {
			// Prepare
			generator := []byte(memberJ.OneTimePubKey())
			fakeGenerator := []byte(testutil.FakePubKey)

			// Sign
			sig, err := tss.Sign(memberI.OneTimePrivKey, suite.data, suite.nonce, nil)
			suite.Require().NoError(err)

			keySym, err := tss.ComputeKeySym(memberI.OneTimePrivKey, generator)
			suite.Require().NoError(err)

			nonceSym, err := tss.ComputeNonceSym(suite.nonce, generator)
			suite.Require().NoError(err)

			// Success case
			err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.data, keySym, generator, nil)
			suite.Require().NoError(err)

			// Wrong msg case
			err = tss.Verify(tss.Point(nonceSym), sig.S(), []byte("fake data"), keySym, generator, nil)
			suite.Require().Error(err)

			// Wrong key sym case
			err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.data, testutil.FakePubKey, generator, nil)
			suite.Require().Error(err)

			// Wrong generator case
			err = tss.Verify(tss.Point(nonceSym), sig.S(), suite.data, keySym, fakeGenerator, nil)
			suite.Require().Error(err)

			// Wrong nonce sym case
			err = tss.Verify(tss.Point(testutil.FakePubKey), sig.S(), suite.data, keySym, fakeGenerator, nil)
			suite.Require().Error(err)
		})
}

func (suite *TSSTestSuite) TestSignAndVerifyWithCustomLagrange() {
	lagrange := tss.Scalar(testutil.HexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8d"))
	fakeLagrange := tss.Scalar(testutil.HexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8e"))

	suite.RunOnMember(suite.testCases, func(tc testutil.TestCase, member testutil.Member) {
		// Sign
		sig, err := tss.Sign(member.OneTimePrivKey, suite.data, suite.nonce, lagrange)
		suite.Require().NoError(err)

		// Success case
		err = tss.Verify(sig.R(), sig.S(), suite.data, member.OneTimePubKey(), nil, lagrange)
		suite.Require().NoError(err)

		// Wrong msg case
		err = tss.Verify(sig.R(), sig.S(), []byte("fake data"), member.OneTimePubKey(), nil, lagrange)
		suite.Require().Error(err)

		// Wrong public key case
		err = tss.Verify(sig.R(), sig.S(), suite.data, testutil.FakePubKey, nil, lagrange)
		suite.Require().Error(err)

		// Wrong lagrange case
		err = tss.Verify(sig.R(), sig.S(), suite.data, member.OneTimePubKey(), nil, fakeLagrange)
		suite.Require().Error(err)
	})
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
