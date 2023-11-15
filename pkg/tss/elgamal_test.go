package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
)

func (suite *TSSTestSuite) TestEncryptAndDecrypt() {
	// Prepare
	expectedValue := "e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8"
	ev, err := hex.DecodeString(expectedValue)
	suite.Require().NoError(err)

	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	// Encrypt and decrypt the value using the key.
	ec, err := tss.Encrypt(ev, key, testutil.MockNonce16Generator{})
	suite.Require().NoError(err)
	value, err := tss.Decrypt(ec, key)
	suite.Require().NoError(err)

	// Ensure the decrypted value matches the original value.
	suite.Require().Equal(expectedValue, hex.EncodeToString(value))
}

/*
func (suite *TSSTestSuite) TestEncryptAndDecrypt() {
	// Prepare
	expectedValue := "e463de6047df20228442e02c1ae58daf95e74e7a5763a94f8afe4d3b8bf97eb8"
	ev, err := hex.DecodeString(expectedValue)
	suite.Require().NoError(err)

	key, err := hex.DecodeString("4ac3ad151305074ba80e6a6abd44a5280a0502e9f06afd3e5aaad455c181ef57")
	suite.Require().NoError(err)

	// Encrypt and decrypt the value using the key.
	ec, err := tss.Encrypt(ev, key, testutil.MockNonce16Generator{})
	suite.Require().NoError(err)
	value, err := tss.Decrypt(ec, key)
	suite.Require().NoError(err)

	// Ensure the decrypted value matches the original value.
	suite.Require().Equal(expectedValue, hex.EncodeToString(value))
}
*/
