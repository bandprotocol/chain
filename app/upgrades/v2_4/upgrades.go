package v2_4

import (
	"github.com/bandprotocol/chain/v2/app/keepers"
	"github.com/bandprotocol/chain/v2/app/upgrades"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	am upgrades.AppManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, _ module.VersionMap) (module.VersionMap, error) {
		// hardcode version of all modules of v2.3.x
		fromVM := map[string]uint64{
			"auth":         2,
			"authz":        1,
			"bank":         2,
			"capability":   1,
			"crisis":       1,
			"distribution": 2,
			"evidence":     1,
			"feegrant":     1,
			"genutil":      1,
			"gov":          2,
			"ibc":          2,
			"mint":         1,
			"oracle":       1,
			"params":       1,
			"slashing":     2,
			"staking":      2,
			"transfer":     1,
			"upgrade":      1,
			"vesting":      1,
		}

		// set version of ica so that it won't run initgenesis again
		fromVM["interchainaccounts"] = 1

		// prepare ICS27 controller and host params
		controllerParams := icacontrollertypes.Params{}
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			AllowMessages: []string{
				sdk.MsgTypeURL(&authz.MsgExec{}),
				sdk.MsgTypeURL(&authz.MsgGrant{}),
				sdk.MsgTypeURL(&authz.MsgRevoke{}),
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
				sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
				sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
				sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
				sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
				sdk.MsgTypeURL(&feegrant.MsgGrantAllowance{}),
				sdk.MsgTypeURL(&feegrant.MsgRevokeAllowance{}),
				sdk.MsgTypeURL(&govv1beta1.MsgVoteWeighted{}),
				sdk.MsgTypeURL(&govv1beta1.MsgSubmitProposal{}),
				sdk.MsgTypeURL(&govv1beta1.MsgDeposit{}),
				sdk.MsgTypeURL(&govv1beta1.MsgVote{}),
				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
				sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
			},
		}

		// Oracle DefaultParams only upgrade BaseRequestGas to 50000
		keepers.OracleKeeper.SetParams(ctx, oracletypes.DefaultParams())

		consensusParam := am.GetConsensusParams(ctx)
		consensusParam.Block.MaxGas = 50_000_000
		am.StoreConsensusParams(ctx, consensusParam)

		// initialize ICS27 module
		icaModule, _ := mm.Modules[icatypes.ModuleName].(ica.AppModule)
		icaModule.InitModule(ctx, controllerParams, hostParams)

		// run migration
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
