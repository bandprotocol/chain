package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialMintPool returns the initial state of MintPool
func InitialMintPool() MintPool {
	return MintPool{
		TreasuryPool: sdk.Coins{},
	}
}

// ValidateGenesis validates the mint pool for a genesis state
func (m MintPool) ValidateGenesis() error {
	if m.TreasuryPool.IsAnyNegative() {
		return fmt.Errorf("negative TreasuryPool in mint pool, is %v", m.TreasuryPool)
	}

	return nil
}
