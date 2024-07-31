package updater

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/bandprotocol/chain/v2/grogu/querier"
	"github.com/bandprotocol/chain/v2/pkg/logger"
)

const (
	CurrentFeedsQuery    = "tm.event = 'NewBlock' AND update_current_feeds.last_update_block EXISTS"
	UpdateRefSourceQuery = "update_reference_source_config.ipfs_hash EXISTS"
	// EventChannelCapacity is a buffer size of channel between node and this program
	EventChannelCapacity = 2000
)

type Updater struct {
	feedQuerier *querier.FeedQuerier
	clients     []*http.HTTP
	logger      *logger.Logger

	maxCurrentFeedsEventHeight    *atomic.Int64
	maxUpdateRefSourceEventHeight *atomic.Int64
}

func New(
	feedQuerier *querier.FeedQuerier,
	clients []*http.HTTP,
	logger *logger.Logger,
	maxCurrentFeedsEventHeight *atomic.Int64,
	maxUpdateRefSourceEventHeight *atomic.Int64,
) *Updater {
	return &Updater{
		feedQuerier:                   feedQuerier,
		clients:                       clients,
		logger:                        logger,
		maxCurrentFeedsEventHeight:    maxCurrentFeedsEventHeight,
		maxUpdateRefSourceEventHeight: maxUpdateRefSourceEventHeight,
	}
}

func (u *Updater) Start(sigChan chan<- os.Signal) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventCurrentFeedsChan := make(chan coretypes.ResultEvent)
	eventUpdateRefSourceChan := make(chan coretypes.ResultEvent)
	var wgCurrentFeeds sync.WaitGroup
	var wgUpdateRefSource sync.WaitGroup

	for _, client := range u.clients {
		wgCurrentFeeds.Add(1)
		go u.subscribeToClient(ctx, client, CurrentFeedsQuery, eventCurrentFeedsChan, &wgCurrentFeeds)
		wgUpdateRefSource.Add(1)
		go u.subscribeToClient(ctx, client, UpdateRefSourceQuery, eventUpdateRefSourceChan, &wgUpdateRefSource)
	}

	go u.waitForCompletion(&wgCurrentFeeds, sigChan)
	go u.waitForCompletion(&wgUpdateRefSource, sigChan)

	for {
		select {
		case ev := <-eventCurrentFeedsChan:
			processEvent(
				ev,
				u.logger,
				CurrentFeedsQuery,
				u.maxCurrentFeedsEventHeight,
				func(ev coretypes.ResultEvent) int64 {
					return ev.Data.(tmtypes.EventDataNewBlock).Block.Height
				},
				u.updateBothanActiveFeeds,
			)
		case ev := <-eventUpdateRefSourceChan:
			processEvent(
				ev,
				u.logger,
				UpdateRefSourceQuery,
				u.maxUpdateRefSourceEventHeight,
				func(ev coretypes.ResultEvent) int64 {
					switch eventData := ev.Data.(type) {
					case tmtypes.EventDataTx:
						return eventData.TxResult.Height
					case tmtypes.EventDataNewBlock:
						return eventData.Block.Height
					default:
						return 0
					}
				},
				u.updateBothanRegistry,
			)
		}
	}
}

func (u *Updater) subscribeToClient(
	ctx context.Context,
	client *http.HTTP,
	query string,
	outChan chan<- coretypes.ResultEvent,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	u.logger.Info("[Updater] Subscribing to events with query: %s...", query)
	eventChan, err := client.Subscribe(ctx, "", query, EventChannelCapacity)
	if err != nil {
		u.logger.Error("[Updater] Error subscribing to events: %s", err)
		return
	}

	for event := range eventChan {
		outChan <- event
	}
}

func (u *Updater) waitForCompletion(
	wg *sync.WaitGroup,
	sigChan chan<- os.Signal,
) {
	wg.Wait()
	sigChan <- syscall.SIGTERM
}

func (u *Updater) updateBothanActiveFeeds() {
	// TODO: Implement the updateBothanActiveFeeds function
}

func (u *Updater) updateBothanRegistry() {
	// TODO: Implement the updateBothanRegistry function
}
