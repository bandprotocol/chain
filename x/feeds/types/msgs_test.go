package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	ValidValidator = "cosmosvaloper1vdhhxmt0wdmxzmr0wpjhyzzdttz"
	ValidAuthority = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"
	ValidAdmin     = "cosmos1quh7acmun7tx6ywkvqr53m3fe39gxu9k00t4ds"
	ValidVoter     = "cosmos13jt28pf6s8rgjddv8wwj8v3ngrfsccpgsdhjhw"
	ValidSignals   = []Signal{
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
			PriceStatus: PriceStatusAvailable,
			SignalID:    "CS:BTC-USD",
			Price:       100000 * 10e9,
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
	require.Equal(t, ValidSignalPrices, msg.Prices)
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
// MsgVoteSignals
// ====================================

func TestNewMsgVoteSignals(t *testing.T) {
	msg := NewMsgVoteSignals(ValidVoter, ValidSignals)
	require.Equal(t, ValidVoter, msg.Voter)
	require.Equal(t, ValidSignals, msg.Signals)
}

func TestMsgVoteSignals_ValidateBasic(t *testing.T) {
	// Valid voter
	msg := NewMsgVoteSignals(ValidVoter, ValidSignals)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid voter
	msg = NewMsgVoteSignals(InvalidVoter, ValidSignals)
	err = msg.ValidateBasic()
	require.Error(t, err)
}
