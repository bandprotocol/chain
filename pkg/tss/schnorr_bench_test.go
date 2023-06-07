package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func BenchmarkSign(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, nil)
	}
}

func BenchmarkVerify(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	sig, _ := tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, nil)
	sigR := sig.R()
	sigS := sig.S()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Verify(sigR, sigS, suite.data, suite.member1.OneTimePubKey, nil, nil)
	}
}

func BenchmarkVerifyWithCustomGenerator(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	sig, _ := tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, nil)
	sigS := sig.S()

	generator := suite.member2.OneTimePubKey
	keySym, _ := tss.ComputeKeySym(suite.member1.OneTimePrivKey, generator)
	nonceSym, _ := tss.ComputeNonceSym(suite.nonce, generator)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Verify(tss.Point(nonceSym), sigS, suite.data, keySym, tss.Point(generator), nil)
	}
}

func BenchmarkVerifyWithCustomLagrange(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	lagrange := tss.Scalar(hexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8d"))
	sig, _ := tss.Sign(suite.member1.OneTimePrivKey, suite.data, suite.nonce, lagrange)
	sigR := sig.R()
	sigS := sig.S()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Verify(sigR, sigS, suite.data, suite.member1.OneTimePubKey, nil, lagrange)
	}
}
