package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v2/x/oracle/keeper"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

// NewWhiteListAnteHandler returns a new ante handler that filter requests from external addresses out
func NewWhiteListAnteHandler(ante sdk.AnteHandler, oracleKeeper keeper.Keeper, requesters []string) sdk.AnteHandler {
	whiteList := make(map[string]bool)
	for _, addr := range requesters {
		whiteList[addr] = true
	}

	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		if ctx.IsCheckTx() && !simulate {
			for _, msg := range tx.GetMsgs() {
				switch m := msg.(type) {
				case *types.MsgRequestData:
					if _, found := whiteList[m.Sender]; !found {
						return ctx, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("%s not in the whitelist", m.Sender))
					}
				case *types.MsgReportData:
					// TODO: check if this is our report
					continue
				case *authz.MsgExec:
					execMsgs, err := m.GetMessages()
					if err != nil {
						return ctx, err
					}

					for _, execMsg := range execMsgs {
						if sdk.MsgTypeURL(&types.MsgReportData{}) != sdk.MsgTypeURL(execMsg) {
							return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "Msg type is not allowed")
						}
					}
				default:
					// reject all other msg type
					return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidType, "Msg type is not allowed")
				}
			}
		}
		return ante(ctx, tx, simulate)
	}
}
