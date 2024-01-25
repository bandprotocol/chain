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

func (ps *PriceService) Validate() error {
	return nil
}
