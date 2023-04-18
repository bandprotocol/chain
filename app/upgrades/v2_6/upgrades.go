package v2_6

import (
	"github.com/bandprotocol/chain/v2/app/keepers"
	"github.com/bandprotocol/chain/v2/app/upgrades"
	"github.com/bandprotocol/chain/v2/x/globalfee"
	globalfeetypes "github.com/bandprotocol/chain/v2/x/globalfee/types"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
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
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		hostParams := icahosttypes.Params{
			HostEnabled: true,
			// Specifying the whole list instead of adding and removing. Less fragile.
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
				// Change: add messages from Group module
				sdk.MsgTypeURL(&group.MsgCreateGroupPolicy{}),
				sdk.MsgTypeURL(&group.MsgCreateGroupWithPolicy{}),
				sdk.MsgTypeURL(&group.MsgCreateGroup{}),
				sdk.MsgTypeURL(&group.MsgExec{}),
				sdk.MsgTypeURL(&group.MsgLeaveGroup{}),
				sdk.MsgTypeURL(&group.MsgSubmitProposal{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupAdmin{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupMembers{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupMetadata{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyAdmin{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyDecisionPolicy{}),
				sdk.MsgTypeURL(&group.MsgUpdateGroupPolicyMetadata{}),
				sdk.MsgTypeURL(&group.MsgVote{}),
				sdk.MsgTypeURL(&group.MsgWithdrawProposal{}),
				// Change: add messages from Oracle module
				sdk.MsgTypeURL(&oracletypes.MsgActivate{}),
				sdk.MsgTypeURL(&oracletypes.MsgCreateDataSource{}),
				sdk.MsgTypeURL(&oracletypes.MsgCreateOracleScript{}),
				sdk.MsgTypeURL(&oracletypes.MsgEditDataSource{}),
				sdk.MsgTypeURL(&oracletypes.MsgEditOracleScript{}),
				sdk.MsgTypeURL(&oracletypes.MsgReportData{}),
				sdk.MsgTypeURL(&oracletypes.MsgRequestData{}),

				sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
				sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
				sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
				sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
				sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
			},
		}
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		minGasPriceGenesisState := &globalfeetypes.GenesisState{
			Params: globalfeetypes.Params{
				MinimumGasPrices: sdk.DecCoins{sdk.NewDecCoinFromDec("uband", sdk.NewDecWithPrec(25, 4))},
			},
		}
		am.GetSubspace(globalfee.ModuleName).SetParamSet(ctx, &minGasPriceGenesisState.Params)

		// set version of globalfee so that it won't run initgenesis again
		fromVM["globalfee"] = 1

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
