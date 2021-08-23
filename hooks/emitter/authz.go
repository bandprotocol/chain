package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v2/hooks/common"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

func (h *Hook) handleMsgGrant(msg *authz.MsgGrant, detail common.JsDict) {
	authorization := msg.Grant.GetAuthorization()
	switch authorization.MsgTypeURL() {
	case oracletypes.ReportAuthorization{}.MsgTypeURL():
		acc, _ := sdk.AccAddressFromBech32(msg.Granter)
		val := sdk.ValAddress(acc).String()
		h.Write("SET_REPORTER", common.JsDict{
			"reporter":  msg.Grantee,
			"validator": val,
		})
		detail["validator_moniker"] = val
	default:
		break
	}
	h.AddAccountsInTx(msg.Grantee)
	detail["url"] = msg.Grant.Authorization.GetTypeUrl()
}

func (h *Hook) handleMsgRevoke(msg *authz.MsgRevoke, detail common.JsDict) {
	switch msg.MsgTypeUrl {
	case oracletypes.ReportAuthorization{}.MsgTypeURL():
		acc, _ := sdk.AccAddressFromBech32(msg.Granter)
		val := sdk.ValAddress(acc).String()
		h.Write("REMOVE_REPORTER", common.JsDict{
			"reporter":  msg.Grantee,
			"validator": val,
		})
		detail["validator_moniker"] = val
	default:
		break
	}
	h.AddAccountsInTx(msg.Grantee)
	detail["url"] = msg.MsgTypeUrl
}

func (h *Hook) handleMsgExec(ctx sdk.Context, txHash []byte, msg *authz.MsgExec, detail common.JsDict) {
	msgs, _ := msg.GetMessages()
	grantee := msg.Grantee
	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *oracletypes.MsgReportData:
			h.handleMsgReportDataFromGrantee(ctx, txHash, msg, grantee)
		default:
			break
		}
	}
}
