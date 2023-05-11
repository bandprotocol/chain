package sender

import (
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Detail struct {
	GroupID types.GroupID
	Type    string
}

func GetDetail(msgs []sdk.Msg) []Detail {
	var details []Detail
	for _, msg := range msgs {
		var detail Detail
		switch t := msg.(type) {
		case *types.MsgSubmitDKGRound1:
			detail = Detail{
				GroupID: t.GroupID,
				Type:    t.Type(),
			}
		}

		details = append(details, detail)
	}

	return details
}
