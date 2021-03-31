package types

import "github.com/GeoDB-Limited/odin-core/x/common/types"

func (v ValidExchanges) Contains(from types.Denom, to types.Denom) bool {
	exchanges, ok := v.Exchanges[from.String()]
	if !ok {
		return false
	}
	for _, e := range exchanges.Value {
		if types.Denom(e).Equal(to) {
			return true
		}
	}
	return false
}
