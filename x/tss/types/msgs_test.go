package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/pkg/tss/testutil"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

var (
	validSender = "cosmos1xxjxtce966clgkju06qp475j663tg8pmklxcy8"

	validTssPoint = tss.Point(
		testutil.HexDecode("03a50a76f243836311dd2fbaaf8b5185f5f7f34bd4cb99ac7309af18f89703960b"),
	)
	validSignature = tss.Signature(
		testutil.HexDecode(
			"022840aa88b0e38a64b445c2e5150c6df84e449a1d8b4c3d74f5ed1790be92b62c71f9828712146a231e9988f6148890f98661913b5356d4da657f59e0f99310a4",
		),
	)
	validEncSecretShare = tss.EncSecretShare(
		testutil.HexDecode(
			"00bf89d839d9b4cbfea51435c7e49ac8696e6c1faf1715e1b343e62f90027d4b7ba8fb095282c02a43d59cd8e1a0708b",
		),
	)
	validComplaintSignature = tss.ComplaintSignature(
		testutil.HexDecode(
			"02a55f7d417d1b51d91e6097473f00f528291aaa0dd11733e83eb85680ed5d4e36034946dba60574" +
				"e576aef1c252e48db7c2c40f828efdb374ec8bd48ea36af06ac89fe3b8aef036713c547118f5a0ad" +
				"b8108dfe19b4067081f26a2fe27a87f60c0b",
		),
	)
)

// ====================================
// MsgSubmitDKGRound1
// ====================================

func TestNewMsgSubmitDKGRound1(t *testing.T) {
	validRound1Info := types.Round1Info{
		MemberID:           1,
		CoefficientCommits: tss.Points{validTssPoint},
		OneTimePubKey:      validTssPoint,
		A0Signature:        validSignature,
		OneTimeSignature:   validSignature,
	}

	msg := types.NewMsgSubmitDKGRound1(1, validRound1Info, validSender)
	require.Equal(t, tss.GroupID(1), msg.GroupID)
	require.Equal(t, validRound1Info, msg.Round1Info)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgSubmitDKGRound1_ValidateBasic(t *testing.T) {
	validRound1Info := types.Round1Info{
		MemberID:           1,
		CoefficientCommits: tss.Points{validTssPoint},
		OneTimePubKey:      validTssPoint,
		A0Signature:        validSignature,
		OneTimeSignature:   validSignature,
	}

	// Valid input
	msg := types.NewMsgSubmitDKGRound1(1, validRound1Info, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid input
	invalidRound1Info := validRound1Info
	invalidRound1Info.OneTimeSignature = tss.Signature(testutil.HexDecode("0020"))
	msg = types.NewMsgSubmitDKGRound1(1, invalidRound1Info, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgSubmitDKGRound2
// ====================================

func TestNewMsgSubmitDKGRound2(t *testing.T) {
	validRound2Info := types.Round2Info{
		MemberID:              1,
		EncryptedSecretShares: tss.EncSecretShares{validEncSecretShare},
	}

	msg := types.NewMsgSubmitDKGRound2(1, validRound2Info, validSender)
	require.Equal(t, tss.GroupID(1), msg.GroupID)
	require.Equal(t, validRound2Info, msg.Round2Info)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgSubmitDKGRound2_ValidateBasic(t *testing.T) {
	validRound2Info := types.Round2Info{
		MemberID:              1,
		EncryptedSecretShares: tss.EncSecretShares{validEncSecretShare},
	}

	// Valid input
	msg := types.NewMsgSubmitDKGRound2(1, validRound2Info, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid input
	invalidRound2Info := validRound2Info
	invalidRound2Info.EncryptedSecretShares = tss.EncSecretShares{tss.EncSecretShare(testutil.HexDecode("0020"))}
	msg = types.NewMsgSubmitDKGRound2(1, invalidRound2Info, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgComplain
// ====================================

func TestNewMsgComplain(t *testing.T) {
	validComplaints := []types.Complaint{
		{Complainant: 1, Respondent: 2, KeySym: validTssPoint, Signature: tss.ComplaintSignature(validSignature)},
	}
	msg := types.NewMsgComplain(1, validComplaints, validSender)
	require.Equal(t, tss.GroupID(1), msg.GroupID)
	require.Equal(t, validComplaints, msg.Complaints)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgComplain_ValidateBasic(t *testing.T) {
	validComplaints := []types.Complaint{
		{Complainant: 1, Respondent: 2, KeySym: validTssPoint, Signature: validComplaintSignature},
	}

	// Valid input
	msg := types.NewMsgComplain(1, validComplaints, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid input
	invalidComplaints := validComplaints
	invalidComplaints[0].Signature = tss.ComplaintSignature(testutil.HexDecode("0020"))
	msg = types.NewMsgComplain(1, invalidComplaints, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)

	// Another invalid input
	invalidComplaints = validComplaints
	invalidComplaints[0].Respondent = validComplaints[0].Complainant
	msg = types.NewMsgComplain(1, invalidComplaints, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgSubmitConfirm
// ====================================

func TestNewMsgConfirm(t *testing.T) {
	msg := types.NewMsgConfirm(1, 1, validSignature, validSender)
	require.Equal(t, tss.GroupID(1), msg.GroupID)
	require.Equal(t, tss.MemberID(1), msg.MemberID)
	require.Equal(t, validSignature, msg.OwnPubKeySig)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgConfirm_ValidateBasic(t *testing.T) {
	// Valid input
	msg := types.NewMsgConfirm(1, 1, validSignature, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// Invalid input
	msg = types.NewMsgConfirm(1, 1, tss.Signature(testutil.HexDecode("0020")), validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgSubmitDE
// ====================================

func TestNewMsgSubmitDE(t *testing.T) {
	pubD := validTssPoint
	pubE := tss.Point(testutil.HexDecode("03a50a76f243836311dd2fbaaf8b5185f5f7f34bd4cb99ac7309af18f89703960c"))

	msg := types.NewMsgSubmitDEs([]types.DE{
		{PubD: pubD, PubE: pubE},
	}, validSender)

	require.Equal(t, pubD, msg.DEs[0].PubD)
	require.Equal(t, pubE, msg.DEs[0].PubE)
	require.Equal(t, validSender, msg.Sender)
}

func TestMsgSubmitDE_ValidateBasic(t *testing.T) {
	pubD := validTssPoint
	pubE := tss.Point(testutil.HexDecode("02117a767c77af0b9630991393ccbfe96930008987ee315ce205ae8b004795ad41"))

	// valid input
	msg := types.NewMsgSubmitDEs([]types.DE{
		{PubD: pubD, PubE: pubE},
	}, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid input
	msg = types.NewMsgSubmitDEs([]types.DE{
		{PubD: pubD, PubE: tss.Point(testutil.HexDecode("0020"))},
	}, validSender)
	err = msg.ValidateBasic()
	require.Error(t, err)
}

// ====================================
// MsgSubmitSignature
// ====================================

func TestNewMsgSubmitSignature(t *testing.T) {
	msg := types.NewMsgSubmitSignature(1, 1, validSignature, validSender)

	require.Equal(t, tss.SigningID(1), msg.SigningID)
	require.Equal(t, tss.MemberID(1), msg.MemberID)
	require.Equal(t, validSignature, msg.Signature)
	require.Equal(t, validSender, msg.Signer)
}

func TestMsgSubmitSignature_ValidateBasic(t *testing.T) {
	msg := types.NewMsgSubmitSignature(1, 1, validSignature, validSender)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	// invalid input
	msg = types.NewMsgSubmitSignature(1, 1, validSignature, "invalid-sender")
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
