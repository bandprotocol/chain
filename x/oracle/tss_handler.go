package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tsslib "github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

var (
	EncodeTypeProtoPrefix      = tsslib.Hash([]byte("proto"))[:4]
	EncodeTypeFullABIPrefix    = tsslib.Hash([]byte("fullABI"))[:4]
	EncodeTypePartialABIPrefix = tsslib.Hash([]byte("partialABI"))[:4]
)

// NewSignatureOrderHandler creates a tss handler to handle oracle result signature order
func NewSignatureOrderHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.OracleResultSignatureOrder:
			result, err := k.GetResult(ctx, c.RequestID)
			if err != nil {
				return nil, err
			}

			switch c.EncodeType {
			case types.ENCODE_TYPE_PROTO:
				bz, err := k.MarshalResult(ctx, result)
				if err != nil {
					return nil, err
				}

				return append(EncodeTypeProtoPrefix, bz...), nil
			case types.ENCODE_TYPE_FULL_ABI:
				bz, err := result.PackFullABI()
				if err != nil {
					return nil, err
				}

				return append(EncodeTypeFullABIPrefix, bz...), nil
			case types.ENCODE_TYPE_PARTIAL_ABI:
				bz, err := result.PackPartialABI()
				if err != nil {
					return nil, err
				}

				return append(EncodeTypePartialABIPrefix, bz...), nil
			default:
				return nil, sdkerrors.ErrUnknownRequest.Wrapf(
					"unrecognized encode type: %d",
					c.EncodeType,
				)
			}

		default:
			return nil, sdkerrors.ErrUnknownRequest.Wrapf(
				"unrecognized tss request signature type: %s",
				c.OrderType(),
			)
		}
	}
}
