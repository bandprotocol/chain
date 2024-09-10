package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	ValidAddress = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"
	ValidVault   = "restake"

	InvalidAddress = ""
	InvalidVault   = ""
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
