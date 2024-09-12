package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Deposit represents a deposit from a user to a tunnel.
func NewDeposit(tunnelID uint64, depositor sdk.AccAddress, amount sdk.Coins) Deposit {
	return Deposit{
		TunnelID:  tunnelID,
		Depositor: depositor.String(),
		Amount:    amount,
	}
}
