package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialOraclePool returns zero oracle pool
func InitialOraclePool() OraclePool {
	return OraclePool{
		DataProvidersPool: sdk.Coins{},
	}
}

// ValidateGenesis validates the oracle pool for a genesis state
func (f OraclePool) ValidateGenesis() error {
	if f.DataProvidersPool.IsAnyNegative() {
		return fmt.Errorf("negative DataProvidersPool in oracle pool, is %v",
			f.DataProvidersPool)
	}

	return nil
}
