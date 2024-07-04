package types

var _ Route = &AxelarRoute{}

func (r *AxelarRoute) ValidateBasic() error {
	return nil
}
