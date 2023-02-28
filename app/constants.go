package band

import sdk "github.com/cosmos/cosmos-sdk/types"

// Its intended to be .0025 uband / gas
// TODO: change it back from 0.05 to 0.0025
var (
	Denom                   = "uband"
	ConsensusMinFee sdk.Dec = sdk.NewDecWithPrec(5, 2)
)
