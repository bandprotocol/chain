package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

const (
	EncoderProtoPrefix      = "\x01\xe2\xad\xb3" // tss.Hash([]byte("Proto"))[:4]
	EncoderFullABIPrefix    = "\x45\xb4\xe7\xea" // tss.Hash([]byte("FullABI"))[:4]
	EncoderPartialABIPrefix = "\x7b\xae\x7c\xd8" // tss.Hash([]byte("PartialABI"))[:4]
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

			switch c.Encoder {
			case types.ENCODER_PROTO:
				bz, err := k.MarshalResult(ctx, result)
				if err != nil {
					return nil, err
				}

				return append([]byte(EncoderProtoPrefix), bz...), nil
			case types.ENCODER_FULL_ABI:
				bz, err := result.PackFullABI()
				if err != nil {
					return nil, err
				}

				return append([]byte(EncoderFullABIPrefix), bz...), nil
			case types.ENCODER_PARTIAL_ABI:
				bz, err := result.PackPartialABI()
				if err != nil {
					return nil, err
				}

				return append([]byte(EncoderPartialABIPrefix), bz...), nil
			default:
				return nil, sdkerrors.ErrUnknownRequest.Wrapf(
					"unrecognized encoder type: %d",
					c.Encoder,
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
