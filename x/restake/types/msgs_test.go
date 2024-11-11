package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ValidAddress   = sdk.AccAddress("1000000001").String()
	ValidAuthority = sdk.AccAddress("636f736d6f7331787963726763336838396e72737671776539337a63").String()
	ValidParams    = Params{
		AllowedDenoms: []string{"uband"},
	}
	ValidVault = "restake"

	InvalidAddress   = ""
	InvalidAuthority = ""
	InvalidVault     = ""
)

// ====================================
// MsgStake
// ====================================

func TestNewMsgStake(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgStake(acc, coins)
	require.Equal(t, ValidAddress, msg.StakerAddress)
	require.Equal(t, coins, msg.Coins)
}

func TestMsgStake_ValidateBasic(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))

	// valid address
	msg := NewMsgStake(acc, coins)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid address
	msg = NewMsgStake([]byte(InvalidAddress), coins)
	err = msg.ValidateBasic()
	require.Error(t, err)

	// invalid coins
	msg = NewMsgStake(acc, []sdk.Coin{
		{
			Denom:  "",
			Amount: sdkmath.NewInt(1),
		},
	})
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgUnstake
// ====================================

func TestNewMsgUnstake(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgUnstake(acc, coins)
	require.Equal(t, ValidAddress, msg.StakerAddress)
	require.Equal(t, coins, msg.Coins)
}

func TestMsgUnstake_ValidateBasic(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))

	// valid address
	msg := NewMsgUnstake(acc, coins)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid address
	msg = NewMsgUnstake([]byte(InvalidAddress), coins)
	err = msg.ValidateBasic()
	require.Error(t, err)

	// invalid coins
	msg = NewMsgUnstake(acc, []sdk.Coin{
		{
			Denom:  "",
			Amount: sdkmath.NewInt(1),
		},
	})
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgUpdateParams
// ====================================

func TestNewMsgUpdateParams(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	require.Equal(t, ValidAuthority, msg.Authority)
	require.Equal(t, ValidParams, msg.Params)
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	// Valid authority
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid authority
	msg = NewMsgUpdateParams(InvalidAuthority, ValidParams)
	err = msg.ValidateBasic()
	require.Error(t, err)
}
