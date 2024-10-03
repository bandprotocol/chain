package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
