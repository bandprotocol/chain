package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialMintPool returns the initial state of MintPool
func InitialMintPool() MintPool {
	return MintPool{
		TreasuryPool:         sdk.Coins{},
		EligibleAccountsPool: []string{},
	}
}

// ValidateGenesis validates the mint pool for a genesis state
func (m MintPool) ValidateGenesis() error {
	if m.TreasuryPool.IsAnyNegative() {
		return fmt.Errorf("negative TreasuryPool in mint pool, is %v", m.TreasuryPool)
	}

	return nil
}

// AddrPool defines a pool of addresses
type AddrPool []sdk.AccAddress

// Contains checks id addr exists in the slice
func (p AddrPool) Contains(addr sdk.AccAddress) bool {
	for _, item := range p {
		if item.Equals(addr) {
			return true
		}
	}
	return false
}
