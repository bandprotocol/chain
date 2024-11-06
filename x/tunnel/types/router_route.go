package types

// RouterRoute defines the Router route for the tunnel module
var _ RouteI = &RouterRoute{}

// RouterRoute defines the Router route for the tunnel module
func (r *RouterRoute) ValidateBasic() error {
	return nil
}
