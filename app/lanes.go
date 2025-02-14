package band

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"

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

// BandLanes returns the default lanes for the Band Protocol blockchain.
func BandLanes(app *BandApp) []*mempool.Lane {
	// 1. Create the signer extractor. This is used to extract the expected signers from
	// a transaction. Each lane can have a different signer extractor if needed.
	signerAdapter := signerextraction.NewDefaultAdapter()

	BankSendLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"bankSend",
		isBankSendTx,
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	DelegateLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"delegate",
		isDelegateTx,
		math.LegacyMustNewDecFromStr("0.05"),
		math.LegacyMustNewDecFromStr("0.3"),
		sdkmempool.DefaultPriorityMempool(),
	)

	OtherLane := mempool.NewLane(
		app.Logger(),
		app.txConfig.TxEncoder(),
		signerAdapter,
		"other",
		isOtherTx,
		math.LegacyMustNewDecFromStr("0.1"),
		math.LegacyMustNewDecFromStr("0.4"),
		sdkmempool.DefaultPriorityMempool(),
	)

	return []*mempool.Lane{BankSendLane, DelegateLane, OtherLane}
}
