package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgClaimRewards{}

// NewMsgClaimRewards creates a new MsgClaimRewards instance
func NewMsgClaimRewards(
	stakerAddr sdk.AccAddress,
	key string,
) *MsgClaimRewards {
	return &MsgClaimRewards{
		StakerAddress: stakerAddr.String(),
		Key:           key,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgClaimRewards) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the sdk.Msg interface.
func (m MsgClaimRewards) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners implements the sdk.Msg interface.
func (m MsgClaimRewards) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.StakerAddress)}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgClaimRewards) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.StakerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid staker address: %s", err)
	}

	if len(m.Key) == 0 {
		return ErrInvalidLength.Wrap("length of key is not correct")
	}

	return nil
}
