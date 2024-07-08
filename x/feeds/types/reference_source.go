package types

// NewReferenceSourceConfig creates a new reference source config instance
func NewReferenceSourceConfig(hash string, version string) ReferenceSourceConfig {
	return ReferenceSourceConfig{
		IPFSHash: hash,
		Version:  version,
	}
}

// DefaultReferenceSourceConfig returns a default set of reference source config's information
func DefaultReferenceSourceConfig() ReferenceSourceConfig {
	return NewReferenceSourceConfig("hash", "0.0.1")
}

// Validate validates the reference source config
func (ps *ReferenceSourceConfig) Validate() error {
	if err := validateString("hash", false, ps.IPFSHash); err != nil {
		return err
	}

	if err := validateString("version", false, ps.Version); err != nil {
		return err
	}

	return nil
}