package tss_test

import (
	"encoding/hex"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/assert"
)

func (suite *TSSTestSuite) TestGenerateRound1Data() {
	data, err := tss.GenerateRound1Data(suite.member1.mid, suite.groupThreshold, suite.groupDKGContext)
	suite.Require().NoError(err)

	err = tss.VerifyOneTimeSig(suite.member1.mid, suite.groupDKGContext, data.OneTimeSig, data.OneTimePubKey)
	suite.Require().NoError(err)

	err = tss.VerifyA0Sig(suite.member1.mid, suite.groupDKGContext, data.A0Sig, data.A0PubKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestSignAndVerifyOneTime() {
	// Sign
	sig, err := tss.SignOneTime(
		suite.member1.mid,
		suite.groupDKGContext,
		suite.member1.OneTimePubKey,
		suite.member1.OneTimePrivKey,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.OneTimeSig, sig)

	// Success case
	err = tss.VerifyOneTimeSig(suite.member1.mid, suite.groupDKGContext, sig, suite.member1.OneTimePubKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestGenerateMessageOneTime() {
	challenge := tss.GenerateMessageOneTime(suite.member1.mid, suite.groupDKGContext, suite.member1.OneTimePubKey)
	assert.Equal(
		suite.T(),
		"726f756e64314f6e6554696d650000000000000001a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c0383764b806848430ed195ef8017fb4e768893ea07782e679c31e5ff1b8b453973",
		hex.EncodeToString(challenge),
	)
}

func (suite *TSSTestSuite) TestSignAndVerifyA0() {
	// Sign
	sig, err := tss.SignA0(
		suite.member1.mid,
		suite.groupDKGContext,
		suite.member1.A0PubKey,
		suite.member1.A0PrivKey,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(suite.member1.A0Sig, sig)

	// Success case
	err = tss.VerifyA0Sig(suite.member1.mid, suite.groupDKGContext, sig, suite.member1.A0PubKey)
	suite.Require().NoError(err)
}

func (suite *TSSTestSuite) TestGenerateMessageA0() {
	msg := tss.GenerateMessageA0(suite.member1.mid, suite.groupDKGContext, suite.member1.OneTimePubKey)
	assert.Equal(
		suite.T(),
		"726f756e643141300000000000000001a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c0383764b806848430ed195ef8017fb4e768893ea07782e679c31e5ff1b8b453973",
		hex.EncodeToString(msg),
	)
}
