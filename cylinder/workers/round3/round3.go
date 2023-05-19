package round3

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

// Round2 is a worker responsible for round2 in the DKG process of TSS module
type Round3 struct {
	context *cylinder.Context

	logger *logger.Logger
	client *client.Client

	eventCh <-chan ctypes.ResultEvent
}

var _ cylinder.Worker = &Round3{}

// New creates a new instance of the Round2 worker.
// It initializes the necessary components and returns the created Round2 instance or an error if initialization fails.
func New(ctx *cylinder.Context) (*Round3, error) {
	cli, err := client.New(ctx.Config, ctx.Keyring)
	if err != nil {
		return nil, err
	}

	return &Round3{
		context: ctx,
		logger:  ctx.Logger.With("worker", "round2"),
		client:  cli,
	}, nil
}

// subscribe subscribes to the round2 events and initializes the event channel for receiving events.
// It returns an error if the subscription fails.
func (r *Round3) subscribe() error {
	var err error
	r.eventCh, err = r.client.Subscribe(
		"round2",
		fmt.Sprintf(
			"tm.event = 'Tx' AND %s.%s EXISTS",
			types.EventTypeRound2Success,
			types.AttributeKeyGroupID,
		),
		1000,
	)
	return err
}

// handleTxResult handles the result of a transaction.
// It extracts the relevant message logs from the transaction result and processes the events.
func (r *Round3) handleTxResult(txResult abci.TxResult) {
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

// handleEvent processes an incoming group event.
func (r *Round3) handleEvent(event *Event) {
	logger := r.logger.With("gid", event.GroupID)
	logger.Info(":delivery_truck: Processing incoming group event")

	// Set group data
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

	var secretShares tss.Scalars
	var complains []map[string]any
	for j := uint64(0); j < gr.Group.Size_; j++ {
		// Calculate you own secret value
		if j+1 == uint64(group.MemberID) {
			secretShare := tss.ComputeSecretShare(group.Coefficients, uint32(group.MemberID))
			secretShares = append(secretShares, secretShare)
			continue
		}

		// Get secret share
		secretShare, err := getSecretShare(group.MemberID, tss.MemberID(j+1), gr, group.OneTimePrivKey)
		if err != nil {
			logger.Error(":cold_sweat: Failed to get secret share with MemberID(%d): %s", j+1, err.Error())
			return
		}

		// Verify secert share
		err = tss.VerifySecretShare(group.MemberID, secretShare, gr.AllRound1Commitments[j+1].CoefficientsCommit)
		if err != nil {
			// Generate complain if we fail to verify secret share
			sig, keySym, nonceSym, err := tss.SignComplain(
				gr.AllRound1Commitments[uint64(group.MemberID)].OneTimePubKey,
				gr.AllRound1Commitments[j+1].OneTimePubKey,
				group.OneTimePrivKey,
			)
			if err != nil {
				logger.Error(":cold_sweat: Failed to generate complain: %s", err.Error())
				return
			}

			// Add complain
			complains = append(complains, map[string]any{
				"i":        group.MemberID,
				"j":        j + 1,
				"sig":      sig,
				"keySym":   keySym,
				"nonceSym": nonceSym,
			})

			continue
		}

		// Add secret share if verification is successful
		secretShares = append(secretShares, secretShare)
	}

	fmt.Printf("shares: %+v\n", secretShares)
	fmt.Printf("complains: %+v\n", complains)

	if len(complains) == 0 {
		// Send message confirm
		ownPrivKey := tss.ComputeOwnPrivateKey(secretShares)
		group.PrivKey = ownPrivKey

		fmt.Printf("ownPrivKey: %+v\n", group.PrivKey)
		r.context.Store.SetGroup(event.GroupID, group)

		// TODO-CYLINDER: USE THE REAL MESSAGE
		// r.context.MsgCh <- &types.MsgSubmitDKGRound2{
		// 	GroupID: event.GroupID,
		// 	// confirm
		// 	Member: r.context.Config.Granter,
		// }
	} else {
		// Send message complains
		// TODO-CYLINDER: USE THE REAL MESSAGE
		// r.context.MsgCh <- &types.MsgSubmitDKGRound2{
		// 	GroupID: event.GroupID,
		// 	// complains
		// 	Member: r.context.Config.Granter,
		// }
	}
}

// Start starts the Round2 worker.
// It subscribes to round2 events and starts processing incoming events.
func (r *Round3) Start() {
	r.logger.Info("start")

	err := r.subscribe()
	if err != nil {
		r.context.ErrCh <- err
		return
	}

	for ev := range r.eventCh {
		go r.handleTxResult(ev.Data.(tmtypes.EventDataTx).TxResult)
	}
}

// Stop stops the Round2 worker.
func (r *Round3) Stop() {
	r.logger.Info("stop")
	r.client.Stop()
}

func getSecretShare(
	i, j tss.MemberID,
	gr *types.QueryGroupResponse,
	privKeyI tss.PrivateKey,
) (tss.Scalar, error) {
	// Calculate keySym
	pubKeyJ := gr.AllRound1Commitments[uint64(j)].OneTimePubKey
	keySym, err := tss.ComputeKeySym(privKeyI, pubKeyJ)
	if err != nil {
		return nil, err
	}

	// Calculate secret share between yourself and J
	var encSecretShare tss.Scalar
	if i < j {
		encSecretShare = gr.Round2Shares[j-1].EncryptedSecretShares[i-1]
	} else {
		encSecretShare = gr.Round2Shares[j-1].EncryptedSecretShares[i-2]
	}

	// Decrypt
	secretShare := tss.Decrypt(encSecretShare, keySym)

	return secretShare, nil
}
