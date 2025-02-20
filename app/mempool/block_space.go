package mempool

import (
	"fmt"

	"cosmossdk.io/math"
)

// BlockSpace defines the block space.
type BlockSpace struct {
	txBytes int64
	gas     uint64
}

// NewBlockSpace returns a new block space.
func NewBlockSpace(txBytes int64, gas uint64) BlockSpace {
	return BlockSpace{
		txBytes: txBytes,
		gas:     gas,
	}
}

// --- Getters ---
func (bs BlockSpace) TxBytes() int64 {
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

// IncreaseBy increases this BlockSpace by another BlockSpace's size/gas.
func (bs *BlockSpace) IncreaseBy(other BlockSpace) {
	bs.txBytes += other.txBytes
	bs.gas += other.gas
}

// DecreaseBy decreases this BlockSpace by another BlockSpace's size/gas.
// Ensures txBytes and gas never go below zero.
func (bs *BlockSpace) DecreaseBy(other BlockSpace) {
	// Decrease txBytes
	if other.txBytes > bs.txBytes {
		bs.txBytes = 0
	} else {
		bs.txBytes -= other.txBytes
	}

	// Decrease gas
	if other.gas > bs.gas {
		bs.gas = 0
	} else {
		bs.gas -= other.gas
	}
}

// Sub returns the difference between this BlockSpace and another BlockSpace.
// Ensures txBytes and gas never go below zero.
func (bs BlockSpace) Sub(other BlockSpace) BlockSpace {
	// Calculate txBytes
	txBytes := bs.txBytes - other.txBytes
	if txBytes < 0 {
		txBytes = 0
	}

	// Calculate gas
	if other.gas > bs.gas {
		return BlockSpace{
			txBytes: txBytes,
			gas:     0,
		}
	}
	gas := bs.gas - other.gas

	return BlockSpace{
		txBytes: txBytes,
		gas:     gas,
	}
}

// Add returns the sum of this BlockSpace and another BlockSpace.
func (bs BlockSpace) Add(other BlockSpace) BlockSpace {
	return BlockSpace{
		txBytes: bs.txBytes + other.txBytes,
		gas:     bs.gas + other.gas,
	}
}

// MulDec returns a new BlockSpace with txBytes and gas multiplied by a decimal.
func (bs BlockSpace) MulDec(dec math.LegacyDec) BlockSpace {
	return BlockSpace{
		txBytes: dec.MulInt64(bs.txBytes).TruncateInt().Int64(),
		gas:     dec.MulInt(math.NewIntFromUint64(bs.gas)).TruncateInt().Uint64(),
	}
}

// --- Stringer ---

// String returns a string representation of the BlockSpace.
func (bs BlockSpace) String() string {
	return fmt.Sprintf("BlockSpace{txBytes: %d, gas: %d}", bs.txBytes, bs.gas)
}
