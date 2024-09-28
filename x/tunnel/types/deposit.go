package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// NewDeposit creates a new Deposit instance
func NewDeposit(tunnelID uint64, depositor sdk.AccAddress, amount sdk.Coins) Deposit {
	return Deposit{
		TunnelID:  tunnelID,
		Depositor: depositor.String(),
		Amount:    amount,
	}
}
