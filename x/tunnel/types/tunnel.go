package types

import (
	"encoding/json"
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	proto "github.com/cosmos/gogoproto/proto"
)

func (t *Tunnel) SetRoute(route Route) error {
	msg, ok := route.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	t.Route = any

	return nil
}

func (t Tunnel) UnpackRoute() (TSSRoute, error) {
	var route TSSRoute
	fmt.Printf("route: %+v\n", t.Route)
	fmt.Printf("route: %+v\n", t.Route.GetValue())
	t.Route.GetCachedValue()
	err := json.Unmarshal(t.Route.GetValue(), &route)
	if err != nil {
		return TSSRoute{}, err
	}
	fmt.Printf("route: %+v\n", route)
	return route, nil
}
