package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewDeposit creates a new Deposit instance
func NewDeposit(tunnelID uint64, depositor string, amount sdk.Coins) Deposit {
	return Deposit{
		TunnelID:  tunnelID,
		Depositor: depositor,
		Amount:    amount,
	}
}
