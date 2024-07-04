package types

var _ Route = &TSSRoute{}

func (r *TSSRoute) ValidateBasic() error {
	return nil
}
