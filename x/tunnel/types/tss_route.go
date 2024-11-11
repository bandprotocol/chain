package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ RouteI = &TSSRoute{}

func (r *TSSRoute) ValidateBasic() error {
	return nil
}

func (r *TSSRoute) Fee() (sdk.Coins, error) {
	return sdk.NewCoins(), nil
}
