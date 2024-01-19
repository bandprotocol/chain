package types

// NewOffChain creates a new OffChain instance
func NewOffChain(hash string, version string, url string) OffChain {
	return OffChain{
		Hash:    hash,
		Version: version,
		Url:     url,
	}
}

// DefaultOffChain returns a default set of off-chain information
func DefaultOffChain() OffChain {
	return NewOffChain("hash", "0.0.1", "https://")
}

func (o *OffChain) Validate() error {
	return nil
}
