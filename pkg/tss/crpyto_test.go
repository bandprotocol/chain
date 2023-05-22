package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestEncryptAndDecrypt() {
	// Prepare
	encryptedValue := "e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8"
	ev, err := hex.DecodeString(encryptedValue)
	suite.Require().NoError(err)

	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	// Encrypt and decrypt the value using the key.
	ec := tss.Encrypt(ev, key)
	value := tss.Decrypt(ec, key)

	// Ensure the decrypted value matches the original value.
	suite.Require().Equal(encryptedValue, hex.EncodeToString(value))
}

func (suite *TSSTestSuite) TestHash() {
	// Hash
	data := []byte("data")
	hash := tss.Hash(data)

	// Ensure the hash matches the expected value.
	suite.Require().Equal("8f54f1c2d0eb5771cd5bf67a6689fcd6eed9444d91a39e5ef32a9b4ae5ca14ff", hex.EncodeToString(hash))
}
