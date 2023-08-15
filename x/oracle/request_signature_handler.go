package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

// NewRequestSignatureHandler creates a new TSS Handler for requesting the signature
func NewRequestSignatureHandler(k keeper.Keeper) tsstypes.Handler {
	return func(ctx sdk.Context, content tsstypes.Content) ([]byte, error) {
		switch c := content.(type) {
		case *types.OracleResultRequestSignature:
			return handleRequestSignatureByRequestID(ctx, k, c)

		default:
			return nil, sdkerrors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"unrecognized tss request signature type: %s",
				c.RequestSignatureType(),
			)
		}
	}
}

func handleRequestSignatureByRequestID(
	ctx sdk.Context,
	k keeper.Keeper,
	rs *types.OracleResultRequestSignature,
) ([]byte, error) {
	r, err := k.GetResult(ctx, rs.RequestID)
	if err != nil {
		return nil, err
	}
	return r.Result, nil
}
