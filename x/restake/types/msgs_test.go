package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	ValidAddress   = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"
	ValidAuthority = "cosmos13jt28pf6s8rgjddv8wwj8v3ngrfsccpgsdhjhw"
	ValidParams    = Params{
		AllowedDenoms: []string{"uband"},
	}
	ValidVault = "restake"

	InvalidAddress   = ""
	InvalidAuthority = ""
	InvalidVault     = ""
)

// ====================================
// MsgClaimRewards
// ====================================

func TestNewMsgClaimRewards(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidVault)
	require.Equal(t, ValidAddress, msg.StakerAddress)
	require.Equal(t, ValidVault, msg.Key)
}

func TestMsgClaimRewards_Route(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidVault)
	require.Equal(t, "/restake.v1beta1.MsgClaimRewards", msg.Route())
}

func TestMsgClaimRewards_Type(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidVault)
	require.Equal(t, "/restake.v1beta1.MsgClaimRewards", msg.Type())
}

func TestMsgClaimRewards_GetSignBytes(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidVault)
	expected := `{"type":"restake/MsgClaimRewards","value":{"key":"restake","staker_address":"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgClaimRewards_GetSigners(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidVault)

	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, acc, signers[0])
}

func TestMsgClaimRewards_ValidateBasic(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)

	// valid address
	msg := NewMsgClaimRewards(acc, ValidVault)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid address
	msg = NewMsgClaimRewards([]byte(InvalidAddress), ValidVault)
	err = msg.ValidateBasic()
	require.Error(t, err)

	// invalid vault
	msg = NewMsgClaimRewards(acc, InvalidVault)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

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

func TestMsgStake_Route(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgStake(acc, coins)
	require.Equal(t, "/restake.v1beta1.MsgStake", msg.Route())
}

func TestMsgStake_Type(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgStake(acc, coins)
	require.Equal(t, "/restake.v1beta1.MsgStake", msg.Type())
}

func TestMsgStake_GetSignBytes(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgStake(acc, coins)
	expected := `{"type":"restake/MsgStake","value":{"coins":[{"amount":"1","denom":"uband"}],"staker_address":"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgStake_GetSigners(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgStake(acc, coins)

	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, acc, signers[0])
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

func TestMsgUnstake_Route(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgUnstake(acc, coins)
	require.Equal(t, "/restake.v1beta1.MsgUnstake", msg.Route())
}

func TestMsgUnstake_Type(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgUnstake(acc, coins)
	require.Equal(t, "/restake.v1beta1.MsgUnstake", msg.Type())
}

func TestMsgUnstake_GetSignBytes(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgUnstake(acc, coins)
	expected := `{"type":"restake/MsgUnstake","value":{"coins":[{"amount":"1","denom":"uband"}],"staker_address":"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgUnstake_GetSigners(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	coins := sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(1)))
	msg := NewMsgUnstake(acc, coins)

	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, acc, signers[0])
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

func TestMsgUpdateParams_Route(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	require.Equal(t, "/restake.v1beta1.MsgUpdateParams", msg.Route())
}

func TestMsgUpdateParams_Type(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	require.Equal(t, "/restake.v1beta1.MsgUpdateParams", msg.Type())
}

func TestMsgUpdateParams_GetSignBytes(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	expected := `{"type":"restake/MsgUpdateParams","value":{"authority":"cosmos13jt28pf6s8rgjddv8wwj8v3ngrfsccpgsdhjhw","params":{"allowed_denoms":["uband"]}}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgUpdateParams_GetSigners(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, sdk.MustAccAddressFromBech32(ValidAuthority), signers[0])
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
