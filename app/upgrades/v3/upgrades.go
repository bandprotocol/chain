package v3

import (
	"context"
	"strings"
	"time"

	cmttypes "github.com/cometbft/cometbft/types"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/app/keepers"
	globalfeetypes "github.com/bandprotocol/chain/v3/x/globalfee/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// Set param key table for params module migration
		ctx := sdk.UnwrapSDKContext(c)
		for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
			var keyTable paramstypes.KeyTable
			switch subspace.Name() {
			// cosmos-sdk types
			case authtypes.ModuleName:
				keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
			case banktypes.ModuleName:
				keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
			case stakingtypes.ModuleName:
				keyTable = stakingtypes.ParamKeyTable() //nolint:staticcheck
			case minttypes.ModuleName:
				keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
			case distrtypes.ModuleName:
				keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
			case slashingtypes.ModuleName:
				keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
			case govtypes.ModuleName:
				keyTable = govv1.ParamKeyTable() //nolint:staticcheck
			case crisistypes.ModuleName:
				keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
			// ibc types
			case ibctransfertypes.ModuleName:
				keyTable = ibctransfertypes.ParamKeyTable()
			case ibcclienttypes.SubModuleName:
				keyTable = ibcclienttypes.ParamKeyTable()
			case ibcconnectiontypes.SubModuleName:
				keyTable = ibcconnectiontypes.ParamKeyTable()
			case icahosttypes.SubModuleName:
				keyTable = icahosttypes.ParamKeyTable()
			// band types
			case oracletypes.ModuleName:
				keyTable = oracletypes.ParamKeyTable()
			default:
				continue
			}

			if !subspace.HasKeyTable() {
				subspace.WithKeyTable(keyTable)
			}
		}

		// Move consensus parameters from the x/params module to the x/consensus module.
		baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		err := baseapp.MigrateParams(ctx, baseAppLegacySS, keepers.ConsensusParamsKeeper.ParamsStore)
		if err != nil {
			return nil, err
		}

		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		// Set new consensus params with same values as before
		consensusParams := cmttypes.DefaultConsensusParams().ToProto()
		consensusParams.Block.MaxBytes = BlockMaxBytes                                     // unchanged
		consensusParams.Block.MaxGas = BlockMaxGas                                         // unchanged
		consensusParams.Validator.PubKeyTypes = []string{cmttypes.ABCIPubKeyTypeSecp256k1} // unchanged
		err = keepers.ConsensusParamsKeeper.ParamsStore.Set(ctx, consensusParams)
		if err != nil {
			return nil, err
		}

		mintParams, err := keepers.MintKeeper.Params.Get(ctx)
		if err != nil {
			return nil, err
		}
		mintParams.BlocksPerYear = 31557600
		err = keepers.MintKeeper.Params.Set(ctx, mintParams)
		if err != nil {
			return nil, err
		}

		slashingParams, err := keepers.SlashingKeeper.GetParams(ctx)
		if err != nil {
			return nil, err
		}
		slashingParams.SignedBlocksWindow = 86400
		err = keepers.SlashingKeeper.SetParams(ctx, slashingParams)
		if err != nil {
			return nil, err
		}

		hostParams := icahosttypes.Params{
			HostEnabled: true,
			// specifying the whole list instead of adding and removing. Less fragile.
			AllowMessages: ICAAllowMessages,
		}
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)

		oracleParams := keepers.OracleKeeper.GetParams(ctx)
		oracleParams.MaxCalldataSize = 512
		oracleParams.MaxReportDataSize = 512
		err = keepers.OracleKeeper.SetParams(ctx, oracleParams)
		if err != nil {
			return nil, err
		}

		err = keepers.GlobalFeeKeeper.SetParams(ctx, globalfeetypes.Params{
			MinimumGasPrices: sdk.DecCoins{sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(25, 4))},
		})
		if err != nil {
			return nil, err
		}

		govParams, err := keepers.GovKeeper.Params.Get(ctx)
		if err != nil {
			return nil, err
		}
		govParams.ExpeditedMinDeposit = sdk.NewCoins(
			sdk.NewCoin(
				"uband",
				sdk.TokensFromConsensusPower(5000, sdk.DefaultPowerReduction)), // 5000 band
		)
		expeditedVotingPeriod := 1 * 24 * time.Hour
		govParams.ExpeditedVotingPeriod = &expeditedVotingPeriod // 1 day

		maxDepositPeriod := 5 * 24 * time.Hour
		govParams.MaxDepositPeriod = &maxDepositPeriod // 5 days

		votingPeriod := 5 * 24 * time.Hour
		govParams.VotingPeriod = &votingPeriod // 5 days

		err = keepers.GovKeeper.Params.Set(ctx, govParams)
		if err != nil {
			return nil, err
		}

		feedParams := keepers.FeedsKeeper.GetParams(ctx)
		feedParams.Admin = "band16ewcf042mtkvhsp84p59n3pt2vps0jqxl0l6gl"
		err = keepers.FeedsKeeper.SetParams(ctx, feedParams)
		if err != nil {
			return nil, err
		}

		type GrantInfo struct {
			Granter sdk.AccAddress
			Grantee sdk.AccAddress
			Grant   authz.Grant
		}

		var grants []GrantInfo
		keepers.AuthzKeeper.IterateGrants(ctx, func(granterAddr, granteeAddr sdk.AccAddress, grant authz.Grant) bool {
			grants = append(grants, GrantInfo{
				Granter: granterAddr,
				Grantee: granteeAddr,
				Grant:   grant,
			})

			return false
		})

		for _, g := range grants {
			author, err := g.Grant.GetAuthorization()
			if err != nil {
				return nil, err
			}

			msgTypeURL := author.MsgTypeURL()

			// Check if author is a generic authorization and oracle messages
			if genAuth, ok := author.(*authz.GenericAuthorization); ok && strings.HasPrefix(msgTypeURL, "/oracle.v1.") {
				newMsgTypeURL := strings.Replace(msgTypeURL, "/oracle.v1.", "/band.oracle.v1.", 1)
				genAuth.Msg = newMsgTypeURL

				// Delete the old grant
				err = keepers.AuthzKeeper.DeleteGrant(ctx, g.Granter, g.Grantee, msgTypeURL)
				if err != nil {
					return nil, err
				}

				// Save the new grant
				err = keepers.AuthzKeeper.SaveGrant(ctx, g.Granter, g.Grantee, genAuth, g.Grant.Expiration)
				if err != nil {
					return nil, err
				}
			}
		}

		return vm, nil
	}
}
