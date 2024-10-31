package updater

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/bandprotocol/chain/v3/pkg/logger"
)

const (
	CurrentFeedsQuery    = "tm.event = 'NewBlock' AND update_current_feeds.last_update_block EXISTS"
	UpdateRefSourceQuery = "update_reference_source_config.ipfs_hash EXISTS"
	// EventChannelCapacity is a buffer size of channel between node and this program
	EventChannelCapacity = 2000
)

type Updater struct {
	feedQuerier  FeedQuerier
	bothanClient BothanClient
	clients      []rpcclient.RemoteClient
	logger       *logger.Logger

	maxCurrentFeedsEventHeight    *atomic.Int64
	maxUpdateRefSourceEventHeight *atomic.Int64
}

func New(
	feedQuerier FeedQuerier,
	bothanClient BothanClient,
	clients []rpcclient.RemoteClient,
	logger *logger.Logger,
	maxCurrentFeedsEventHeight *atomic.Int64,
	maxUpdateRefSourceEventHeight *atomic.Int64,
) *Updater {
	return &Updater{
		feedQuerier:                   feedQuerier,
		bothanClient:                  bothanClient,
		clients:                       clients,
		logger:                        logger,
		maxCurrentFeedsEventHeight:    maxCurrentFeedsEventHeight,
		maxUpdateRefSourceEventHeight: maxUpdateRefSourceEventHeight,
	}
}

func (u *Updater) Start(sigChan chan<- os.Signal) {
	// initialize the updater
	if err := u.Init(); err != nil {
		u.logger.Error("[Updater] failed to initialize updater: %v", err)
		sigChan <- syscall.SIGTERM
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			go processEvent(
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
			go processEvent(
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

// Init initialize the updater
func (u *Updater) Init() error {
	if err := u.updateBothanRegistry(); err != nil {
		return err
	}

	if err := u.updateBothanActiveFeeds(); err != nil {
		return err
	}

	return nil
}

func (u *Updater) subscribeToClient(
	ctx context.Context,
	client rpcclient.RemoteClient,
	query string,
	outChan chan<- coretypes.ResultEvent,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	u.logger.Info("[Updater] Subscribing to events of client with URI: %s, with query: %s ", client.Remote(), query)
	eventChan, err := client.Subscribe(ctx, "", query, EventChannelCapacity)
	if err != nil {
		u.logger.Error(
			"[Updater] Error subscribing to events of client with URI: %s, with error: %v",
			client.Remote(),
			err,
		)
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

func (u *Updater) updateBothanActiveFeeds() error {
	queryResp, err := u.feedQuerier.QueryCurrentFeeds()
	if err != nil {
		u.logger.Error("[Updater] failed to query current feeds: %v", err)
		return err
	}

	currentFeeds := queryResp.CurrentFeeds.Feeds

	// create a list of signal IDs
	signalIDs := make([]string, 0, len(currentFeeds))
	for _, feed := range currentFeeds {
		signalIDs = append(signalIDs, feed.SignalID)
	}

	err = u.bothanClient.SetActiveSignalIDs(signalIDs)
	if err != nil {
		u.logger.Error("[Updater] failed to update active feeds: %v", err)
		return err
	}

	u.logger.Info("[Updater] successfully updated active feeds with signal IDs: %v", signalIDs)
	return nil
}

func (u *Updater) updateBothanRegistry() error {
	queryResp, err := u.feedQuerier.QueryReferenceSourceConfig()
	if err != nil {
		u.logger.Error("[Updater] failed to query reference source config: %v", err)
		return err
	}

	rfc := queryResp.ReferenceSourceConfig

	if rfc.RegistryIPFSHash == "[NOT_SET]" || rfc.RegistryVersion == "[NOT_SET]" {
		u.logger.Warn("[Updater] reference source config is not set, skipping update")
		return nil
	}

	err = u.bothanClient.UpdateRegistry(rfc.RegistryIPFSHash, rfc.RegistryVersion)
	if err != nil {
		u.logger.Error("[Updater] failed to update registry: %v", err)
		return err
	}

	u.logger.Info("[Updater] successfully updated registry with IPFS hash: %s", rfc.RegistryIPFSHash)
	return nil
}
