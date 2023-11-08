package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// NewRequestingSignatureHandler creates a new TSS Handler for requesting the signature
func NewRequestingSignatureHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.OracleResultRequestingSignature:
			result, err := k.GetResult(ctx, c.RequestID)
			if err != nil {
				return nil, err
			}

			switch c.EncodeType {
			case types.ENCODE_TYPE_UNSPECIFIED, types.ENCODE_TYPE_PROTO:
				return k.MarshalResult(ctx, result)
			case types.ENCODE_TYPE_FULL_ABI:
				return result.PackFullABI()
			case types.ENCODE_TYPE_PARTIAL_ABI:
				return result.PackPartialABI()
			default:
				return nil, sdkerrors.Wrapf(
					sdkerrors.ErrUnknownRequest,
					"unrecognized encode type: %d",
					c.EncodeType,
				)
			}

		default:
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"unrecognized tss request signature type: %s",
				c.RequestingSignatureType(),
			)
		}
	}
}
