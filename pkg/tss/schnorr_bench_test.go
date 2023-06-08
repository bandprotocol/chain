package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func BenchmarkSign(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Sign(suite.privKey, suite.data, suite.nonce, nil)
	}
}

func BenchmarkVerify(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	sig, _ := tss.Sign(suite.privKey, suite.data, suite.nonce, nil)
	sigR := sig.R()
	sigS := sig.S()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Verify(sigR, sigS, suite.data, suite.pubKey, nil, nil)
	}
}

func BenchmarkVerifyWithCustomGenerator(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	sig, _ := tss.Sign(suite.privKey, suite.data, suite.nonce, nil)
	sigS := sig.S()

	generator := suite.pubKey
	keySym, _ := tss.ComputeKeySym(suite.privKey, generator)
	nonceSym, _ := tss.ComputeNonceSym(suite.nonce, generator)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Verify(tss.Point(nonceSym), sigS, suite.data, keySym, tss.Point(generator), nil)
	}
}

func BenchmarkVerifyWithCustomLagrange(b *testing.B) {
	suite := new(TSSTestSuite)
	suite.SetupTest()

	lagrange := tss.Scalar(testutil.HexDecode("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0336f8d"))
	sig, _ := tss.Sign(suite.privKey, suite.data, suite.nonce, lagrange)
	sigR := sig.R()
	sigS := sig.S()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tss.Verify(sigR, sigS, suite.data, suite.pubKey, nil, lagrange)
	}
}
