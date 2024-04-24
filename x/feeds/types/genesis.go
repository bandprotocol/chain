package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGenesisState(params Params, feeds []Feed, ps PriceService, ds []DelegatorSignals) *GenesisState {
	return &GenesisState{
		Params:           params,
		Feeds:            feeds,
		PriceService:     ps,
		DelegatorSignals: ds,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []Feed{}, DefaultPriceService(), []DelegatorSignals{})
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, feed := range gs.Feeds {
		if err := validateInt64("power", true)(feed.Power); err != nil {
			return err
		}
		if err := validateInt64("interval", true)(feed.Interval); err != nil {
			return err
		}
		if err := validateInt64("timestamp", true)(feed.LastIntervalUpdateTimestamp); err != nil {
			return err
		}
	}

	if err := gs.PriceService.Validate(); err != nil {
		return err
	}

	for _, ds := range gs.DelegatorSignals {
		if _, err := sdk.AccAddressFromBech32(ds.Delegator); err != nil {
			return errorsmod.Wrap(err, "invalid delegator address")
		}
		for _, signal := range ds.Signals {
			if signal.ID == "" || signal.Power == 0 {
				return sdkerrors.ErrInvalidRequest.Wrap(
					"signal id cannot be empty and its power cannot be zero",
				)
			}
		}
	}

	return nil
}
