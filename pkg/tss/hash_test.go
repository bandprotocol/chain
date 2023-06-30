package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func (suite *TSSTestSuite) TestHash() {
	// Hash
	data := []byte("data")
	hash := tss.Hash(data)

	// Ensure the hash matches the expected value.
	suite.Require().Equal("8f54f1c2d0eb5771cd5bf67a6689fcd6eed9444d91a39e5ef32a9b4ae5ca14ff", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashRound1A0() {
	// Hash
	hash, err := tss.HashRound1A0([]byte("pubNonce"), 1, []byte("dkgContext"), []byte("a0Pub"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("6cca3434526a8b8d6926bb915a52fdc2a2d80c8468bb362618cdc5fca24a9f3c", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashRound1OneTime() {
	// Hash
	hash, err := tss.HashRound1OneTime([]byte("pubNonce"), 1, []byte("dkgContext"), []byte("oneTimePub"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("2e67e56a0eff4a2e60a30672159b9c1e8283cb861a40cfccef9850e1d03da213", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashRound3Complain() {
	// Hash
	hash, err := tss.HashRound3Complain(
		[]byte("pubNonce"),
		[]byte("nonceSym"),
		[]byte("oneTimePubI"),
		[]byte("oneTimePubJ"),
		[]byte("keySym"),
	)

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("955537bd86b09a963c27afff3fb965e897fda379ff04625b953da753843bc63a", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashRound3OwnPubKey() {
	// Hash
	hash, err := tss.HashRound3OwnPubKey(
		[]byte("pubNonce"),
		1,
		[]byte("dkgContext"),
		[]byte("ownPub"),
	)

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("36ad74673cc89cd164a7ce7f9356e3f02b4f2b0ce9249c7d6cc855cefecda7d9", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashSignMsg() {
	hash := tss.HashSignMsg([]byte("message"))
	suite.Require().Equal("cf92b5fe097d55d929ce3d4fc9e01fd3ae29526ba9bb33ee5dcef777b6187e6e", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashSignCommitment() {
	hash := tss.HashSignCommitment([]byte("commitment"))
	suite.Require().Equal("c4daed3c7f86d67b6906d940ca494f46614f8d28ff03e35813f4425f63bc799a", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashBindingFactor() {
	// Hash
	hash, err := tss.HashBindingFactor(1, []byte("data"), []byte("commitment"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("15dbecdbb2f2f7bedfa7e20eb87a0c553d9215c3ff1780bfc9d68eb021d7d76a", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashSigning() {
	// Hash
	hash, err := tss.HashSigning([]byte("groupPubNonce"), []byte("rawGroupPubKey"), []byte("data"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("756451b0fd9f323a627d8a5b70c8526fa627f427b76c037b1ea8842d702c0828", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashNonce() {
	// Hash
	hash, err := tss.HashNonce([]byte("random"), []byte("secretKey"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("d7313afe4b418952f705acc0247f26c92bf8e52a541f931314b6b1ec24da0dc1", hex.EncodeToString(hash))
}
