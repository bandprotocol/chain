package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgClaimRewards{}

// NewMsgClaimRewards creates a new MsgClaimRewards instance
func NewMsgClaimRewards(
	lockerAddr sdk.AccAddress,
	key string,
) *MsgClaimRewards {
	return &MsgClaimRewards{
		LockerAddress: lockerAddr.String(),
		Key:           key,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgClaimRewards) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the sdk.Msg interface.
func (m MsgClaimRewards) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners implements the sdk.Msg interface.
func (m MsgClaimRewards) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.LockerAddress)}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgClaimRewards) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgClaimRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.LockerAddress)
	if err != nil {
		return err
	}

	if len(m.Key) == 0 {
		return ErrInvalidLength.Wrap("length of key is not correct")
	}

	return nil
}
