package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _, _, _, _ sdk.Msg = &MsgClaimRewards{}, &MsgLockPower{}, &MsgAddRewards{}, &MsgDeactivateKey{}

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
	_, err := sdk.AccAddressFromBech32(m.StakerAddress)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgLockPower creates a new MsgLockPower instance
func NewMsgLockPower(
	stakerAddr sdk.AccAddress,
	key string,
	amount math.Int,
) *MsgLockPower {
	return &MsgLockPower{
		StakerAddress: stakerAddr.String(),
		Key:           key,
		Amount:        amount,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgLockPower) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the sdk.Msg interface.
func (m MsgLockPower) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners implements the sdk.Msg interface.
func (m MsgLockPower) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.StakerAddress)}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgLockPower) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgLockPower) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.StakerAddress)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgAddRewards creates a new MsgAddRewards instance
func NewMsgAddRewards(
	sender sdk.AccAddress,
	key string,
	rewards sdk.Coins,
) *MsgAddRewards {
	return &MsgAddRewards{
		Sender:  sender.String(),
		Key:     key,
		Rewards: rewards,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgAddRewards) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the sdk.Msg interface.
func (m MsgAddRewards) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners implements the sdk.Msg interface.
func (m MsgAddRewards) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgAddRewards) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgAddRewards) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}

	return nil
}

// NewMsgDeactivateKey creates a new MsgDeactivateKey instance
func NewMsgDeactivateKey(
	sender sdk.AccAddress,
	key string,
) *MsgDeactivateKey {
	return &MsgDeactivateKey{
		Sender: sender.String(),
		Key:    key,
	}
}

// Route implements the sdk.Msg interface.
func (m MsgDeactivateKey) Route() string { return sdk.MsgTypeURL(&m) }

// Type implements the sdk.Msg interface.
func (m MsgDeactivateKey) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners implements the sdk.Msg interface.
func (m MsgDeactivateKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgDeactivateKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgDeactivateKey) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return err
	}

	return nil
}
