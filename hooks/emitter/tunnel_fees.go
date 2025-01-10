package emitter

import (
	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
)

// Fees stores the base fee and route fee for each sender.
type Fees struct {
	BaseFee  sdk.Coins
	RouteFee sdk.Coins
}

// TunnelFees stores the fees for each sender in event transfers.
type TunnelFees struct {
	Ctx           sdk.Context
	Hook          Hook
	Events        []abci.Event
	SenderFeesMap map[string]Fees
}

// NewTunnelFees creates a new TunnelFees instance
func NewTunnelFees(ctx sdk.Context, hook Hook, events []abci.Event) *TunnelFees {
	return &TunnelFees{
		Ctx:           ctx,
		Hook:          hook,
		Events:        events,
		SenderFeesMap: make(map[string]Fees),
	}
}

// CalculateFee calculates the base fee and route fee for each sender from event transfers.
// This function is called at the end of the block, specifically for use in the tunnel module.
func (tf *TunnelFees) CalculateFee() {
	for _, event := range tf.Events {
		if event.Type == banktypes.EventTypeTransfer {
			evMap := parseEvents([]abci.Event{event})

			sender := evMap[banktypes.EventTypeTransfer+"."+banktypes.AttributeKeySender][0]
			recipient := evMap[banktypes.EventTypeTransfer+"."+banktypes.AttributeKeyRecipient][0]

			amount, err := sdk.ParseCoinsNormalized(
				evMap[banktypes.EventTypeTransfer+"."+sdk.AttributeKeyAmount][0],
			)
			if err != nil {
				continue
			}

			tunnelModuleAcc := tf.Hook.accountKeeper.GetModuleAccount(tf.Ctx, tunneltypes.ModuleName)

			fees, found := tf.SenderFeesMap[sender]
			if !found {
				fees = Fees{}
			}

			if recipient == tunnelModuleAcc.GetAddress().String() {
				fees.BaseFee = fees.BaseFee.Add(amount...)
			} else {
				fees.RouteFee = fees.RouteFee.Add(amount...)
			}

			tf.SenderFeesMap[sender] = fees
		}
	}
}

// GetFees returns the fees for the given fee payer.
func (tf *TunnelFees) GetFees(feePayer string) (Fees, bool) {
	fees, found := tf.SenderFeesMap[feePayer]
	return fees, found
}
