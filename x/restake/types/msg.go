package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgWithdrawRewards{}
)

// NewMsgWithdrawRewards creates a new MsgWithdrawRewards instance
func NewMsgWithdrawRewards(
	address string,
) *MsgWithdrawRewards {
	return &MsgWithdrawRewards{
		Address: address,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgWithdrawRewards) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the sdk.Msg interface.
func (m MsgWithdrawRewards) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners implements the sdk.Msg interface.
func (m MsgWithdrawRewards) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Address)}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgWithdrawRewards) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgWithdrawRewards) ValidateBasic() error {
	return nil
}
