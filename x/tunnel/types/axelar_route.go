package types

var _ RouteI = &AxelarRoute{}

func (r *AxelarRoute) ValidateBasic() error {
	return nil
}
