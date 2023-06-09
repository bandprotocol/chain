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

func (s *utilsTestSuite) TestCombinedGasPricesRequirement() {
	zeroCoin1 := sdk.NewDecCoin("photon", sdk.ZeroInt())
	zeroCoin2 := sdk.NewDecCoin("stake", sdk.ZeroInt())
	zeroCoin3 := sdk.NewDecCoin("quark", sdk.ZeroInt())
	coin1 := sdk.NewDecCoin("photon", sdk.NewInt(1))
	coin2 := sdk.NewDecCoin("stake", sdk.NewInt(2))
	coin1High := sdk.NewDecCoin("photon", sdk.NewInt(10))
	coin2High := sdk.NewDecCoin("stake", sdk.NewInt(20))
	coinNewDenom1 := sdk.NewDecCoin("Newphoton", sdk.NewInt(1))
	coinNewDenom2 := sdk.NewDecCoin("Newstake", sdk.NewInt(1))
	// coins must be valid !!! and sorted!!!
	coinsEmpty := sdk.DecCoins{}
	coinsNonEmpty := sdk.DecCoins{coin1, coin2}.Sort()
	coinsNonEmptyHigh := sdk.DecCoins{coin1High, coin2High}.Sort()
	coinsNonEmptyOneHigh := sdk.DecCoins{coin1High, coin2}.Sort()
	coinsNewDenom := sdk.DecCoins{coinNewDenom1, coinNewDenom2}.Sort()
	coinsNewOldDenom := sdk.DecCoins{coin1, coinNewDenom1}.Sort()
	coinsNewOldDenomHigh := sdk.DecCoins{coin1High, coinNewDenom1}.Sort()
	coinsCointainZero := sdk.DecCoins{coin1, zeroCoin2}.Sort()
	coinsCointainZeroNewDenom := sdk.DecCoins{coin1, zeroCoin3}.Sort()
	coinsAllZero := sdk.DecCoins{zeroCoin1, zeroCoin2}.Sort()
	tests := map[string]struct {
		cGlobal  sdk.DecCoins
		c        sdk.DecCoins
		combined sdk.DecCoins
	}{
		"global price empty, min price empty, combined price empty": {
			cGlobal:  coinsEmpty,
			c:        coinsEmpty,
			combined: coinsEmpty,
		},
		"global price empty, min price nonempty, combined price nonempty": {
			cGlobal:  coinsEmpty,
			c:        coinsNonEmpty,
			combined: coinsNonEmpty,
		},
		"global price nonempty, min price nonempty, combined price nonempty": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmpty,
			combined: coinsNonEmpty,
		},
		"global price and min price have overlapping denom, min prices amounts are all higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyHigh,
			combined: coinsNonEmptyHigh,
		},
		"global price and min price have overlapping denom, one of min prices amounts is higher": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNonEmptyOneHigh,
			combined: coinsNonEmptyOneHigh,
		},
		"global price and min price have no overlapping denom, combined price = global price": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewDenom,
			combined: coinsNonEmpty,
		},
		"global prices and min prices have partial overlapping denom, min price amount <= global price amount, combined prices = global prices": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenom,
			combined: coinsNonEmpty,
		},
		"global prices and min prices have partial overlapping denom, one min price amount > global price amount, combined price = overlapping highest": {
			cGlobal:  coinsNonEmpty,
			c:        coinsNewOldDenomHigh,
			combined: sdk.DecCoins{coin1High, coin2},
		},
		"global prices have zero prices, min prices have overlapping non-zero prices, combined prices = overlapping highest": {
			cGlobal:  coinsCointainZero,
			c:        coinsNonEmpty,
			combined: sdk.DecCoins{coin1, coin2},
		},
		"global prices have zero prices, min prices have overlapping zero prices": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
		"global prices have zero prices, min prices have non-overlapping zero prices": {
			cGlobal:  coinsCointainZero,
			c:        coinsCointainZeroNewDenom,
			combined: coinsCointainZero,
		},
		"global prices are all zero prices, min prices have overlapping zero prices": {
			cGlobal:  coinsAllZero,
			c:        coinsAllZero,
			combined: coinsAllZero,
		},
		"global prices are all zero prices, min prices have overlapping non-zero prices, combined price = overlapping highest": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZeroNewDenom,
			combined: sdk.DecCoins{coin1, zeroCoin2},
		},
		"global prices are all zero prices, prices have one overlapping non-zero price": {
			cGlobal:  coinsAllZero,
			c:        coinsCointainZero,
			combined: coinsCointainZero,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			allPrices := feechecker.CombinedGasPricesRequirement(test.cGlobal, test.c)
			s.Require().Equal(test.combined, allPrices)
		})
	}
}
