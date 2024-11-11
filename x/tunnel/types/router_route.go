package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// RouterRoute defines the Router route for the tunnel module
var _ RouteI = &RouterRoute{}

// RouterRoute defines the Router route for the tunnel module
func (r *RouterRoute) ValidateBasic() error {
	return nil
}

func (r *RouterRoute) Fee() (sdk.Coins, error) {
	return sdk.NewCoins(), nil
}
