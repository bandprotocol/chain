package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _, _, _, _ sdk.Msg = &MsgSubmitPrices{}, &MsgUpdateParams{}, &MsgUpdatePriceService{}, &MsgSubmitSignals{}

// ====================================
// MsgSubmitPrices
// ====================================

// NewMsgSubmitPrices creates a new MsgSubmitPrices instance.
func NewMsgSubmitPrices(
	validator string,
	timestamp int64,
	prices []SubmitPrice,
) *MsgSubmitPrices {
	return &MsgSubmitPrices{
		Validator: validator,
		Timestamp: timestamp,
		Prices:    prices,
	}
}

// Route Implements Msg.
func (m MsgSubmitPrices) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitPrices) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgSubmitPrices) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgSubmitPrices) GetSigners() []sdk.AccAddress {
	validator, _ := sdk.ValAddressFromBech32(m.Validator)
	return []sdk.AccAddress{sdk.AccAddress(validator)}
}

// ValidateBasic does a check on the provided data.
func (m *MsgSubmitPrices) ValidateBasic() error {
	valAddr, err := sdk.ValAddressFromBech32(m.Validator)
	if err != nil {
		return err
	}

	if err := sdk.VerifyAddressFormat(valAddr); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("validator: %s", m.Validator)
	}

	for _, price := range m.Prices {
		if price.PriceStatus != PriceStatusAvailable && price.Price != 0 {
			return sdkerrors.ErrInvalidRequest.Wrap(
				"price must be initial value if price status is unsupported or unavailable",
			)
		}
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

// Route Implements Msg.
func (m MsgUpdateParams) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateParams) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
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
// MsgUpdatePriceService
// ====================================

// NewMsgUpdatePriceService creates a new MsgUpdatePriceService instance.
func NewMsgUpdatePriceService(
	admin string,
	priceService PriceService,
) *MsgUpdatePriceService {
	return &MsgUpdatePriceService{
		Admin:        admin,
		PriceService: priceService,
	}
}

// Route Implements Msg.
func (m MsgUpdatePriceService) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdatePriceService) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdatePriceService) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgUpdatePriceService) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Admin)}
}

// ValidateBasic does a check on the provided data.
func (m *MsgUpdatePriceService) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errorsmod.Wrap(err, "invalid admin address")
	}

	if err := m.PriceService.Validate(); err != nil {
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

// Route Implements Msg.
func (m MsgSubmitSignals) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitSignals) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgSubmitSignals) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgSubmitSignals) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Delegator)}
}

// ValidateBasic does a check on the provided data.
func (m *MsgSubmitSignals) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Delegator); err != nil {
		return errorsmod.Wrap(err, "invalid delegator address")
	}
	for _, signal := range m.Signals {
		if signal.ID == "" || signal.Power <= 0 {
			return sdkerrors.ErrInvalidRequest.Wrap(
				"signal id cannot be empty and its power must be positive",
			)
		}
	}

	return nil
}
