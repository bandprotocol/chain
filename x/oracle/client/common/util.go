package common

func ValueOrDefault(val string, def interface{}) interface{} {
	if val == "" {
		return def
	}
	return val
}
