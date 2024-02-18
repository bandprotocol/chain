package types

// NewParams creates a new Params instance
func NewParams() Params {
	return Params{}
}

// DefaultParams returns default parameters
func DefaultParams() Params {
	return Params{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}
