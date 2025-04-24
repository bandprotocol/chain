package mempool

import (
	"reflect"
	"slices"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// TxWithInfo holds metadata required for a transaction to be included in a proposal.
type TxWithInfo struct {
	// Hash is the hex-encoded hash of the transaction.
	Hash string
	// BlockSpace is the block space used by the transaction.
	BlockSpace BlockSpace
	// TxBytes is the raw transaction bytes.
	TxBytes []byte
}

type TxMatchFn func(sdk.Context, sdk.Tx) bool

func NewLaneTxMatchFn(msgs []sdk.Msg, onlyFree bool) TxMatchFn {
	msgTypes := make([]reflect.Type, len(msgs))

	for i, msg := range msgs {
		msgTypes[i] = reflect.TypeOf(msg)
	}

	var matchMsgFn func(sdk.Msg) bool

	matchMsgFn = func(msg sdk.Msg) bool {
		msgExec, ok := msg.(*authz.MsgExec)
		if ok {
			subMsgs, err := msgExec.GetMessages()
			if err != nil {
				return false
			}
			for _, m := range subMsgs {
				if !matchMsgFn(m) {
					return false
				}
			}
			return true
		} else {
			return slices.Contains(msgTypes, reflect.TypeOf(msg))
		}
	}

	return func(_ sdk.Context, tx sdk.Tx) bool {
		if onlyFree {
			gasTx, ok := tx.(sdk.FeeTx)
			if !ok {
				return false
			}

			if !gasTx.GetFee().IsZero() {
				return false
			}
		}

		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}
		for _, msg := range msgs {
			if !matchMsgFn(msg) {
				return false
			}
		}
		return true
	}
}
