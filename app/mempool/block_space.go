package mempool

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

// IsReached returns true if the block space is reached.
func (bs BlockSpace) IsReached(size int64, gas uint64) bool {
	return size >= bs.txBytes || gas >= bs.gas
}

// IsExceeded returns true if the block space is exceeded.
func (bs BlockSpace) IsExceeded(size int64, gas uint64) bool {
	return size > bs.txBytes || gas > bs.gas
}

// Decrease decreases the block space.
func (bs *BlockSpace) DecreaseBy(size int64, gas uint64) {
	bs.txBytes -= size
	if bs.txBytes < 0 {
		bs.txBytes = 0
	}
	bs.gas -= gas
}
