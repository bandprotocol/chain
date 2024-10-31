package types

// NewReferenceSourceConfig creates a new reference source config instance
func NewReferenceSourceConfig(hash string, version string) ReferenceSourceConfig {
	return ReferenceSourceConfig{
		RegistryIPFSHash: hash,
		RegistryVersion:  version,
	}
}

// DefaultReferenceSourceConfig returns a default set of reference source config's information
func DefaultReferenceSourceConfig() ReferenceSourceConfig {
	return NewReferenceSourceConfig("[NOT_SET]", "[NOT_SET]")
}

// Validate validates the reference source config
func (ps *ReferenceSourceConfig) Validate() error {
	if err := validateString("registry ipfs hash", false, ps.RegistryIPFSHash); err != nil {
		return err
	}

	if err := validateString("registry version", false, ps.RegistryVersion); err != nil {
		return err
	}

	if err := validateVersion("registry version", ps.RegistryVersion); err != nil {
		return err
	}

	return nil
}
