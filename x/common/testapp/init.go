package testapp

import (
	bandapp "github.com/GeoDB-Limited/odin-core/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
)

const (
	Seed = 42
)

var (
	RAND *rand.Rand
)

func init() {
	RAND = rand.New(rand.NewSource(Seed))
	bandapp.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
}
