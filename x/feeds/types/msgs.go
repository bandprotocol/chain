package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_, _, _, _, _ sdk.Msg = &MsgUpdateSymbols{}, &MsgRemoveSymbols{}, &MsgSubmitPrices{}, &MsgUpdateParams{}, &MsgUpdateOffChain{}
)

// ====================================
// MsgUpdateSymbols
// ====================================

// Route Implements Msg.
func (m MsgUpdateSymbols) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateSymbols) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateSymbols) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgUpdateSymbols) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Admin)}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateSymbols) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errors.Wrap(err, "invalid admin address")
	}

	for i, us := range m.Symbols {
		if err := us.Validate(); err != nil {
			return errors.Wrapf(err, "symbol: %d", i+1)
		}
	}

	return nil
}

// ====================================
// MsgRemoveSymbols
// ====================================

// Route Implements Msg.
func (m MsgRemoveSymbols) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgRemoveSymbols) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgRemoveSymbols) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgRemoveSymbols) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Admin)}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgRemoveSymbols) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errors.Wrap(err, "invalid admin address")
	}

	return nil
}

// ====================================
// MsgSubmitPrices
// ====================================

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

// ValidateBasic does a sanity check on the provided data.
func (m *MsgSubmitPrices) ValidateBasic() error {
	valAddr, err := sdk.ValAddressFromBech32(m.Validator)
	if err != nil {
		return err
	}

	if err := sdk.VerifyAddressFormat(valAddr); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("validator: %s", m.Validator)
	}

	return nil
}

// ====================================
// MsgUpdateParams
// ====================================

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

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// ====================================
// MsgUpdateOffChain
// ====================================

// Route Implements Msg.
func (m MsgUpdateOffChain) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateOffChain) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateOffChain) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for the message.
func (m *MsgUpdateOffChain) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Admin)}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateOffChain) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return errors.Wrap(err, "invalid admin address")
	}

	if err := m.OffChain.Validate(); err != nil {
		return err
	}

	return nil
}
