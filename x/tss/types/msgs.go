package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgCreateGroup{}

// Route Implements Msg.
func (m MsgCreateGroup) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroup) ValidateBasic() error {
	// Validate members address
	for _, member := range m.Members {
		_, err := sdk.AccAddressFromBech32(member)
		if err != nil {
			return sdkerrors.Wrap(fmt.Errorf("validate basic error"), fmt.Sprintf("member address %s is incorrect: %s", member, err.Error()))
		}
	}

	// Check duplicate member
	if DuplicateInArray(m.Members) {
		return sdkerrors.Wrap(fmt.Errorf("validate basic error"), "members can not duplicate")
	}

	// Validate sender address
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(fmt.Errorf("validate basic error"), fmt.Sprintf("sender address %s is incorrect: %s", m.Sender, err.Error()))
	}

	// Validate threshold must be less than or equal to members but more than zero
	if m.Threshold > uint64(len(m.Members)) && m.Threshold > 0 {
		return sdkerrors.Wrap(fmt.Errorf("validate basic error"), "threshold must be less than or equal to the members but more than zero")
	}

	return nil
}

var _ sdk.Msg = &MsgSubmitDKGRound1{}

// Route Implements Msg.
func (m MsgSubmitDKGRound1) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitDKGRound1) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDKGRound1) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgSubmitDKGRound1) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound1) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return sdkerrors.Wrap(err, "member")
	}

	return nil
}

var _ sdk.Msg = &MsgSubmitDKGRound2{}

// Route Implements Msg.
func (m MsgSubmitDKGRound2) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSubmitDKGRound2) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgSubmitDKGRound2) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgSubmitDKGRound2) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgSubmitDKGRound2) ValidateBasic() error {
	// Validate member address
	_, err := sdk.AccAddressFromBech32(m.Member)
	if err != nil {
		return sdkerrors.Wrap(fmt.Errorf("validate basic error"), fmt.Sprintf("member address %s is incorrect: %s", m.Member, err.Error()))
	}

	// Validate encrypted secret shares
	for _, e := range m.Round2Share.EncryptedSecretShares {
		if len(e) != 32 {
			return sdkerrors.Wrap(fmt.Errorf("validate basic error"), fmt.Sprintf("encrypted secret shares length is not 32"))
		}
	}

	return nil
}
