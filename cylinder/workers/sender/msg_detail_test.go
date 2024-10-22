package sender_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder/workers/sender"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestGetMsgDetails(t *testing.T) {
	tests := []struct {
		name   string
		msgs   []sdk.Msg
		expect []string
	}{
		{
			"no msg",
			[]sdk.Msg{},
			nil,
		},
		{
			"one msg",
			[]sdk.Msg{&types.MsgSubmitDKGRound1{
				GroupID: 1,
			}},
			[]string{"Type: /tss.v1beta1.MsgSubmitDKGRound1, GroupID: 1"},
		},
		{
			"multiple messages",
			[]sdk.Msg{
				&types.MsgSubmitDKGRound1{
					GroupID: 1,
				},
				&types.MsgSubmitDKGRound2{
					GroupID: 2,
				},
				&types.MsgConfirm{
					GroupID: 3,
				},
				&types.MsgComplain{
					GroupID: 4,
				},
				&types.MsgSubmitDEs{},
				&types.MsgSubmitSignature{
					SigningID: 1,
				},
			},
			[]string{
				"Type: /tss.v1beta1.MsgSubmitDKGRound1, GroupID: 1",
				"Type: /tss.v1beta1.MsgSubmitDKGRound2, GroupID: 2",
				"Type: /tss.v1beta1.MsgConfirm, GroupID: 3",
				"Type: /tss.v1beta1.MsgComplain, GroupID: 4",
				"Type: /tss.v1beta1.MsgSubmitDEs",
				"Type: /tss.v1beta1.MsgSubmitSignature, SigningID: 1",
			},
		},
		{
			"unknown msg",
			[]sdk.Msg{
				&oracletypes.MsgRequestData{},
			},
			[]string{"Type: Unknown"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			details := sender.GetMsgDetails(test.msgs...)
			assert.Equal(t, test.expect, details)
		})
	}
}
