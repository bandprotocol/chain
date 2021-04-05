package types

func (v ValidExchanges) Contains(from string, to string) bool {
	exchanges, ok := v.Exchanges[from]
	if !ok {
		return false
	}
	for _, e := range exchanges.Value {
		if e == to {
			return true
		}
	}
	return false
}
