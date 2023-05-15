package round2

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	abci "github.com/tendermint/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

type Round2 struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round2{}

func New(ctx *cylinder.Context) (*Round2, error) {
	// create http client
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Round2{
		context: ctx,
		logger:  ctx.Logger.With("worker", "round2"),
		client:  cli,
	}, nil
}

func (r *Round2) subscribe() error {
	var err error
	r.eventCh, err = r.client.Subscribe(
		"round2",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s EXISTS",
			types.EventTypeRound1Success,
			types.AttributeKeyGroupID,
		),
		1000,
	)
	return err
}

func (r *Round2) handleTxResult(txResult abci.TxResult) {
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

func (r *Round2) handleEvent(event *Event) {
	logger := r.logger.With("gid", event.GroupID)
	logger.Info(":delivery_truck: Processing incoming group event")

	// set group data
	group, err := r.context.Store.GetGroup(event.GroupID)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err.Error())
		return
	}

	gr, err := r.client.QueryGroup(event.GroupID)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query group information: %s", err.Error())
		return
	}

	// get all one time public keys in the group
	oneTimePubKeys := make(tss.PublicKeys, gr.Group.Size_)
	for mid, commitment := range gr.AllRound1Commitments {
		oneTimePubKeys[mid-1] = commitment.OneTimePubKey
	}

	// calculate encrypted secret shares
	encSecretShares, err := tss.ComputeEncryptedSecretShares(
		group.MemberID,
		group.OneTimePrivKey,
		oneTimePubKeys,
		group.Coefficients,
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to genrate encrypted secret shares: %s", err.Error())
		return
	}

	// generate message
	// TODO-CYLINDER: generate round2 message
	fmt.Printf("%+v", encSecretShares)
	msg := &types.MsgSubmitDKGRound1{}

	r.context.MsgCh <- msg
}

func (r *Round2) Start() {
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

func (r *Round2) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}
