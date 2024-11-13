package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v3/pkg/tss"
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
	suite.Require().Equal("1b8081d885dc6226a2737228e91270a27b488475079c3e2e46ba8accdfa928ce", hex.EncodeToString(hash))
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

func (suite *TSSTestSuite) TestHashChallenge() {
	// Hash
	r, _ := hex.DecodeString("028438d0c62660fa061c49e65b5f5b613a7334776f2486c4e58e2b52e36ea6a783")
	y, _ := hex.DecodeString("026f544da336a6f74af8f3dea80226ccab1cae265362ebee7546530797e3d565d2")
	hash, err := tss.HashChallenge(r, y, []byte("data"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("8381f0fe6ea8135c3f71866adb24b64639069ac0556603f7a9f80ab4fceef76a", hex.EncodeToString(hash))
}

func (suite *TSSTestSuite) TestHashNonce() {
	// Hash
	hash, err := tss.HashNonce([]byte("random"), []byte("secretKey"))

	// Ensure the hash matches the expected value.
	suite.Require().Nil(err)
	suite.Require().Equal("d7313afe4b418952f705acc0247f26c92bf8e52a541f931314b6b1ec24da0dc1", hex.EncodeToString(hash))
}
