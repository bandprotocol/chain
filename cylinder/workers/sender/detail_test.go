package sender_test

import (
	"testing"

	"github.com/bandprotocol/chain/v2/cylinder/workers/sender"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestGetDetail(t *testing.T) {
	tests := []struct {
		name   string
		msgs   []sdk.Msg
		expect []sender.Detail
	}{
		{
			"no msg",
			[]sdk.Msg{},
			[]sender.Detail{},
		},
		{
			"one msg",
			[]sdk.Msg{&types.MsgSubmitDKGRound1{
				GroupID: 1,
			}},
			[]sender.Detail{
				{
					Type:    "/tss.v1beta1.MsgSubmitDKGRound1",
					GroupID: 1,
				},
			},
		},
		{
			"two msgs with the same order",
			[]sdk.Msg{
				&types.MsgSubmitDKGRound1{
					GroupID: 1,
				},
				&types.MsgSubmitDKGRound1{
					GroupID: 2,
				},
			},
			[]sender.Detail{
				{
					Type:    "/tss.v1beta1.MsgSubmitDKGRound1",
					GroupID: 1,
				},
				{
					Type:    "/tss.v1beta1.MsgSubmitDKGRound1",
					GroupID: 2,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			details := sender.GetDetail(test.msgs)
			assert.Equal(t, test.expect, details)
		})
	}
}
