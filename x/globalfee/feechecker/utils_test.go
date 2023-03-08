package feechecker_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/x/globalfee/feechecker"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type utilsTestSuite struct {
	suite.Suite
}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(utilsTestSuite))
}

func (s *utilsTestSuite) TestCombinedFeeRequirement() {
	zeroCoin1 := sdk.NewCoin("photon", sdk.ZeroInt())
	zeroCoin2 := sdk.NewCoin("stake", sdk.ZeroInt())
	zeroCoin3 := sdk.NewCoin("quark", sdk.ZeroInt())
	coin1 := sdk.NewCoin("photon", sdk.NewInt(1))
	coin2 := sdk.NewCoin("stake", sdk.NewInt(2))
	coin1High := sdk.NewCoin("photon", sdk.NewInt(10))
	coin2High := sdk.NewCoin("stake", sdk.NewInt(20))
	coinNewDenom1 := sdk.NewCoin("Newphoton", sdk.NewInt(1))
	coinNewDenom2 := sdk.NewCoin("Newstake", sdk.NewInt(1))
	// coins must be valid !!! and sorted!!!
	coinsEmpty := sdk.Coins{}
	coinsNonEmpty := sdk.Coins{coin1, coin2}.Sort()
	coinsNonEmptyHigh := sdk.Coins{coin1High, coin2High}.Sort()
	coinsNonEmptyOneHigh := sdk.Coins{coin1High, coin2}.Sort()
	coinsNewDenom := sdk.Coins{coinNewDenom1, coinNewDenom2}.Sort()
	coinsNewOldDenom := sdk.Coins{coin1, coinNewDenom1}.Sort()
	coinsNewOldDenomHigh := sdk.Coins{coin1High, coinNewDenom1}.Sort()
	coinsCointainZero := sdk.Coins{coin1, zeroCoin2}.Sort()
	coinsCointainZeroNewDenom := sdk.Coins{coin1, zeroCoin3}.Sort()
	coinsAllZero := sdk.Coins{zeroCoin1, zeroCoin2}.Sort()
	tests := map[string]struct {
		cGlobal  sdk.Coins
		c        sdk.Coins
		combined sdk.Coins
	}{
		"global fee empty, min fee empty, combined fee empty": {
			cGlobal:  coinsEmpty,
			c:        coinsEmpty,
			combined: coinsEmpty,
		},
		"global fee empty, min fee nonempty, combined fee empty": {
			cGlobal:  coinsEmpty,
			c:        coinsNonEmpty,
			combined: coinsEmpty,
		},
		"global fee nonempty, min fee empty, combined fee = global fee": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmpty,
			combined: coinsNonEmpty,
		},
		"global fee and min fee have overlapping denom, min fees amounts are all higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyHigh,
			combined: coinsNonEmptyHigh,
		},
		"global fee and min fee have overlapping denom, one of min fees amounts is higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyOneHigh,
			combined: coinsNonEmptyOneHigh,
		},
		"global fee and min fee have no overlapping denom, combined fee = global fee": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewDenom,
			combined: coinsNonEmpty,
		},
		"global fees and min fees have partial overlapping denom, min fee amount <= global fee amount, combined fees = global fees": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenom,
			combined: coinsNonEmpty,
		},
		"global fees and min fees have partial overlapping denom, one min fee amount > global fee amount, combined fee = overlapping highest": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenomHigh,
			combined: sdk.Coins{coin1High, coin2},
		},
		"global fees have zero fees, min fees have overlapping non-zero fees, combined fees = overlapping highest": {
			cGlobal:  coinsCointainZero,
			c:        coinsNonEmpty,
			combined: sdk.Coins{coin1, coin2},
		},
		"global fees have zero fees, min fees have overlapping zero fees": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
		"global fees have zero fees, min fees have non-overlapping zero fees": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZeroNewDenom,
			combined: coinsCointainZero,
		},
		"global fees are all zero fees, min fees have overlapping zero fees": {
			cGlobal:  coinsAllZero,
			c:        coinsAllZero,
			combined: coinsAllZero,
		},
		"global fees are all zero fees, min fees have overlapping non-zero fees, combined fee = overlapping highest": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZeroNewDenom,
			combined: sdk.Coins{coin1, zeroCoin2},
		},
		"global fees are all zero fees, fees have one overlapping non-zero fee": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			allFees := feechecker.CombinedFeeRequirement(test.cGlobal, test.c)
			s.Require().Equal(test.combined, allFees)
		})
	}
}
