package types

var _ RouteI = &TSSRoute{}

func (r *TSSRoute) ValidateBasic() error {
	return nil
}
