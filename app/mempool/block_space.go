package mempool

import (
	"fmt"
	"math"

	ethmath "github.com/ethereum/go-ethereum/common/math"

	sdkmath "cosmossdk.io/math"
)

// BlockSpace defines the block space.
type BlockSpace struct {
	txBytes uint64
	gas     uint64
}

// NewBlockSpace returns a new block space.
func NewBlockSpace(txBytes uint64, gas uint64) BlockSpace {
	return BlockSpace{
		txBytes: txBytes,
		gas:     gas,
	}
}

// --- Getters ---
func (bs BlockSpace) TxBytes() uint64 {
	return bs.txBytes
}

func (bs BlockSpace) Gas() uint64 {
	return bs.gas
}

// --- Comparison Methods ---

// IsReachedBy returns true if 'other' usage has reached this BlockSpace's limits.
func (bs BlockSpace) IsReachedBy(other BlockSpace) bool {
	return other.txBytes >= bs.txBytes || other.gas >= bs.gas
}

// IsExceededBy returns true if 'other' usage has exceeded this BlockSpace's limits.
func (bs BlockSpace) IsExceededBy(other BlockSpace) bool {
	return other.txBytes > bs.txBytes || other.gas > bs.gas
}

// --- Math Methods ---

// Sub returns the difference between this BlockSpace and another BlockSpace.
// Ensures txBytes and gas never go below zero.
func (bs BlockSpace) Sub(other BlockSpace) BlockSpace {
	var txBytes uint64
	var gas uint64

	// Calculate txBytes
	txBytes, borrowOut := ethmath.SafeSub(bs.txBytes, other.txBytes)
	if borrowOut {
		txBytes = 0
	}

	// Calculate gas
	gas, borrowOut = ethmath.SafeSub(bs.gas, other.gas)
	if borrowOut {
		gas = 0
	}

	return BlockSpace{
		txBytes: txBytes,
		gas:     gas,
	}
}

// Add returns the sum of this BlockSpace and another BlockSpace.
func (bs BlockSpace) Add(other BlockSpace) BlockSpace {
	var txBytes uint64
	var gas uint64

	// Calculate txBytes
	txBytes, carry := ethmath.SafeAdd(bs.txBytes, other.txBytes)
	if carry {
		txBytes = math.MaxUint64
	}

	// Calculate gas
	gas, carry = ethmath.SafeAdd(bs.gas, other.gas)
	if carry {
		gas = math.MaxUint64
	}

	return BlockSpace{
		txBytes: txBytes,
		gas:     gas,
	}
}

// Scale returns a new BlockSpace with txBytes and gas multiplied by a decimal.
func (bs BlockSpace) Scale(dec sdkmath.LegacyDec) (BlockSpace, error) {
	txBytes := dec.MulInt(sdkmath.NewIntFromUint64(bs.txBytes)).TruncateInt()
	gas := dec.MulInt(sdkmath.NewIntFromUint64(bs.gas)).TruncateInt()

	if !txBytes.IsUint64() || !gas.IsUint64() {
		return BlockSpace{}, fmt.Errorf("block space scaling overflow: block_space %s, dec %s", bs, dec)
	}

	return BlockSpace{
		txBytes: txBytes.Uint64(),
		gas:     gas.Uint64(),
	}, nil
}

// --- Stringer ---

// String returns a string representation of the BlockSpace.
func (bs BlockSpace) String() string {
	return fmt.Sprintf("BlockSpace{txBytes: %d, gas: %d}", bs.txBytes, bs.gas)
}
