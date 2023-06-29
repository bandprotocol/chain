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
