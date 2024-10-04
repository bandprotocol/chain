package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ RouteI = &AxelarRoute{}

func (r *AxelarRoute) ValidateBasic() error {
	return nil
}

func (r *AxelarRoute) Fee() (sdk.Coins, error) {
	return sdk.NewCoins(), nil
}
