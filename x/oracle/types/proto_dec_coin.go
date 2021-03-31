package types

import (
	"github.com/GeoDB-Limited/odin-core/x/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CoinDecProto sdk.DecCoin

func NewCoinDecProto(denom types.Denom) CoinDecProto {
	return CoinDecProto(sdk.NewInt64DecCoin(denom.String(), 0))
}

func (p CoinDecProto) Value() sdk.DecCoin {
	return sdk.DecCoin(p)
}

func (p CoinDecProto) Size() int {
	return len(amino.MustMarshalJSON(p))
}

func (p CoinDecProto) MarshalTo(dst []byte) (int, error) {
	res, err := amino.MarshalJSON(p)
	if err != nil {
		return 0, err
	}
	copy(dst, res)
	return len(dst), nil
}

func (p CoinDecProto) Unmarshal(src []byte) error {
	return amino.UnmarshalJSON(src, &p)
}

func (p CoinDecProto) Equal(other CoinDecProto) bool {
	return p.Denom == other.Denom && p.Amount.Equal(other.Amount)
}
