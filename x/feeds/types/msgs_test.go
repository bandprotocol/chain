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
			ID:    "crypto_price.bandusd",
			Power: 10000000000,
		},
	}
	ValidParams       = DefaultParams()
	ValidPriceService = DefaultPriceService()
	ValidTimestamp    = int64(1234567890)
	ValidSubmitPrices = []SubmitPrice{
		{
			PriceStatus: PriceStatusAvailable,
			SignalID:    "crypto_price.btcusd",
			Price:       100000 * 10e9,
		},
	}

	InvalidValidator = "invalidValidator"
	InvalidAuthority = "invalidAuthority"
	InvalidAdmin     = "invalidAdmin"
	InvalidDelegator = "invalidDelegator"
)

// ====================================
// MsgSubmitPrices
// ====================================

func TestNewMsgSubmitPrices(t *testing.T) {
	msg := NewMsgSubmitPrices(ValidValidator, ValidTimestamp, ValidSubmitPrices)
	require.Equal(t, ValidValidator, msg.Validator)
	require.Equal(t, ValidTimestamp, msg.Timestamp)
	require.Equal(t, ValidSubmitPrices, msg.Prices)
}

func TestMsgSubmitPrices_Route(t *testing.T) {
	msg := NewMsgSubmitPrices(ValidValidator, ValidTimestamp, ValidSubmitPrices)
	require.Equal(t, "/feeds.v1beta1.MsgSubmitPrices", msg.Route())
}

func TestMsgSubmitPrices_Type(t *testing.T) {
	msg := NewMsgSubmitPrices(ValidValidator, ValidTimestamp, ValidSubmitPrices)
	require.Equal(t, "/feeds.v1beta1.MsgSubmitPrices", msg.Type())
}

func TestMsgSubmitPrices_GetSignBytes(t *testing.T) {
	msg := NewMsgSubmitPrices(ValidValidator, ValidTimestamp, ValidSubmitPrices)
	expected := `{"type":"feeds/MsgSubmitPrices","value":{"prices":[{"price":"1000000000000000","price_status":3,"signal_id":"crypto_price.btcusd"}],"timestamp":"1234567890","validator":"cosmosvaloper1vdhhxmt0wdmxzmr0wpjhyzzdttz"}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgSubmitPrices_GetSigners(t *testing.T) {
	msg := NewMsgSubmitPrices(ValidValidator, ValidTimestamp, ValidSubmitPrices)
	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))

	val, _ := sdk.ValAddressFromBech32(ValidValidator)
	require.Equal(t, sdk.AccAddress(val), signers[0])
}

func TestMsgSubmitPrices_ValidateBasic(t *testing.T) {
	// Valid validator
	msg := NewMsgSubmitPrices(ValidValidator, ValidTimestamp, ValidSubmitPrices)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid validator
	msg = NewMsgSubmitPrices(InvalidValidator, ValidTimestamp, ValidSubmitPrices)
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
	expected := "{\"type\":\"feeds/MsgUpdateParams\",\"value\":{\"authority\":\"cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8\",\"params\":{\"admin\":\"[NOT_SET]\",\"allowable_block_time_discrepancy\":\"60\",\"blocks_per_feeds_update\":\"28800\",\"cooldown_time\":\"30\",\"max_deviation_in_thousandth\":\"300\",\"max_interval\":\"3600\",\"max_signal_id_characters\":\"256\",\"max_supported_feeds\":\"300\",\"min_deviation_in_thousandth\":\"5\",\"min_interval\":\"60\",\"power_threshold\":\"1000000000\",\"transition_time\":\"30\"}}}"
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
// MsgUpdatePriceService
// ====================================

func TestNewMsgUpdatePriceService(t *testing.T) {
	msg := NewMsgUpdatePriceService(ValidAdmin, ValidPriceService)
	require.Equal(t, ValidAdmin, msg.Admin)
	require.Equal(t, ValidPriceService, msg.PriceService)
}

func TestMsgUpdatePriceService_Route(t *testing.T) {
	msg := NewMsgUpdatePriceService(ValidAdmin, ValidPriceService)
	require.Equal(t, "/feeds.v1beta1.MsgUpdatePriceService", msg.Route())
}

func TestMsgUpdatePriceService_Type(t *testing.T) {
	msg := NewMsgUpdatePriceService(ValidAdmin, ValidPriceService)
	require.Equal(t, "/feeds.v1beta1.MsgUpdatePriceService", msg.Type())
}

func TestMsgUpdatePriceService_GetSignBytes(t *testing.T) {
	msg := NewMsgUpdatePriceService(ValidAdmin, ValidPriceService)
	expected := `{"type":"feeds/MsgUpdatePriceService","value":{"admin":"cosmos1quh7acmun7tx6ywkvqr53m3fe39gxu9k00t4ds","price_service":{"hash":"hash","url":"https://","version":"0.0.1"}}}`
	require.Equal(t, expected, string(msg.GetSignBytes()))
}

func TestMsgUpdatePriceService_GetSigners(t *testing.T) {
	msg := NewMsgUpdatePriceService(ValidAdmin, ValidPriceService)
	signers := msg.GetSigners()
	require.Equal(t, 1, len(signers))
	require.Equal(t, sdk.MustAccAddressFromBech32(ValidAdmin), signers[0])
}

func TestMsgUpdatePriceService_ValidateBasic(t *testing.T) {
	// Valid admin
	msg := NewMsgUpdatePriceService(ValidAdmin, ValidPriceService)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid admin
	msg = NewMsgUpdatePriceService(InvalidAdmin, ValidPriceService)
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
	expected := `{"type":"feeds/MsgSubmitSignals","value":{"delegator":"cosmos13jt28pf6s8rgjddv8wwj8v3ngrfsccpgsdhjhw","signals":[{"id":"crypto_price.bandusd","power":"10000000000"}]}}`
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
