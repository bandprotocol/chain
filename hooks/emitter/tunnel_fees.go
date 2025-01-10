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

// getTunnelSenderFeesMap returns the fees for each sender in event transfers.
func getTunnelSenderFeesMap(ctx sdk.Context, hook Hook, events []abci.Event) map[string]Fees {
	senderFeesMap := make(map[string]Fees)
	for _, event := range events {
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

			tunnelModuleAcc := hook.accountKeeper.GetModuleAccount(ctx, tunneltypes.ModuleName)

			fees := senderFeesMap[sender]

			if recipient == tunnelModuleAcc.GetAddress().String() {
				fees.BaseFee = fees.BaseFee.Add(amount...)
			} else {
				fees.RouteFee = fees.RouteFee.Add(amount...)
			}

			senderFeesMap[sender] = fees
		}
	}

	return senderFeesMap
}
