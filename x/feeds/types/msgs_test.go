package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ValidAuthority = sdk.AccAddress("636f736d6f7331787963726763336838396e72737671776539337a63").String()

	ValidAdmin = sdk.AccAddress("1000000001").String()
	ValidVoter = sdk.AccAddress("1000000002").String()

	ValidValidator = sdk.ValAddress("2000000001").String()

	ValidSignals = []Signal{
		{
			ID:    "CS:BAND-USD",
			Power: 10000000000,
		},
	}
	ValidParams                = DefaultParams()
	ValidReferenceSourceConfig = DefaultReferenceSourceConfig()
	ValidTimestamp             = int64(1234567890)
	ValidSignalPrices          = []SignalPrice{
		{
			Status:   SignalPriceStatusAvailable,
			SignalID: "CS:BTC-USD",
			Price:    100000 * 10e9,
		},
	}

	InvalidValidator = "invalidValidator"
	InvalidAuthority = "invalidAuthority"
	InvalidAdmin     = "invalidAdmin"
	InvalidVoter     = "invalidVoter"
)

// ====================================
// MsgSubmitSignalPrices
// ====================================

func TestNewMsgSubmitSignalPrices(t *testing.T) {
	msg := NewMsgSubmitSignalPrices(ValidValidator, ValidTimestamp, ValidSignalPrices)
	require.Equal(t, ValidValidator, msg.Validator)
	require.Equal(t, ValidTimestamp, msg.Timestamp)
	require.Equal(t, ValidSignalPrices, msg.SignalPrices)
}

func TestMsgSubmitSignalPrices_ValidateBasic(t *testing.T) {
	// Valid validator
	msg := NewMsgSubmitSignalPrices(ValidValidator, ValidTimestamp, ValidSignalPrices)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid validator
	msg = NewMsgSubmitSignalPrices(InvalidValidator, ValidTimestamp, ValidSignalPrices)
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

// ====================================
// MsgUpdateReferenceSourceConfig
// ====================================

func TestNewMsgUpdateReferenceSourceConfig(t *testing.T) {
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	require.Equal(t, ValidAdmin, msg.Admin)
	require.Equal(t, ValidReferenceSourceConfig, msg.ReferenceSourceConfig)
}

func TestMsgUpdateReferenceSourceConfig_ValidateBasic(t *testing.T) {
	// Valid admin
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid admin
	msg = NewMsgUpdateReferenceSourceConfig(InvalidAdmin, ValidReferenceSourceConfig)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgVote
// ====================================

func TestNewMsgVote(t *testing.T) {
	msg := NewMsgVote(ValidVoter, ValidSignals)
	require.Equal(t, ValidVoter, msg.Voter)
	require.Equal(t, ValidSignals, msg.Signals)
}

func TestMsgVote_ValidateBasic(t *testing.T) {
	// Valid voter
	msg := NewMsgVote(ValidVoter, ValidSignals)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid voter
	msg = NewMsgVote(InvalidVoter, ValidSignals)
	err = msg.ValidateBasic()
	require.Error(t, err)
}
