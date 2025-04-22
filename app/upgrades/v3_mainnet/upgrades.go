package v3_mainnet

import (
	"context"

	cmttypes "github.com/cometbft/cometbft/types"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

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

		var mintParams minttypes.Params
		if has, err := keepers.MintKeeper.Params.Has(ctx); has && err != nil {
			mintParams, err = keepers.MintKeeper.Params.Get(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			mintParams = minttypes.DefaultParams()
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

		vm, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}

		err = keepers.GlobalFeeKeeper.SetParams(ctx, globalfeetypes.Params{
			MinimumGasPrices: sdk.DecCoins{sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(25, 4))},
		})
		if err != nil {
			return nil, err
		}

		return vm, nil
	}
}
