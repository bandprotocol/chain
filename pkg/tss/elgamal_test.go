package tss_test

import (
	"encoding/hex"
	"fmt"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestEncryptAndDecrypt() {
	// Prepare
	secretShare := "e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8"
	ev, err := hex.DecodeString(secretShare)
	suite.Require().NoError(err)

	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	// Encrypt and decrypt the value using the key.
	ec, err := tss.Encrypt(ev, key, testutil.MockNonce16Generator{})
	suite.Require().NoError(err)
	decResult, err := tss.Decrypt(ec, key)
	suite.Require().NoError(err)

	// Ensure the decrypted value matches the original value.
	suite.Require().Equal(secretShare, hex.EncodeToString(decResult))
}

func (suite *TSSTestSuite) TestEncryptError() {
	// Prepare
	ev, err := hex.DecodeString("e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8")
	suite.Require().NoError(err)
	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	_, err = tss.Encrypt(ev, key, testutil.MockNonce16Generator{
		MockGenerateFunc: func() ([]byte, error) {
			return nil, fmt.Errorf("mock error")
		},
	})
	suite.Require().EqualError(err, "mock error")

	_, err = tss.Encrypt(ev, key, testutil.MockNonce16Generator{
		MockGenerateFunc: func() ([]byte, error) {
			return []byte{1, 2, 3}, nil
		},
	})
	suite.Require().EqualError(err, "EncryptHKDF: nonce must be bytes16 but got bytes3")
}

func (suite *TSSTestSuite) TestEncryptHKDFAndDecryptHKDF() {
	tests := []struct {
		secretShare []byte
		aesKey      []byte
		nonce       []byte
		encResult   []byte
	}{
		{
			testutil.HexDecode("4739753bf14154514f1864f59d9d81f82796239a88b873f17ab27783bdf1d500"),
			testutil.HexDecode("4c5e286f39a31aee12e3583a82b973273530512030e41a5b043eab0469c28036"),
			testutil.HexDecode("6cf7af5b38ef340d44e93c60b467888e"),
			testutil.HexDecode("eb196ff9e7759f6a550e6394f515d902bff334f6ca14418c1a51ac4f4bc2af59"),
		},
		{
			testutil.HexDecode("cb0b29556849ad4219a5bb6fd7e12ac15805c9166371bcf2c4e931eeaf502807"),
			testutil.HexDecode("64540a84e00ca07eb2f34bfa98caf96c8db3b09918427bca2863ee0b2d6df31f"),
			testutil.HexDecode("d8e4136601557341913837f01885d307"),
			testutil.HexDecode("27e7475c9f5b21dc72a08055eeecbad4f049e30206265ca6652e6d681778fc3c"),
		},
		{
			testutil.HexDecode("aa445e9d2ad5e4e9ee171e6bb7ab474d09a2daf1fbc0b11c2ccd8957b3089e0d"),
			testutil.HexDecode("1e90b817f77e19e20803d0f60df96df6de500ff632cb08b4965708c165474006"),
			testutil.HexDecode("1076b23c3b93ae36a6e801e1e5923e8e"),
			testutil.HexDecode("d5788bed366003b0b2aed67a72b65868890abd5d521d52af6de60a40f9b605c3"),
		},
		{
			testutil.HexDecode("68339a2bd58cee3f89a7fab0edf43d2a52bab41107c52b6ad4de0b7f93dfd604"),
			testutil.HexDecode("44bbb210c6b34d63c740f049b9e920bfcd2e7a43a489bdab88491b9fb2da393b"),
			testutil.HexDecode("c36452595c1bb9bbdd3972b4d14d67a4"),
			testutil.HexDecode("5ccebeed4a344337100b3d957c52bd4051dac5c2e960b5db81b1077942982be9"),
		},
		{
			testutil.HexDecode("ecd308d4ff5c00559eec4a5847e9ca7f1ba37ff6a9f0b625fed95463cf5e7a0e"),
			testutil.HexDecode("a0e47458208bf8de095533461667c8f81f753aba3e2666e590316bca5f92371a"),
			testutil.HexDecode("73a4967189b5822d34f9f12ed489acbd"),
			testutil.HexDecode("2f336de9ac4579495cd84118000caaac915ddb13d3d08ccfac51ed707b001fae"),
		},
	}

	for _, t := range tests {
		ec, err := tss.EncryptHKDF(t.secretShare, t.aesKey, t.nonce)
		suite.Require().NoError(err)
		suite.Require().Equal(ec, t.encResult)
		decResult, err := tss.DecryptHKDF(append(ec, t.nonce...), t.aesKey)
		suite.Require().NoError(err)
		suite.Require().Equal(t.secretShare, decResult)
	}
}

func (suite *TSSTestSuite) TestEncryptHKDFError() {
	ev, err := hex.DecodeString("e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8")
	suite.Require().NoError(err)
	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)
	nonce, err := hex.DecodeString("eecb1c63687e0d64299b7e1409168e1e")
	suite.Require().NoError(err)

	_, err = tss.EncryptHKDF(append(ev, []byte{1}...), key, nonce)
	suite.Require().EqualError(err, "EncryptHKDF: share must be bytes32 but got bytes33")

	_, err = tss.EncryptHKDF(ev, key[:31], nonce)
	suite.Require().EqualError(err, "EncryptHKDF: aesKey must be bytes32 but got bytes31")

	_, err = tss.EncryptHKDF(ev, key, append(nonce, []byte{1}...))
	suite.Require().EqualError(err, "EncryptHKDF: nonce must be bytes16 but got bytes17")
}

func (suite *TSSTestSuite) TestDecryptHKDFError() {
	encShare, err := hex.DecodeString("e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8" + "7e0d64299b7e1409168e1e68157f0393")
	suite.Require().NoError(err)
	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	_, err = tss.DecryptHKDF(append(encShare, []byte{1}...), key)
	suite.Require().EqualError(err, "DecryptHKDF: EncSecretShare: invalid size")

	_, err = tss.DecryptHKDF(encShare, append(key, []byte{1}...))
	suite.Require().EqualError(err, "DecryptHKDF: aesKey must be bytes32 but got bytes33")
}
