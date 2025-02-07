package band

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/app/mempool"
)

// isBankSendTx returns true if this transaction is strictly a bank send transaction (MsgSend).
func isBankSendTx(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*banktypes.MsgSend); !ok {
			return false
		}
	}
	return true
}

// isDelegateTx returns true if this transaction is strictly a staking delegate transaction (MsgDelegate).
func isDelegateTx(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		if _, ok := msg.(*stakingtypes.MsgDelegate); !ok {
			return false
		}
	}
	return true
}

// isOtherTx returns true if the transaction is neither a pure bank send nor a pure delegate transaction.
func isOtherTx(tx sdk.Tx) bool {
	return !isBankSendTx(tx) && !isDelegateTx(tx)
}

// BankSendLane returns a lane named "bankSend" that matches only MsgSend transactions,
// assigning 30% of the block's gas limit to this lane.
func BankSendLane() *mempool.Lane {
	return mempool.NewLane(
		"bankSend",
		isBankSendTx,
		30,    // percentage
		false, // EnforceOneTxPerSigner? set to true if you want one tx per signer
	)
}

// DelegateLane returns a lane named "delegate" that matches only MsgDelegate transactions,
// assigning 30% of the block's gas limit to this lane.
func DelegateLane() *mempool.Lane {
	return mempool.NewLane(
		"delegate",
		isDelegateTx,
		30,    // percentage
		false, // EnforceOneTxPerSigner? set to true if you want one tx per signer
	)
}

// OtherLane returns a lane named "other" for any transaction that does not strictly
// match isBankSendTx or isDelegateTx. It allocates 40% of the block's gas limit.
func OtherLane() *mempool.Lane {
	return mempool.NewLane(
		"other",
		isOtherTx,
		40,    // percentage
		false, // EnforceOneTxPerSigner? set to true if you want one tx per signer
	)
}

// DefaultLanes is a convenience helper returning the typical three lanes:
// bankSend, delegate, and other (30%, 30%, and 40%).
func DefaultLanes() []*mempool.Lane {
	return []*mempool.Lane{
		BankSendLane(),
		DelegateLane(),
		OtherLane(),
	}
}
