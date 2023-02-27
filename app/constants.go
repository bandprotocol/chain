package band

import sdk "github.com/cosmos/cosmos-sdk/types"

// Its intended to be .0025 uband / gas
var (
	Denom                   = "uband"
	ConsensusMinFee sdk.Dec = sdk.NewDecWithPrec(25, 4)
)
