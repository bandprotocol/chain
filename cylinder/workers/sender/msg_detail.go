package sender

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/tss/types"
)

const MAX_ALLOWED_GAS = uint64(2_500_000)

// GetMsgDetail represents the detail string of a message for logging.
func GetMsgDetail(msg sdk.Msg) (detail string) {
	switch t := msg.(type) {
	case *types.MsgSubmitDKGRound1:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", sdk.MsgTypeURL(t), t.GroupID)
	case *types.MsgSubmitDKGRound2:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", sdk.MsgTypeURL(t), t.GroupID)
	case *types.MsgConfirm:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", sdk.MsgTypeURL(t), t.GroupID)
	case *types.MsgComplain:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", sdk.MsgTypeURL(t), t.GroupID)
	case *types.MsgSubmitDEs:
		detail = fmt.Sprintf("Type: %s", sdk.MsgTypeURL(t))
	case *types.MsgSubmitSignature:
		detail = fmt.Sprintf("Type: %s, SigningID: %d", sdk.MsgTypeURL(t), t.SigningID)
	default:
		detail = "Type: Unknown"
	}

	return detail
}

// EstimateGas estimates the gas of the given message.
func EstimateGas(msg sdk.Msg) (gas uint64, err error) {
	switch msg.(type) {
	case *types.MsgSubmitDKGRound1, *types.MsgSubmitDKGRound2, *types.MsgConfirm, *types.MsgComplain:
		gas = 500_000
	case *types.MsgSubmitDEs:
		gas = 500_000
	case *types.MsgSubmitSignature:
		gas = 75_000
	default:
		return 0, fmt.Errorf("unsupported message type: %T", msg)
	}

	return gas, nil
}

// GetMsgDetails extracts the detail from SDK messages.
func GetMsgDetails(msgs ...sdk.Msg) (details []string) {
	for _, msg := range msgs {
		details = append(details, GetMsgDetail(msg))
	}

	return details
}
