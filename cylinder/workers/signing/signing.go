package signing

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/cylinder"
	"github.com/bandprotocol/chain/v3/cylinder/client"
	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/pkg/logger"
	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// Signing is a worker responsible for the signing process of the TSS module.
type Signing struct {
	context *context.Context
	logger  *logger.Logger
	client  *client.Client
	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Signing{}

// New creates a new instance of the Signing worker.
// It initializes the necessary components and returns the created Signing instance or an error if initialization fails.
func New(ctx *context.Context) (*Signing, error) {
	cli, err := client.New(ctx)
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
		types.AttributeKeyAddress,
		s.context.Config.Granter,
	)
	s.eventCh, err = s.client.Subscribe("Signing", subscriptionQuery, 1000)
	return
}

// handleABCIEvents signs the specific signingID if the given events contain a request_signature event.
func (s *Signing) handleABCIEvents(abciEvents []abci.Event) {
	events := sdk.StringifyEvents(abciEvents)
	for _, ev := range events {
		if ev.Type == types.EventTypeRequestSignature {
			events, err := ParseEvents(sdk.StringEvents{ev})
			if err != nil {
				s.logger.Error(":cold_sweat: Failed to parse event with error: %s", err)
				return
			}

			for _, event := range events {
				go s.handleSigning(event.SigningID)
			}
		}
	}
}

// handleSigning processes an incoming signing request.
func (s *Signing) handleSigning(sid tss.SigningID) {
	logger := s.logger.With("sid", sid)

	// Query signing detail
	signingRes, err := s.client.QuerySigning(sid)
	if err != nil {
		logger.Error(":cold_sweat: Failed to query signing information: %s", err)
		return
	}

	signing := signingRes.SigningResult.Signing
	assignedMember, err := signingRes.GetAssignedMember(s.context.Config.Granter)
	if err != nil {
		return
	}

	// Log
	logger.Info(":delivery_truck: Processing incoming signing request")

	// Set group data
	group, err := s.context.Store.GetGroup(signing.GroupPubKey)
	if err != nil {
		logger.Error(":cold_sweat: Failed to find group in store: %s", err)
		return
	}

	// Get private keys of DE
	provide, err := s.context.Store.GetDE(types.DE{
		PubD: assignedMember.PubD,
		PubE: assignedMember.PubE,
	})
	if err != nil {
		logger.Error(":cold_sweat: Failed to get private DE from store: %s", err)
		return
	}

	// Compute own private nonce
	privNonce, err := tss.ComputeOwnPrivNonce(provide.PrivD, provide.PrivE, assignedMember.BindingFactor)
	if err != nil {
		logger.Error(":cold_sweat: Failed to compute own private nonce: %s", err)
		return
	}

	// Compute lagrange
	lagrange, err := tss.ComputeLagrangeCoefficient(group.MemberID, signingRes.GetMemberIDs())
	if err != nil {
		logger.Error(":cold_sweat: Failed to compute lagrange coefficient: %s", err)
		return
	}

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
	s.context.MsgCh <- types.NewMsgSubmitSignature(sid, group.MemberID, sig, s.context.Config.Granter)
}

// handlePendingSignings processes the pending signing requests.
func (s *Signing) handlePendingSignings() {
	res, err := s.client.QueryPendingSignings(s.context.Config.Granter)
	if err != nil {
		s.logger.Error(":cold_sweat: Failed to get pending signings: %s", err)
		return
	}

	for _, sid := range res.PendingSignings {
		go s.handleSigning(sid)
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
			go s.handleABCIEvents(data.TxResult.Result.Events)
		case tmtypes.EventDataNewBlock:
			go s.handleABCIEvents(data.ResultFinalizeBlock.Events)
		}
	}
}

// Stop stops the Signing worker.
func (s *Signing) Stop() error {
	s.logger.Info("stop")
	return s.client.Stop()
}
