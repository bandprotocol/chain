package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	ValidAddress = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"
	ValidKey     = "restake"

	InvalidAddress = ""
	InvalidKey     = ""
)

// ====================================
// MsgClaimRewards
// ====================================

func TestNewMsgClaimRewards(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidKey)
	require.Equal(t, ValidAddress, msg.StakerAddress)
	require.Equal(t, ValidKey, msg.Key)
}

func TestMsgClaimRewards_Route(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidKey)
	require.Equal(t, "/restake.v1beta1.MsgClaimRewards", msg.Route())
}

func TestMsgClaimRewards_Type(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidKey)
	require.Equal(t, "/restake.v1beta1.MsgClaimRewards", msg.Type())
}

func TestMsgClaimRewards_GetSignBytes(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidKey)
	expected := `{"type":"restake/MsgClaimRewards","value":{"key":"restake","staker_address":"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgClaimRewards_GetSigners(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)
	msg := NewMsgClaimRewards(acc, ValidKey)

	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, acc, signers[0])
}

func TestMsgClaimRewards_ValidateBasic(t *testing.T) {
	acc := sdk.MustAccAddressFromBech32(ValidAddress)

	// valid address
	msg := NewMsgClaimRewards(acc, ValidKey)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid address
	msg = NewMsgClaimRewards([]byte(InvalidAddress), ValidKey)
	err = msg.ValidateBasic()
	require.Error(t, err)

	// invalid key
	msg = NewMsgClaimRewards(acc, InvalidKey)
	err = msg.ValidateBasic()
	require.Error(t, err)
}
