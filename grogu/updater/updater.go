package updater

import (
	"os"
	"time"

	rpcclient "github.com/cometbft/cometbft/rpc/client"

	"github.com/bandprotocol/chain/v3/pkg/logger"
)

type Updater struct {
	feedQuerier  FeedQuerier
	bothanClient BothanClient
	clients      []rpcclient.RemoteClient
	logger       *logger.Logger

	queryInterval time.Duration
}

func New(
	feedQuerier FeedQuerier,
	bothanClient BothanClient,
	clients []rpcclient.RemoteClient,
	logger *logger.Logger,
	queryInterval time.Duration,
) *Updater {
	return &Updater{
		feedQuerier:   feedQuerier,
		bothanClient:  bothanClient,
		clients:       clients,
		logger:        logger,
		queryInterval: queryInterval,
	}
}

func (u *Updater) Start(sigChan chan<- os.Signal) {
	ticker := time.NewTicker(u.queryInterval)
	defer ticker.Stop()

	for range ticker.C {
		u.checkAndUpdateBothan()
	}
}

func (u *Updater) checkAndUpdateBothan() {
	chainConfig, err := u.feedQuerier.QueryReferenceSourceConfig()
	if err != nil {
		u.logger.Error("[Updater] failed to query chain config: %v", err)
		return
	}

	rfc := chainConfig.ReferenceSourceConfig

	if rfc.RegistryIPFSHash == "[NOT_SET]" || rfc.RegistryVersion == "[NOT_SET]" {
		u.logger.Debug("[Updater] reference source config is not set, skipping update")
		return
	}

	bothanInfo, err := u.bothanClient.GetInfo()
	if err != nil {
		u.logger.Error("[Updater] failed to query Bothan info: %v", err)
		return
	}

	if rfc.RegistryIPFSHash == bothanInfo.RegistryIpfsHash {
		u.logger.Debug("[Updater] chain and Bothan config match, skipping update")
		return
	}

	u.logger.Info("[Updater] chain and Bothan config mismatch detected, updating registry")
	err = u.bothanClient.UpdateRegistry(rfc.RegistryIPFSHash, rfc.RegistryVersion)
	if err != nil {
		u.logger.Error("[Updater] failed to update registry: %v", err)
		return
	}

	u.logger.Info("[Updater] successfully updated registry with IPFS hash: %s", rfc.RegistryIPFSHash)
}
