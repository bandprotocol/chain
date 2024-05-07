package types

// NewPriceService creates a new price service instance
func NewPriceService(hash string, version string, url string) PriceService {
	return PriceService{
		Hash:    hash,
		Version: version,
		Url:     url,
	}
}

// DefaultPriceService returns a default set of price service's information
func DefaultPriceService() PriceService {
	return NewPriceService("hash", "0.0.1", "https://")
}

// Validate validates the price service
func (ps *PriceService) Validate() error {
	if err := validateString("hash", false)(ps.Hash); err != nil {
		return err
	}

	if err := validateString("version", false)(ps.Version); err != nil {
		return err
	}

	if err := validateString("url", false)(ps.Url); err != nil {
		return err
	}

	if err := validateURL("url")(ps.Url); err != nil {
		return err
	}

	return nil
}
