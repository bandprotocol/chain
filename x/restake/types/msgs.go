package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg              = &MsgClaimRewards{}
	_ sdk.HasValidateBasic = &MsgClaimRewards{}
)

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
