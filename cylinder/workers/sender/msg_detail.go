package sender

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// GetMsgDetail represents the detail string of a message for logging.
func GetMsgDetail(msg sdk.Msg) (detail string) {
	switch t := msg.(type) {
	case *types.MsgSubmitDKGRound1:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", t.Type(), t.GroupID)
	case *types.MsgSubmitDKGRound2:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", t.Type(), t.GroupID)
	case *types.MsgConfirm:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", t.Type(), t.GroupID)
	case *types.MsgComplain:
		detail = fmt.Sprintf("Type: %s, GroupID: %d", t.Type(), t.GroupID)
	case *types.MsgSubmitDEs:
		detail = fmt.Sprintf("Type: %s", t.Type())
	case *types.MsgSubmitSignature:
		detail = fmt.Sprintf("Type: %s, SigningID: %d", t.Type(), t.SigningID)
	case *types.MsgHealthCheck:
		detail = fmt.Sprintf("Type: %s", t.Type())
	default:
		detail = "Type: Unknown"
	}

	return detail
}

// GetMsgDetails extracts the detail from SDK messages.
func GetMsgDetails(msgs ...sdk.Msg) (details []string) {
	for _, msg := range msgs {
		details = append(details, GetMsgDetail(msg))
	}

	return details
}
