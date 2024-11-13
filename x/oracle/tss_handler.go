package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	tsslib "github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/oracle/keeper"
	"github.com/bandprotocol/chain/v3/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
)

var (
	// EncoderProtoPrefix is the prefix for proto encoding type
	// The value is tss.Hash([]byte("proto"))[:4]
	EncoderProtoPrefix = tsslib.Hash([]byte("proto"))[:4]
	// EncoderFullABIPrefix is the prefix for full ABI encoding type
	// The value is tss.Hash([]byte("fullABI"))[:4]
	EncoderFullABIPrefix = tsslib.Hash([]byte("fullABI"))[:4]
	// EncoderPartialABIPrefix is the prefix for partial ABI encoding type
	// The value is tss.Hash([]byte("partialABI"))[:4]
	EncoderPartialABIPrefix = tsslib.Hash([]byte("partialABI"))[:4]
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

				return append(EncoderProtoPrefix, bz...), nil
			case types.ENCODER_FULL_ABI:
				bz, err := result.PackFullABI()
				if err != nil {
					return nil, err
				}

				return append(EncoderFullABIPrefix, bz...), nil
			case types.ENCODER_PARTIAL_ABI:
				bz, err := result.PackPartialABI()
				if err != nil {
					return nil, err
				}

				return append(EncoderPartialABIPrefix, bz...), nil
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
