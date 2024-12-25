package band

import (
	signerextraction "github.com/skip-mev/block-sdk/v2/adapters/signer_extraction_adapter"
	"github.com/skip-mev/block-sdk/v2/block/base"
	defaultlane "github.com/skip-mev/block-sdk/v2/lanes/base"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"

	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
)

const (
	maxTxPerFeeds       = 500  // this is the maximum # of bids that will be held in the app-side in-memory mempool
	maxTxPerDefaultLane = 3000 // all other txs
)

var (
	defaultLaneBlockspacePercentage = math.LegacyMustNewDecFromStr("0.90")
	FeedsBlockspacePercentage       = math.LegacyMustNewDecFromStr("0.10")
)

// CreateLanes walks through the process of creating the lanes for the block sdk. In this function
// we create three separate lanes - MEV, Free, and Default - and then return them.
func CreateLanes(app *BandApp, txConfig client.TxConfig) (*base.BaseLane, *base.BaseLane) {
	// Create the signer extractor. This is used to extract the expected signers from
	// a transaction. Each lane can have a different signer extractor if needed.
	signerAdapter := signerextraction.NewDefaultAdapter()

	// Create the configurations for each lane. These configurations determine how many
	// transactions the lane can store, the maximum block space the lane can consume, and
	// the signer extractor used to extract the expected signers from a transaction.

	// Create a mev configuration that accepts maxTxPerFeeds transactions and consumes FeedsBlockspacePercentage of the
	// block space.
	feedsLaneConfig := base.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       txConfig.TxEncoder(),
		TxDecoder:       txConfig.TxDecoder(),
		MaxBlockSpace:   FeedsBlockspacePercentage,
		SignerExtractor: signerAdapter,
		MaxTxs:          maxTxPerFeeds,
	}

	// Create a default configuration that accepts maxTxPerDefaultLane transactions and consumes defaultLaneBlockspacePercentage of the
	// block space.
	defaultConfig := base.LaneConfig{
		Logger:          app.Logger(),
		TxEncoder:       txConfig.TxEncoder(),
		TxDecoder:       txConfig.TxDecoder(),
		MaxBlockSpace:   defaultLaneBlockspacePercentage,
		SignerExtractor: signerAdapter,
		MaxTxs:          maxTxPerDefaultLane,
	}

	// Create the match handlers for each lane. These match handlers determine whether or not
	// a transaction belongs in the lane.

	// Create the final match handler for the default lane. I.e this will direct all txs that are
	// not free nor mev to this lane
	defaultMatchHandler := base.DefaultMatchHandler()

	options := []base.LaneOption{
		base.WithMatchHandler(FeedsMatchHandler()),
	}

	// Create the lanes.
	feedsLane, err := base.NewBaseLane(feedsLaneConfig, "feeds", options...)
	if err != nil {
		panic(err)
	}

	defaultLane := defaultlane.NewDefaultLane(
		defaultConfig,
		defaultMatchHandler,
	)

	return feedsLane, defaultLane
}

func FeedsMatchHandler() base.MatchHandler {
	return func(_ sdk.Context, tx sdk.Tx) bool {
		msgs := tx.GetMsgs()
		if len(msgs) == 0 {
			return false
		}

		for _, msg := range msgs {
			if _, ok := msg.(*feedstypes.MsgSubmitSignalPrices); !ok {
				return false
			}
		}
		return true
	}
}
