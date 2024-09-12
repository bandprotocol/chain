package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	ValidValidator = "cosmosvaloper1vdhhxmt0wdmxzmr0wpjhyzzdttz"
	ValidAuthority = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"
	ValidAdmin     = "cosmos1quh7acmun7tx6ywkvqr53m3fe39gxu9k00t4ds"
	ValidDelegator = "cosmos13jt28pf6s8rgjddv8wwj8v3ngrfsccpgsdhjhw"
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
	InvalidDelegator = "invalidDelegator"
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

func TestMsgSubmitSignalPrices_Route(t *testing.T) {
	msg := NewMsgSubmitSignalPrices(ValidValidator, ValidTimestamp, ValidSignalPrices)
	require.Equal(t, "/feeds.v1beta1.MsgSubmitSignalPrices", msg.Route())
}

func TestMsgSubmitSignalPrices_Type(t *testing.T) {
	msg := NewMsgSubmitSignalPrices(ValidValidator, ValidTimestamp, ValidSignalPrices)
	require.Equal(t, "/feeds.v1beta1.MsgSubmitSignalPrices", msg.Type())
}

func TestMsgSubmitSignalPrices_GetSignBytes(t *testing.T) {
	msg := NewMsgSubmitSignalPrices(ValidValidator, ValidTimestamp, ValidSignalPrices)
	expected := `{"type":"feeds/MsgSubmitSignalPrices","value":{"prices":[{"price":"1000000000000000","price_status":3,"signal_id":"CS:BTC-USD"}],"timestamp":"1234567890","validator":"cosmosvaloper1vdhhxmt0wdmxzmr0wpjhyzzdttz"}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgSubmitSignalPrices_GetSigners(t *testing.T) {
	msg := NewMsgSubmitSignalPrices(ValidValidator, ValidTimestamp, ValidSignalPrices)
	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))

	val, _ := sdk.ValAddressFromBech32(ValidValidator)
	require.Equal(t, sdk.AccAddress(val), signers[0])
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

func TestMsgUpdateParams_Route(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	require.Equal(t, "/feeds.v1beta1.MsgUpdateParams", msg.Route())
}

func TestMsgUpdateParams_Type(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	require.Equal(t, "/feeds.v1beta1.MsgUpdateParams", msg.Type())
}

func TestMsgUpdateParams_GetSignBytes(t *testing.T) {
	msg := NewMsgUpdateParams(ValidAuthority, ValidParams)
	expected := "{\"type\":\"feeds/MsgUpdateParams\",\"value\":{\"authority\":\"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8\",\"params\":{\"admin\":\"[NOT_SET]\",\"allowable_block_time_discrepancy\":\"60\",\"cooldown_time\":\"30\",\"current_feeds_update_interval\":\"28800\",\"grace_period\":\"30\",\"max_current_feeds\":\"300\",\"max_deviation_basis_point\":\"3000\",\"max_interval\":\"3600\",\"max_signal_ids_per_signing\":\"10\",\"min_deviation_basis_point\":\"50\",\"min_interval\":\"60\",\"power_step_threshold\":\"1000000000\",\"price_quorum\":\"0.300000000000000000\"}}}"
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

// ====================================
// MsgUpdateReferenceSourceConfig
// ====================================

func TestNewMsgUpdateReferenceSourceConfig(t *testing.T) {
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	require.Equal(t, ValidAdmin, msg.Admin)
	require.Equal(t, ValidReferenceSourceConfig, msg.ReferenceSourceConfig)
}

func TestMsgUpdateReferenceSourceConfig_Route(t *testing.T) {
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	require.Equal(t, "/feeds.v1beta1.MsgUpdateReferenceSourceConfig", msg.Route())
}

func TestMsgUpdateReferenceSourceConfig_Type(t *testing.T) {
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	require.Equal(t, "/feeds.v1beta1.MsgUpdateReferenceSourceConfig", msg.Type())
}

func TestMsgUpdateReferenceSourceConfig_GetSignBytes(t *testing.T) {
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	expected := `{"type":"feeds/MsgUpdateReferenceSourceConfig","value":{"admin":"cosmos1quh7acmun7tx6ywkvqr53m3fe39gxu9k00t4ds","reference_source_config":{"ipfs_hash":"hash","version":"0.0.1"}}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgUpdateReferenceSourceConfig_GetSigners(t *testing.T) {
	msg := NewMsgUpdateReferenceSourceConfig(ValidAdmin, ValidReferenceSourceConfig)
	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, sdk.MustAccAddressFromBech32(ValidAdmin), signers[0])
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
// MsgSubmitSignals
// ====================================

func TestNewMsgSubmitSignals(t *testing.T) {
	msg := NewMsgSubmitSignals(ValidDelegator, ValidSignals)
	require.Equal(t, ValidDelegator, msg.Delegator)
	require.Equal(t, ValidSignals, msg.Signals)
}

func TestMsgSubmitSignals_Route(t *testing.T) {
	msg := NewMsgSubmitSignals(ValidDelegator, ValidSignals)
	require.Equal(t, "/feeds.v1beta1.MsgSubmitSignals", msg.Route())
}

func TestMsgSubmitSignals_Type(t *testing.T) {
	msg := NewMsgSubmitSignals(ValidDelegator, ValidSignals)
	require.Equal(t, "/feeds.v1beta1.MsgSubmitSignals", msg.Type())
}

func TestMsgSubmitSignals_GetSignBytes(t *testing.T) {
	msg := NewMsgSubmitSignals(ValidDelegator, ValidSignals)
	expected := `{"type":"feeds/MsgSubmitSignals","value":{"delegator":"cosmos13jt28pf6s8rgjddv8wwj8v3ngrfsccpgsdhjhw","signals":[{"id":"CS:BAND-USD","power":"10000000000"}]}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgSubmitSignals_GetSigners(t *testing.T) {
	msg := NewMsgSubmitSignals(ValidDelegator, ValidSignals)
	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, sdk.MustAccAddressFromBech32(ValidDelegator), signers[0])
}

func TestMsgSubmitSignals_ValidateBasic(t *testing.T) {
	// Valid delegator
	msg := NewMsgSubmitSignals(ValidDelegator, ValidSignals)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid delegator
	msg = NewMsgSubmitSignals(InvalidDelegator, ValidSignals)
	err = msg.ValidateBasic()
	require.Error(t, err)
}
