package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	types "github.com/bandprotocol/chain/v3/x/bandtss/types"
)

var (
	validMembers = []string{
		"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8",
		"cosmos1quh7acmun7tx6ywkvqr53m3fe39gxu9k00t4ds",
	}

	validSender = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"
)

// ====================================
// MsgTransitionGroup
// ====================================

func TestNewMsgTransitionGroup(t *testing.T) {
	execTime := time.Now().Add(time.Hour)
	msg := types.NewMsgTransitionGroup(validMembers, 1, execTime, validSender)
	require.Equal(t, validMembers, msg.Members)
	require.Equal(t, uint64(1), msg.Threshold)
	require.Equal(t, execTime, msg.ExecTime)
	require.Equal(t, validSender, msg.Authority)
}

func TestMsgTransitionGroup_ValidateBasic(t *testing.T) {
	// Valid input
	execTime := time.Now().Add(time.Hour)
	msg := types.NewMsgTransitionGroup(validMembers, 1, execTime, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// duplicate members
	duplicatedMembers := []string{validMembers[0], validMembers[0]}
	msg = types.NewMsgTransitionGroup(duplicatedMembers, 1, execTime, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)

	// validate threshold
	msg = types.NewMsgTransitionGroup(validMembers, 3, execTime, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgForceTransitionGroup
// ====================================

func TestNewMsgForceTransitionGroup(t *testing.T) {
	execTime := time.Now().Add(time.Hour)
	msg := types.NewMsgForceTransitionGroup(1, execTime, validSender)
	require.Equal(t, tss.GroupID(1), msg.IncomingGroupID)
	require.Equal(t, execTime, msg.ExecTime)
	require.Equal(t, validSender, msg.Authority)
}

func TestMsgForceTransitionGroup_ValidateBasic(t *testing.T) {
	// Valid input
	execTime := time.Now().Add(time.Hour)
	msg := types.NewMsgForceTransitionGroup(1, execTime, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid input
	msg = types.NewMsgForceTransitionGroup(0, execTime, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgRequestSignature
// ====================================

func TestNewMsgRequestSignature(t *testing.T) {
	content := &types.GroupTransitionSignatureOrder{}
	feeLimit := sdk.NewCoins(sdk.NewInt64Coin("uband", 100))

	msg, err := types.NewMsgRequestSignature(content, feeLimit, validSender)
	require.NoError(t, err)
	require.Equal(t, feeLimit, msg.FeeLimit)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgRequestSignature_ValidateBasic(t *testing.T) {
	// Valid input
	content := &types.GroupTransitionSignatureOrder{}
	feeLimit := sdk.NewCoins(sdk.NewInt64Coin("uband", 100))

	msg, err := types.NewMsgRequestSignature(content, feeLimit, validSender)
	require.NoError(t, err)
	err = msg.ValidateBasic()
	require.NoError(t, err)

	// zero coins
	msg, err = types.NewMsgRequestSignature(content, sdk.NewCoins(), validSender)
	require.NoError(t, err)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgActivate
// ====================================

func TestNewMsgActivate(t *testing.T) {
	msg := types.NewMsgActivate(validSender, 1)
	require.Equal(t, tss.GroupID(1), msg.GroupID)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgActivate_ValidateBasic(t *testing.T) {
	// Valid input
	msg := types.NewMsgActivate(validSender, 1)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid input
	msg = types.NewMsgActivate(validSender, 0)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgHeartbeat
// ====================================

func TestNewMsgHeartbeat(t *testing.T) {
	msg := types.NewMsgHeartbeat(validSender, 1)
	require.Equal(t, tss.GroupID(1), msg.GroupID)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgHeartbeat_ValidateBasic(t *testing.T) {
	// Valid input
	msg := types.NewMsgHeartbeat(validSender, 1)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid input
	msg = types.NewMsgHeartbeat(validSender, 0)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgUpdateParams
// ====================================

func TestNewMsgUpdateParams(t *testing.T) {
	params := types.DefaultParams()
	msg := types.NewMsgUpdateParams(validSender, params)

	require.Equal(t, params, msg.Params)
	require.Equal(t, validSender, msg.Authority)
}

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	params := types.DefaultParams()
	msg := types.NewMsgUpdateParams(validSender, params)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid input
	msg = types.NewMsgUpdateParams("invalid-sender", params)
	err = msg.ValidateBasic()
	require.Error(t, err)
}
