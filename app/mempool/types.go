package mempool

import (
	"reflect"
	"slices"

	"github.com/cosmos/cosmos-sdk/codec"
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

func NewTxMatchFn(cdc codec.Codec, msgs []sdk.Msg, onlyFree bool) TxMatchFn {
	msgTypes := make([]reflect.Type, len(msgs))

	for i, msg := range msgs {
		msgTypes[i] = reflect.TypeOf(msg)
	}

	var matchMsgFn func(sdk.Context, sdk.Msg, codec.Codec) bool

	matchMsgFn = func(ctx sdk.Context, msg sdk.Msg, cdc codec.Codec) bool {
		msgExec, ok := msg.(*authz.MsgExec)
		if ok {
			subMsgs, err := msgExec.GetMessages()
			if err != nil {
				return false
			}
			for _, m := range subMsgs {
				if !matchMsgFn(ctx, m, cdc) {
					return false
				}
			}
			return true

		} else {
			return slices.Contains(msgTypes, reflect.TypeOf(msg))
		}
	}

	return func(ctx sdk.Context, tx sdk.Tx) bool {
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
			if !matchMsgFn(ctx, msg, cdc) {
				return false
			}
		}
		return true
	}
}
