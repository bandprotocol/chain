package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

var (
	validAuthority = sdk.AccAddress("authority")
	validCreator   = sdk.AccAddress("creator")
	validDepositor = sdk.AccAddress("depositor")
	validParams    = types.DefaultParams()
)

// ====================================
// MsgCreateTunnel
// ====================================

func TestMsgCreateTunnel_ValidateBasic(t *testing.T) {
	signalDeviations := []types.SignalDeviation{
		{SignalID: "signal1", SoftDeviationBPS: 5000, HardDeviationBPS: 1000},
	}
	initialDeposit := sdk.NewCoins(sdk.NewInt64Coin("uband", 100))

	route := types.NewTSSRoute("chain-1", "contract-1", feedstypes.ENCODER_FIXED_POINT_ABI)
	msg, err := types.NewMsgCreateTunnel(
		signalDeviations,
		10,
		&route,
		initialDeposit,
		sdk.AccAddress([]byte("creator1")).String(),
	)
	require.NoError(t, err)

	// Valid case
	err = msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid creator
	msg.Creator = "invalidCreator"
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgUpdateRoute
// ====================================

func TestMsgUpdateRoute_ValidateBasic(t *testing.T) {
	// Valid case - empty channel
	route := types.NewIBCRoute("")
	msg, err := types.NewMsgUpdateRoute(1, route, validCreator.String())
	require.NoError(t, err)
	err = msg.ValidateBasic()
	require.NoError(t, err)

	// Valid case - correct channel format
	route = types.NewIBCRoute("channel-0")
	msg, err = types.NewMsgUpdateRoute(1, route, validCreator.String())
	require.NoError(t, err)
	err = msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid creator
	msg.Creator = "invalidCreator"
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Invalid route
	invalidRoute := types.NewIBCRoute("invalid channel")
	msg, err = types.NewMsgUpdateRoute(1, invalidRoute, validCreator.String())
	require.NoError(t, err)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgUpdateSignalsAndInterval
// ====================================

func TestMsgUpdateSignalsAndInterval_ValidateBasic(t *testing.T) {
	signalDeviations := []types.SignalDeviation{
		{SignalID: "signal1", SoftDeviationBPS: 5000, HardDeviationBPS: 1000},
	}
	msg := types.NewMsgUpdateSignalsAndInterval(1, signalDeviations, 10, validCreator.String())

	// Valid case
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid creator
	msg.Creator = "invalidCreator"
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Empty signal deviations
	msg.Creator = validCreator.String()
	msg.SignalDeviations = []types.SignalDeviation{}
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgActivate
// ====================================

func TestMsgActivate_ValidateBasic(t *testing.T) {
	msg := types.NewMsgActivate(1, validCreator.String())

	// Valid case
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid creator
	msg.Creator = "invalidCreator"
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgDeactivate
// ====================================

func TestMsgDeactivate_ValidateBasic(t *testing.T) {
	msg := types.NewMsgDeactivate(1, validCreator.String())

	// Valid case
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid creator
	msg.Creator = "invalidCreator"
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgTriggerTunnel
// ====================================

func TestMsgTriggerTunnel_ValidateBasic(t *testing.T) {
	msg := types.NewMsgTriggerTunnel(1, validCreator.String())

	// Valid case
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid creator
	msg.Creator = "invalidCreator"
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgWithdrawFromTunnel
// ====================================

func TestMsgWithdrawFromTunnel_ValidateBasic(t *testing.T) {
	// Valid withdrawer
	amount := sdk.NewCoins(sdk.NewInt64Coin("uband", 100))
	msg := types.NewMsgWithdrawFromTunnel(1, amount, validDepositor.String())
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid withdrawer
	msg.Withdrawer = "invalidWithdrawer"
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Invalid amount
	msg.Withdrawer = validDepositor.String()
	msg.Amount = sdk.Coins{}
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgUpdateParams
// ====================================

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	// Valid authority
	msg := types.NewMsgUpdateParams(validAuthority.String(), validParams)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid authority
	msg.Authority = "invalidAuthority"
	err = msg.ValidateBasic()
	require.Error(t, err)
}
