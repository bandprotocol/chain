package tss_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/pkg/tss/testutil"
	"github.com/stretchr/testify/suite"
)

type TSSTestSuite struct {
	testutil.Suite

	testCases []testutil.TestCase

	challenge []byte
	privKey   tss.Scalar
	pubKey    tss.Point
	nonce     tss.Scalar
}

func (suite *TSSTestSuite) SetupTest() {
	suite.testCases = testutil.TestCases
	suite.challenge = tss.Hash([]byte("data"))
	suite.privKey = testutil.HexDecode("83127264737dd61b4b7f8058a8418874f0e0e52ada48b39a497712a487096304")
	suite.pubKey = testutil.HexDecode("0383764b806848430ed195ef8017fb4e768893ea07782e679c31e5ff1b8b453973")
	suite.nonce = testutil.HexDecode("0000000000000000000000000000000000000000000000000000006e6f6e6365")
}

func TestTSSTestSuite(t *testing.T) {
	suite.Run(t, new(TSSTestSuite))
}
