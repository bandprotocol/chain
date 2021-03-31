package types

import (
	"github.com/GeoDB-Limited/odin-core/x/common/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO move to ModuleCdc or cdc
type CoinProto sdk.Coin

func NewCoinProto(denom types.Denom) CoinProto {
	return CoinProto(sdk.NewInt64Coin(denom.String(), 0))
}

func (p CoinProto) Value() sdk.Coin {
	return sdk.Coin(p)
}

func (p CoinProto) Size() int {
	return len(amino.MustMarshalJSON(p))
}

func (p CoinProto) MarshalTo(dst []byte) (int, error) {
	res, err := amino.MarshalJSON(p)
	if err != nil {
		return 0, err
	}
	copy(dst, res)
	return len(dst), nil
}

func (p CoinProto) Unmarshal(src []byte) error {
	return amino.UnmarshalJSON(src, &p)
}

func (p CoinProto) Equal(other CoinProto) bool {
	return p.Denom == other.Denom && p.Amount.Equal(other.Amount)
}
