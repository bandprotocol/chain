package types

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = (*MsgSubmitSignalPrices)(nil)
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgUpdateReferenceSourceConfig)(nil)
	_ sdk.Msg = (*MsgSubmitSignals)(nil)

	_ sdk.HasValidateBasic = (*MsgSubmitSignalPrices)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateParams)(nil)
	_ sdk.HasValidateBasic = (*MsgUpdateReferenceSourceConfig)(nil)
	_ sdk.HasValidateBasic = (*MsgSubmitSignals)(nil)
)

// ====================================
// MsgSubmitSignalPrices
// ====================================

// NewMsgSubmitSignalPrices creates a new MsgSubmitSignalPrices instance.
func NewMsgSubmitSignalPrices(
	validator string,
	timestamp int64,
	prices []SignalPrice,
) *MsgSubmitSignalPrices {
	return &MsgSubmitSignalPrices{
		Validator: validator,
		Timestamp: timestamp,
		Prices:    prices,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgSubmitSignalPrices) ValidateBasic() error {
	if _, err := sdk.ValAddressFromBech32(m.Validator); err != nil {
		return err
	}

	// Map to track signal IDs for duplicate check
	signalIDSet := make(map[string]struct{})

	for _, price := range m.Prices {
		if price.PriceStatus != PriceStatusAvailable && price.Price != 0 {
			return sdkerrors.ErrInvalidRequest.Wrap(
				"price must be initial value if price status is unsupported or unavailable",
			)
		}

		// Check for duplicate signal IDs
		if _, exists := signalIDSet[price.SignalID]; exists {
			return ErrDuplicateSignalID.Wrapf(
				"duplicate signal ID found: %s", price.SignalID,
			)
		}
		signalIDSet[price.SignalID] = struct{}{}
	}

	return nil
}

// ====================================
// MsgUpdateParams
// ====================================

// NewMsgUpdateParams creates a new MsgUpdateParams instance.
func NewMsgUpdateParams(
	authority string,
	params Params,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgUpdateReferenceSourceConfig
// ====================================

// NewMsgUpdateReferenceSourceConfig creates a new MsgUpdateReferenceSourceConfig instance.
func NewMsgUpdateReferenceSourceConfig(
	admin string,
	referenceSourceConfig ReferenceSourceConfig,
) *MsgUpdateReferenceSourceConfig {
	return &MsgUpdateReferenceSourceConfig{
		Admin:                 admin,
		ReferenceSourceConfig: referenceSourceConfig,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgUpdateReferenceSourceConfig) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrap(err, "invalid admin address")
	}

	if err := m.ReferenceSourceConfig.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgSubmitSignals
// ====================================

// NewMsgSubmitSignals creates a new MsgSubmitSignals instance.
func NewMsgSubmitSignals(
	delegator string,
	signals []Signal,
) *MsgSubmitSignals {
	return &MsgSubmitSignals{
		Delegator: delegator,
		Signals:   signals,
	}
}

// ValidateBasic does a check on the provided data.
func (m *MsgSubmitSignals) ValidateBasic() error {
	// Check if the delegator address is valid
	if _, err := sdk.AccAddressFromBech32(m.Delegator); err != nil {
		return errorsmod.Wrap(err, "invalid delegator address")
	}

	// Map to track signal IDs for duplicate check
	signalIDSet := make(map[string]struct{})

	for _, signal := range m.Signals {
		// Validate Signal
		if err := signal.Validate(); err != nil {
			return err
		}

		// Check for duplicate signal IDs
		if _, exists := signalIDSet[signal.ID]; exists {
			return ErrDuplicateSignalID.Wrapf(
				"duplicate signal ID found: %s", signal.ID,
			)
		}
		signalIDSet[signal.ID] = struct{}{}
	}

	return nil
}
