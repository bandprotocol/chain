package round1

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

type Round1 struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round1{}

func New(ctx *cylinder.Context) (*Round1, error) {
	// create http client
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Round1{
		context: ctx,
		logger:  ctx.Logger.With("worker", "round1"),
		client:  cli,
	}, nil
}

func (r *Round1) subscribe() error {
	var err error
	r.eventCh, err = r.client.Subscribe(
		"round1",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s EXISTS AND %s.%s = '%s'",
			types.EventTypeCreateGroup,
			types.AttributeKeyGroupID,
			types.EventTypeCreateGroup,
			types.AttributeKeyMember,
			r.context.Config.Granter,
		),
		1000,
	)
	return err
}

func (r *Round1) handleTxResult(txResult abci.TxResult) {
	msgLogs, err := event.GetMessageLogs(txResult)
	if err != nil {
		r.logger.Error("Failed to get message logs: %s", err.Error())
		return
	}

	for _, log := range msgLogs {
		event, err := ParseEvent(log)
		if err != nil {
			r.logger.Error(":cold_sweat: Failed to parse event with error: %s", err.Error())
			return
		}

		go r.handleEvent(event)
	}
}

func (r *Round1) handleEvent(event *Event) {
	logger := r.logger.With("gid", event.GroupID)
	logger.Info(":delivery_truck: Processing incoming group event")

	var mid types.MemberID
	for idx, member := range event.Members {
		if member == r.context.Config.Granter {
			mid = types.MemberID(idx)
		}
	}

	data, err := tss.GenerateRound1Data(event.GroupID, mid, event.Threshold, event.DKGContext)
	if err != nil {
		logger.Error(":cold_sweat: Failed to generate round1 data with error: %s", err.Error())
		return
	}

	// set group data
	r.context.Store.SetGroup(event.GroupID, store.Group{
		MemberID:       mid,
		Coefficients:   data.Coefficients,
		OneTimePrivKey: data.OneTimePrivKey,
	})

	// generate message
	msg := &types.MsgSubmitDKGRound1{
		GroupID:            event.GroupID,
		MemberID:           mid,
		CoefficientsCommit: data.CoefficientsCommit,
		OneTimePubKey:      data.OneTimePubKey,
		A0Sig:              data.A0Sig,
		OneTimeSig:         data.OneTimeSig,
		Member:             r.context.Config.Granter,
	}

	r.context.MsgCh <- msg
}

func (r *Round1) Start() {
	r.logger.Info("start")

	err := r.subscribe()
	if err != nil {
		r.context.ErrCh <- err
		return
	}

	for {
		ev := <-r.eventCh
		go r.handleTxResult(ev.Data.(tmtypes.EventDataTx).TxResult)
	}
}

func (r *Round1) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
