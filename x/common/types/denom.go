package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

type Denom string

func ParseDenom(rawDenom string) (Denom, error) {
	if err := sdk.ValidateDenom(rawDenom); err != nil {
		return "", err
	}
	return Denom(rawDenom), nil
}

func (d Denom) Equal(other Denom) bool {
	return d.Base() == other.Base()
}

func (d Denom) Base() string {
	return strings.ToLower(d.String())
}

func (d Denom) String() string {
	return string(d)
}

func (d Denom) IsEmpty() bool {
	return len(d) == 0
}
