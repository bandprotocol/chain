package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
)

func BenchmarkSign(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	for b.Loop() {
		if _, err := tss.Sign(suite.privKey, suite.challenge, suite.nonce, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerify(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	signature, _ := tss.Sign(suite.privKey, suite.challenge, suite.nonce, nil)
	signatureR := signature.R()
	signatureS := signature.S()

	for b.Loop() {
		if err := tss.Verify(signatureR, signatureS, suite.challenge, suite.pubKey, nil, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyWithCustomGenerator(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	signature, _ := tss.Sign(suite.privKey, suite.challenge, suite.nonce, nil)
	signatureS := signature.S()

	generator := suite.pubKey
	keySym, _ := tss.ComputeSecretSym(suite.privKey, generator)
	nonceSym, _ := tss.ComputeSecretSym(suite.nonce, generator)

	for b.Loop() {
		if err := tss.Verify(nonceSym, signatureS, suite.challenge, keySym, generator, nil); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyWithCustomLagrange(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	lagrange := tss.Scalar(testutil.HexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8d"))
	signature, _ := tss.Sign(suite.privKey, suite.challenge, suite.nonce, lagrange)
	signatureR := signature.R()
	signatureS := signature.S()

	for b.Loop() {
		if err := tss.Verify(signatureR, signatureS, suite.challenge, suite.pubKey, nil, lagrange); err != nil {
			b.Fatal(err)
		}
	}
}
