package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_, _, _ sdk.Msg              = &MsgStake{}, &MsgUnstake{}, &MsgUpdateParams{}
	_, _, _ sdk.HasValidateBasic = &MsgStake{}, &MsgUnstake{}, &MsgUpdateParams{}
)

// NewMsgStake creates a new MsgStake instance
func NewMsgStake(
	stakerAddr sdk.AccAddress,
	coins sdk.Coins,
) *MsgStake {
	return &MsgStake{
		StakerAddress: stakerAddr.String(),
		Coins:         coins,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgStake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.StakerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid staker address: %s", err)
	}

	if !m.Coins.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.Coins.String())
	}

	if !m.Coins.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.Coins.String())
	}

	return nil
}

// NewMsgUnstake creates a new MsgUnstake instance
func NewMsgUnstake(
	stakerAddr sdk.AccAddress,
	coins sdk.Coins,
) *MsgUnstake {
	return &MsgUnstake{
		StakerAddress: stakerAddr.String(),
		Coins:         coins,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgUnstake) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.StakerAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid staker address: %s", err)
	}

	if !m.Coins.IsValid() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.Coins.String())
	}

	if !m.Coins.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap(m.Coins.String())
	}

	return nil
}

// NewMsgUpdateParams creates a new MsgUpdateParams instance
func NewMsgUpdateParams(
	authority string,
	params Params,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
