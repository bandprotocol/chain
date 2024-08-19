package types

// IBCRoute defines the IBC route for the tunnel module
var _ RouteI = &IBCRoute{}

// Route defines the IBC route for the tunnel module
func (r *IBCRoute) ValidateBasic() error {
	return nil
}
