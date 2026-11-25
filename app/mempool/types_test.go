package mempool

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	protov2 "google.golang.org/protobuf/proto"
)

type laneMatchTestTx struct {
	msgs []sdk.Msg
}

func (tx laneMatchTestTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

func (laneMatchTestTx) GetMsgsV2() ([]protov2.Message, error) {
	return nil, nil
}

func TestNewLaneTxMatchFnWithPredicates(t *testing.T) {
	matcher := NewLaneTxMatchFnWithPredicates(
		[]sdk.Msg{&channeltypes.MsgRecvPacket{}},
		false,
		MsgMatchPredicate{
			Msg: &channeltypes.MsgRecvPacket{},
			Predicate: func(msg sdk.Msg) bool {
				return msg.(*channeltypes.MsgRecvPacket).Packet.DestinationPort == "oracle"
			},
		},
	)

	for _, tc := range []struct {
		name        string
		destination string
		wantMatch   bool
	}{
		{name: "oracle port", destination: "oracle", wantMatch: true},
		{name: "other port", destination: "transfer", wantMatch: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := laneMatchTestTx{msgs: []sdk.Msg{
				&channeltypes.MsgRecvPacket{
					Packet: channeltypes.Packet{DestinationPort: tc.destination},
				},
			}}
			require.Equal(t, tc.wantMatch, matcher(sdk.Context{}, tx))
		})
	}
}

func TestNewLaneTxMatchFnWithPredicatesRecursesIntoAuthz(t *testing.T) {
	matcher := NewLaneTxMatchFnWithPredicates(
		[]sdk.Msg{&channeltypes.MsgRecvPacket{}},
		false,
		MsgMatchPredicate{
			Msg: &channeltypes.MsgRecvPacket{},
			Predicate: func(msg sdk.Msg) bool {
				return msg.(*channeltypes.MsgRecvPacket).Packet.DestinationPort == "oracle"
			},
		},
	)

	msgExec := authz.NewMsgExec(sdk.AccAddress("grantee"), []sdk.Msg{
		&channeltypes.MsgRecvPacket{
			Packet: channeltypes.Packet{DestinationPort: "oracle"},
		},
	})

	require.True(t, matcher(sdk.Context{}, laneMatchTestTx{msgs: []sdk.Msg{&msgExec}}))
}
