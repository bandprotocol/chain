package emitter

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	"github.com/bandprotocol/chain/v2/hooks/common"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
)

func (h *Hook) handleMsgGrant(msg *authz.MsgGrant, detail common.JsDict) {
	authorization, _ := msg.Grant.GetAuthorization()
	switch authorization.MsgTypeURL() {
	case sdk.MsgTypeURL(&oracletypes.MsgReportData{}):
		acc, _ := sdk.AccAddressFromBech32(msg.Granter)
		val := sdk.ValAddress(acc).String()
		h.Write("SET_REPORTER", common.JsDict{
			"reporter":  msg.Grantee,
			"validator": val,
		})
		detail["validator_moniker"] = val
	case sdk.MsgTypeURL((&feedstypes.MsgSubmitSignalPrices{})):
		acc, _ := sdk.AccAddressFromBech32(msg.Granter)
		val := sdk.ValAddress(acc).String()
		h.Write("SET_FEEDER", common.JsDict{
			"feeder":    msg.Grantee,
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
	case sdk.MsgTypeURL(&oracletypes.MsgReportData{}):
		acc, _ := sdk.AccAddressFromBech32(msg.Granter)
		val := sdk.ValAddress(acc).String()
		h.Write("REMOVE_REPORTER", common.JsDict{
			"reporter":  msg.Grantee,
			"validator": val,
		})
		detail["validator_moniker"] = val
	case sdk.MsgTypeURL(&feedstypes.MsgSubmitSignalPrices{}):
		acc, _ := sdk.AccAddressFromBech32(msg.Granter)
		val := sdk.ValAddress(acc).String()
		h.Write("REMOVE_FEEDER", common.JsDict{
			"feeder":    msg.Grantee,
			"validator": val,
		})
		detail["validator_moniker"] = val
	default:
		break
	}
	h.AddAccountsInTx(msg.Grantee)
	detail["url"] = msg.MsgTypeUrl
}

func (h *Hook) handleMsgExec(
	ctx sdk.Context,
	txHash []byte,
	emsg *authz.MsgExec,
	log sdk.ABCIMessageLog,
	detail common.JsDict,
) {
	msgs, _ := emsg.GetMessages()
	grantee := emsg.Grantee

	// If cannot cast or invalid length it will panic and fix later
	subMsgs := detail["msgs"].([]common.JsDict)
	for i, msg := range msgs {
		switch msg := msg.(type) {
		case *oracletypes.MsgReportData:
			h.handleMsgReportData(ctx, txHash, msg, grantee)
		default:
			// add signers for this message into the transaction
			signers := msg.GetSigners()
			addrs := make([]string, len(signers))
			for idx, signer := range signers {
				addrs[idx] = signer.String()
			}
			h.AddAccountsInTx(addrs...)
			h.handleMsg(ctx, txHash, msg, log, subMsgs[i]["msg"].(common.JsDict))
		}
	}
}
