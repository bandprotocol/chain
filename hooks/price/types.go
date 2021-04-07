package price

type Input struct {
	Symbols    []string `json:"symbols"`
	Multiplier uint64   `json:"multiplier"`
}

type Output struct {
	Rates []uint64 `json:"rates"`
}
