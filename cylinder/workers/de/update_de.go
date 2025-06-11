package de

import (
	"fmt"
	"time"

	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/metrics"
	"github.com/bandprotocol/chain/v3/cylinder/msg"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

const MAX_DE_BATCH_SIZE = 50

// UpdateDE is a worker responsible for updating DEs in the store and chains
type UpdateDE struct {
	context          *context.Context
	logger           *logger.Logger
	client           *client.Client
	eventCh          <-chan ctypes.ResultEvent
	deCounter        *DECounter
	maxDESizeOnChain uint64
	receiver         msg.ResponseReceiver
	reqID            uint64
	cacheDEs         map[uint64]int64
}

var _ cylinder.Worker = &UpdateDE{}

// NewUpdateDE creates a new UpdateDE worker.
func NewUpdateDE(ctx *context.Context) (*UpdateDE, error) {
	cli, err := client.New(ctx)
	if err != nil {
		return nil, err
	}

	params, err := cli.QueryTssParams()
	if err != nil {
		return nil, err
	}

	receiver := msg.ResponseReceiver{
		ReqType:    msg.RequestTypeUpdateDE,
		ResponseCh: make(chan msg.Response),
	}

	return &UpdateDE{
		context:          ctx,
		logger:           ctx.Logger.With("worker", "UpdateDE"),
		client:           cli,
		maxDESizeOnChain: params.MaxDESize,
		receiver:         receiver,
		deCounter:        NewDECounter(),
		cacheDEs:         make(map[uint64]int64),
	}, nil
}

// Start starts the UpdateDE worker.
func (u *UpdateDE) Start() {
	u.logger.Info("start")

	if err := u.subscribe(); err != nil {
		u.context.ErrCh <- err
		return
	}

	// Update one time when starting worker first time.
	if err := u.intervalUpdateDE(); err != nil {
		u.context.ErrCh <- err
		return
	}

	go u.listenMsgResponses()

	// Update DE if there is assigned DE event or DE is used.
	ticker := time.NewTicker(u.context.Config.CheckDEInterval)
	for {
		select {
		case <-ticker.C:
			if err := u.intervalUpdateDE(); err != nil {
				u.logger.Error(":cold_sweat: Failed to do an interval update DE: %s", err)
			}
		case resultEvent := <-u.eventCh:
			if err := u.updateDEFromEvent(resultEvent); err != nil {
				u.logger.Error(":cold_sweat: Failed to update DE from assigned DE event: %s", err)
			}
		}
	}
}

// Stop stops the UpdateDE worker.
func (u *UpdateDE) Stop() error {
	u.logger.Info("stop")
	return u.client.Stop()
}

// subscribe subscribes to the events that trigger the DE update.
func (u *UpdateDE) subscribe() (err error) {
	assignedDEQuery := fmt.Sprintf(
		"%s.%s = '%s'",
		types.EventTypeRequestSignature,
		types.AttributeKeyAddress,
		u.context.Config.Granter,
	)

	u.eventCh, err = u.client.Subscribe("AssignedDE", assignedDEQuery, 1000)
	return err
}

// updateDE generates new DEs and submit them to the chain.
func (u *UpdateDE) updateDE(numNewDE uint64) error {
	u.logger.Info(":delivery_truck: Updating DE")

	// Generate new DE pairs
	privDEs, err := GenerateDEs(
		numNewDE,
		u.context.Config.RandomSecret,
		u.context.Store,
	)
	if err != nil {
		return fmt.Errorf("failed to generate new DE pairs: %s", err)
	}

	// Store all DEs in the store
	var pubDEs []types.DE
	for _, privDE := range privDEs {
		pubDEs = append(pubDEs, privDE.PubDE)

		if err := u.context.Store.SetDE(privDE); err != nil {
			return fmt.Errorf("failed to set new DE in the store: %s", err)
		}

		metrics.IncOffChainDELeftGauge()
	}

	u.logger.Info(":white_check_mark: Successfully generated %d new DE pairs", numNewDE)

	// Send MsgDEs to the chain (chunked by MAX_DE_BATCH_SIZE)
	for i := 0; i < len(pubDEs); i += MAX_DE_BATCH_SIZE {
		u.reqID += 1
		end_idx := min(i+MAX_DE_BATCH_SIZE, len(pubDEs))
		u.cacheDEs[u.reqID] = int64(end_idx - i)

		u.context.PriorityMsgRequestCh <- msg.NewRequest(
			msg.RequestTypeUpdateDE,
			u.reqID,
			types.NewMsgSubmitDEs(pubDEs[i:end_idx], u.context.Config.Granter),
			0,
		)
	}

	return nil
}

// isTssMember checks if the granter is a tss member.
func (u *UpdateDE) isTssMember() (bool, error) {
	resp, err := u.client.QueryMember(u.context.Config.Granter)
	if err != nil {
		return false, fmt.Errorf("failed to query member information: %w", err)
	}

	if resp.CurrentGroupMember.Address == u.context.Config.Granter ||
		resp.IncomingGroupMember.Address == u.context.Config.Granter {
		return true, nil
	}

	return false, nil
}

// isGasPriceSet checks if the gas price is set.
func (u *UpdateDE) isGasPriceSet() (bool, error) {
	gasPrices, err := sdk.ParseDecCoins(u.context.Config.GasPrices)
	if err != nil {
		return false, fmt.Errorf("failed to parse gas prices from config: %w", err)
	}

	// If the gas price is non-zero, it indicates that the user is willing to pay
	// a transaction fee for submitting DEs to the chain.
	if gasPrices != nil && !gasPrices.IsZero() {
		return true, nil
	}

	return false, nil
}

// shouldContinueUpdateDE checks if the program should generate and submit new DEs.
// It returns true if the user is a tss member or voluntarily pay for the gas
// (set gas price in the config).
func (u *UpdateDE) shouldContinueUpdateDE() (bool, error) {
	isTssMember, err := u.isTssMember()
	if err != nil {
		return false, fmt.Errorf("isTssMember error: %s", err)
	}

	isGasPriceSet, err := u.isGasPriceSet()
	if err != nil {
		return false, fmt.Errorf("isGasPriceSet error: %s", err)
	}

	if !isTssMember && !isGasPriceSet {
		u.logger.Debug(":cold_sweat: Skip updating DE; not a tss member and gas price isn't set")
		return false, nil
	}

	return true, nil
}

// intervalUpdateDE updates DE on the chain so that the remaining DE is
// always above the minimum threshold.
func (u *UpdateDE) intervalUpdateDE() error {
	// also update the maxDESizeOnChain
	params, err := u.client.QueryTssParams()
	if err != nil {
		return err
	}
	u.maxDESizeOnChain = params.MaxDESize

	deCount, blockHeight, err := u.getDECount()
	if err != nil {
		return err
	}

	metrics.SetOnChainDELeftGauge(float64(deCount))

	numDEToBeCreated := u.deCounter.AfterSyncWithChain(deCount, u.maxDESizeOnChain, blockHeight)
	u.logger.Debug(":eyes: deCounter after AfterSyncWithChain [intervalUpdateDE]: %s", u.deCounter.String())
	if numDEToBeCreated == 0 {
		u.logger.Debug(":eyes: the number of DEs is sufficient, skip interval update DE")
		return nil
	}

	if ok, err := u.shouldContinueUpdateDE(); err != nil || !ok {
		u.deCounter.AfterDEsRejected(numDEToBeCreated)
		return err
	}

	if numDEToBeCreated > 0 {
		u.logger.Info(
			":delivery_truck: the number of DEs is less than the expected size, do an interval update len = %d",
			numDEToBeCreated,
		)
		if err := u.updateDE(uint64(numDEToBeCreated)); err != nil {
			u.deCounter.AfterDEsRejected(numDEToBeCreated)
			return err
		}
	}

	return nil
}

// updateDEFromEvent updates DEs from the subscribed event.
func (u *UpdateDE) updateDEFromEvent(resultEvent ctypes.ResultEvent) error {
	memberAddress := u.context.Config.Granter

	var blockHeight, deUsed int64
	switch data := resultEvent.Data.(type) {
	case tmtypes.EventDataTx:
		blockHeight = data.Height
		deUsed = CountAssignedSignings(sdk.StringifyEvents(data.Result.Events), memberAddress)
	case tmtypes.EventDataNewBlock:
		blockHeight = data.Block.Height
		deUsed = CountAssignedSignings(sdk.StringifyEvents(data.ResultFinalizeBlock.Events), memberAddress)
	default:
		return nil
	}

	threshold := min(u.maxDESizeOnChain/6, MAX_DE_BATCH_SIZE)
	numDEToBECreated := u.deCounter.EvaluateDECreationFromUsage(deUsed, threshold, blockHeight)
	u.logger.Debug(":eyes: deCounter after EvaluateDECreationFromUsage [updateDEFromEvent]: %s", u.deCounter.String())

	if numDEToBECreated == 0 {
		u.logger.Debug(":eyes: DEs are sufficient, skip update DE from event")
		return nil
	}

	if ok, err := u.shouldContinueUpdateDE(); err != nil || !ok {
		u.deCounter.AfterDEsRejected(numDEToBECreated)
		return err
	}

	u.logger.Info(":delivery_truck: DEs are used over the threshold, adding new DEs len = %d", numDEToBECreated)

	if err := u.updateDE(uint64(numDEToBECreated)); err != nil {
		u.deCounter.AfterDEsRejected(numDEToBECreated)
		return err
	}

	return nil
}

// getDECount queries the number of DEs on the chain.
func (u *UpdateDE) getDECount() (uint64, int64, error) {
	// Query DE information
	deRes, err := u.client.QueryDE(u.context.Config.Granter, 0, 1)
	if err != nil {
		u.logger.Error(":cold_sweat: Failed to query DE information: %s", err)
		return 0, 0, err
	}

	return deRes.GetRemaining(), deRes.GetBlockHeight(), nil
}

// listenMsgResponses listens to the MsgResponseReceiver channel and handle properly.
func (u *UpdateDE) listenMsgResponses() {
	for res := range u.receiver.ResponseCh {
		lenDEs := u.cacheDEs[res.Request.ID]

		if res.Success {
			u.logger.Info(":smiling_face_with_sunglasses: Successfully submitted DEs ReqID: %d", res.Request.ID)

			u.deCounter.AfterDEsCommitted(lenDEs)
			u.logger.Debug(
				":eyes: deCounter after AfterDEsCommitted [listenMsgResponses] ReqID: %d: %s",
				res.Request.ID,
				u.deCounter.String(),
			)
		} else {
			u.logger.Error(
				":cold_sweat: Failed to submit DEs; need to revert pending DEs ReqID: %d, (len(DE): %d); error: %s",
				res.Request.ID,
				lenDEs,
				res.Err,
			)

			u.deCounter.AfterDEsRejected(lenDEs)
			u.logger.Debug(":eyes: deCounter after AfterDEsRejected [listenMsgResponses]: %s", u.deCounter.String())
		}

		delete(u.cacheDEs, res.Request.ID)
	}
}

// GetResponseReceivers returns the message response receivers of the worker.
func (u *UpdateDE) GetResponseReceivers() []*msg.ResponseReceiver {
	return []*msg.ResponseReceiver{&u.receiver}
}
