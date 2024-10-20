package band

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cast"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	"github.com/bandprotocol/chain/v3/app/keepers"
	"github.com/bandprotocol/chain/v3/hooks/common"
	"github.com/bandprotocol/chain/v3/hooks/emitter"
	"github.com/bandprotocol/chain/v3/hooks/price"
	"github.com/bandprotocol/chain/v3/hooks/request"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

const (
	FlagWithEmitter            = "with-emitter"
	FlagWithPricer             = "with-pricer"
	FlagWithRequestSearch      = "with-request-search"
	FlagRequestSearchCacheSize = "request-search-cache-size"
	FlagWithOwasmCacheSize     = "oracle-script-cache-size"
)

func NewAppHooks(appCodec codec.Codec,
	txConfig client.TxConfig,
	keepers *keepers.AppKeepers,
	homePath string,
	appOpts servertypes.AppOptions,
) common.Hooks {
	hooks := make(common.Hooks, 0)

	if emitterURI := cast.ToString(appOpts.Get(FlagWithEmitter)); emitterURI != "" {
		hooks = append(hooks, emitter.NewHook(
			appCodec,
			txConfig,
			keepers.AccountKeeper,
			keepers.BankKeeper,
			keepers.StakingKeeper,
			keepers.MintKeeper,
			keepers.DistrKeeper,
			keepers.GovKeeper,
			keepers.GroupKeeper,
			keepers.OracleKeeper,
			keepers.FeedsKeeper,
			keepers.ICAHostKeeper,
			keepers.IBCKeeper.ClientKeeper,
			keepers.IBCKeeper.ConnectionKeeper,
			keepers.IBCKeeper.ChannelKeeper,
			keepers.GetKey(group.StoreKey),
			emitterURI,
			false,
		))
	}

	if requestSearchURI := cast.ToString(appOpts.Get(FlagWithRequestSearch)); requestSearchURI != "" {
		hooks = append(hooks,
			request.NewHook(appCodec, keepers.OracleKeeper, requestSearchURI, 10),
		)
	}

	if pricerDetail := cast.ToString(appOpts.Get(FlagWithPricer)); pricerDetail != "" {
		pricerStrArgs := strings.Split(pricerDetail, "/")
		var defaultAskCount, defaultMinCount uint64
		if len(pricerStrArgs) == 3 {
			var err error
			defaultAskCount, err = strconv.ParseUint(pricerStrArgs[1], 10, 64)
			if err != nil {
				panic(err)
			}
			defaultMinCount, err = strconv.ParseUint(pricerStrArgs[2], 10, 64)
			if err != nil {
				panic(err)
			}
		} else if len(pricerStrArgs) == 2 || len(pricerStrArgs) > 3 {
			panic(fmt.Errorf("accepts 1 or 3 arg(s), received %d", len(pricerStrArgs)))
		}
		rawOracleIDs := strings.Split(pricerStrArgs[0], ",")
		var oracleIDs []oracletypes.OracleScriptID
		for _, rawOracleID := range rawOracleIDs {
			oracleID, err := strconv.ParseInt(rawOracleID, 10, 64)
			if err != nil {
				panic(err)
			}
			oracleIDs = append(oracleIDs, oracletypes.OracleScriptID(oracleID))
		}
		hooks = append(hooks,
			price.NewHook(appCodec, keepers.OracleKeeper, oracleIDs,
				filepath.Join(homePath, "prices"),
				defaultAskCount, defaultMinCount))
	}

	return hooks
}

// ABCI app to call hook
// FinalizeBlock overrides the default BaseApp's ABCI FinalizeBlock to process transaction in blocks
func (app *BandApp) FinalizeBlock(req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	// Finalize on base app first
	res, err := app.BaseApp.FinalizeBlock(req)

	beginBlockEvents, endBlockEvents := splitEvents(res.Events)
	ctx := app.BaseApp.GetContextForFinalizeBlock(nil)
	app.hooks.AfterBeginBlock(ctx, req, beginBlockEvents)

	for i := range len(req.Txs) {
		tx, _ := app.txConfig.TxDecoder()(req.Txs[i])
		resTx := res.TxResults[i]
		app.hooks.AfterDeliverTx(ctx, tx, resTx)
	}

	app.hooks.AfterEndBlock(ctx, endBlockEvents)

	return res, err
}

// Commit overrides the default BaseApp's ABCI commit to commit some state on hooks
func (app *BandApp) Commit() (res *abci.ResponseCommit, err error) {
	app.hooks.BeforeCommit()

	return app.BaseApp.Commit()
}

func splitEvents(events []abci.Event) (begins []abci.Event, ends []abci.Event) {
	for _, event := range events {
		n := len(event.Attributes)
		attrType := event.Attributes[n-1]
		if attrType.Key != "mode" {
			panic("The last attribute of begin/end block event should be mode")
		}
		if attrType.Value == "BeginBlock" {
			begins = append(begins, event)
		} else if attrType.Value == "EndBlock" {
			ends = append(ends, event)
		} else {
			panic(fmt.Sprintf("Mode of event should be BeginBlock/EndBlock got %s", attrType.Value))
		}
	}
	return begins, ends
}
