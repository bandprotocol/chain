package signing

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/pkg/event"
	"github.com/bandprotocol/chain/v2/pkg/logger"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// Signing is a worker responsible for the signing process of the TSS module.
type Signing struct {
	context *cylinder.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Signing{}

// New creates a new instance of the Signing worker.
// It initializes the necessary components and returns the created Signing instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Signing, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Signing{
		context: ctx,
		logger:  ctx.Logger.With("worker", "Signing"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the request_sign events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (s *Signing) subscribe() (err error) {
	subscriptionQuery := fmt.Sprintf(
		"%s.%s = '%s'",
		types.EventTypeRequestSignature,
		types.AttributeKeyMember,
		s.context.Config.Granter,
	)
	s.eventCh, err = s.client.Subscribe("Signing", subscriptionQuery, 1000)
	return
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (s *Signing) handleTxResult(txResult abci.TxResult) {
	msgLogs, err := event.GetMessageLogs(txResult)
	if err != nil {
		s.logger.Error("Failed to get message logs: %s", err)
		return
	}

	for _, log := range msgLogs {
		event, err := ParseEvent(log.Events)
		if err != nil {
			s.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
			return
		}

		go s.handleSigning(event.SigningID)
	}
}

// handleABCIEvents handles the end block events.
func (s *Signing) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeRequestSignature {
			event, err := ParseEvent(sdk.StringEvents{ev})
			if err != nil {
				s.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			go s.handleSigning(event.SigningID)
		}
	}
}

// handleSigning processes an incoming signing request.
func (s *Signing) handleSigning(sid tss.SigningID) {
	logger := s.logger.With("sid", sid)

	// Log
	logger.Info(":delivery_truck: Processing incoming signing request")

	// Query signing detail
	signingRes, err := s.client.QuerySigning(sid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query signing information: %s", err)
		return
	}

	signing := signingRes.Signing
	assignedMember, err := signingRes.GetAssignedMember(s.context.Config.Granter)
	if err != nil {
		logger.Error(":cold_sweat: Failed to get assigned member: %s", err)
		return
	}

	// Set group data
	group, err := s.context.Store.GetGroup(signing.GroupPubKey)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err)
		return
	}

	// Get private keys of DE
	privDE, err := s.context.Store.GetDE(types.DE{
		PubD: assignedMember.PubD,
		PubE: assignedMember.PubE,
	})
	if err != nil {
		logger.Error(":cold_sweat: Failed to get private DE from store: %s", err)
		return
	}

	// Compute own private nonce
	privNonce, err := tss.ComputeOwnPrivNonce(privDE.PrivD, privDE.PrivE, assignedMember.BindingFactor)
	if err != nil {
		logger.Error(":cold_sweat: Failed to compute own private nonce: %s", err)
		return
	}

	// Compute lagrange
	lagrange := tss.ComputeLagrangeCoefficient(group.MemberID, signingRes.GetMemberIDs())

	// Sign the signing
	sig, err := tss.SignSigning(
		signing.GroupPubNonce,
		signing.GroupPubKey,
		signing.Message,
		lagrange,
		privNonce,
		group.PrivKey,
	)
	if err != nil {
		logger.Error(":cold_sweat: Failed to sign signing: %s", err)
		return
	}

	// Send MsgSigning
	s.context.MsgCh <- &types.MsgSubmitSignature{
		SigningID: sid,
		MemberID:  group.MemberID,
		Signature: sig,
		Address:   s.context.Config.Granter,
	}
}

// handlePendingSignings processes the pending signing requests.
func (s *Signing) handlePendingSignings() {
	res, err := s.client.QueryPendingSignings(s.context.Config.Granter)
	if err != nil {
		s.logger.Error(":cold_sweat: Failed to get pending signings: %s", err)
		return
	}

	for _, sid := range res.PendingSignings {
		go s.handleSigning(tss.SigningID(sid))
	}
}

// Start starts the Signing worker.
// It subscribes to events and starts processing incoming events.
func (s *Signing) Start() {
	s.logger.Info("start")

	err := s.subscribe()
	if err != nil {
		s.context.ErrCh <- err
		return
	}

	s.handlePendingSignings()

	for ev := range s.eventCh {
		switch data := ev.Data.(type) {
		case tmtypes.EventDataTx:
			go s.handleTxResult(data.TxResult)
		case tmtypes.EventDataNewBlock:
			go s.handleABCIEvents(data.ResultEndBlock.Events)
		}
	}
}

// Stop stops the Signing worker.
func (s *Signing) Stop() {
	s.logger.Info("stop")
	s.client.Stop()
}
