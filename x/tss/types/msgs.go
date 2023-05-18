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
			return sdkerrors.Wrap(err, "member")
		}
	}

	// Validate signer address
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(err, "sender")
	}

	// Validate threshold must be less than or equal to members
	if m.Threshold > uint64(len(m.Members)) {
		return sdkerrors.Wrap(fmt.Errorf("validate basic error"), "threshold must be less than or equal to the members")
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
	// Validate members address
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
		return sdkerrors.Wrap(err, "member")
	}

	return nil
}

var _ sdk.Msg = &MsgComplain{}

// Route Implements Msg.
func (m MsgComplain) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgComplain) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgComplain) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgComplain) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgComplain) ValidateBasic() error {
	return nil
}

var _ sdk.Msg = &MsgConfirm{}

// Route Implements Msg.
func (m MsgConfirm) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgConfirm) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgConfirm) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Member)}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgConfirm) ValidateBasic() error {
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
	return nil
}
