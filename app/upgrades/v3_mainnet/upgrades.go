package v3_mainnet

import (
	"context"

	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	cmttypes "github.com/cometbft/cometbft/types"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

		// Set new consensus params with same values as before
		consensusParams := cmttypes.DefaultConsensusParams().ToProto()
		consensusParams.Block.MaxBytes = BlockMaxBytes                                     // unchanged
		consensusParams.Block.MaxGas = BlockMaxGas                                         // unchanged
		consensusParams.Validator.PubKeyTypes = []string{cmttypes.ABCIPubKeyTypeSecp256k1} // unchanged
		err := keepers.ConsensusParamsKeeper.ParamsStore.Set(ctx, consensusParams)
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

		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		feemarketParams := feemarkettypes.DefaultParams()
		feemarketParams.MaxBlockUtilization = uint64(BlockMaxGas)
		feemarketParams.MinBaseGasPrice = MinimumGasPrice
		feemarketParams.FeeDenom = Denom
		feemarketParams.Enabled = false
		err = keepers.FeeMarketKeeper.SetParams(ctx, feemarketParams)
		if err != nil {
			return nil, err
		}

		state, err := keepers.FeeMarketKeeper.GetState(ctx)
		if err != nil {
			return nil, err
		}

		state.BaseGasPrice = MinimumGasPrice
		err = keepers.FeeMarketKeeper.SetState(ctx, state)
		if err != nil {
			return nil, err
		}

		return vm, nil
	}
}
