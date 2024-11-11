package types

const (
	// MaxSignalIDCharacters defines the maximum number of characters allowed in a signal ID.
	MaxSignalIDCharacters uint64 = 32

	// MaxGuaranteeBlockTime specifies the maximum capped block time (in seconds) during a grace period.
	// If block times are slower, they will be capped at this value to prevent validator deactivation,
	// as long as the block height remains within the calculated threshold for MaxGuaranteeBlockTime.
	MaxGuaranteeBlockTime int64 = 3
)
