package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	app "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/x/restake/keeper"
	restaketestutil "github.com/bandprotocol/chain/v2/x/restake/testutil"
	"github.com/bandprotocol/chain/v2/x/restake/types"
)

var (
	ValidAddress1     = sdk.AccAddress("1000000001")
	ValidAddress2     = sdk.AccAddress("1000000002")
	ValidAddress3     = sdk.AccAddress("1000000003")
	ValidPoolAddress1 = sdk.AccAddress("2000000001")
	ValidPoolAddress2 = sdk.AccAddress("2000000002")
)

type KeeperTestSuite struct {
	suite.Suite

	ctx           sdk.Context
	storeKey      storetypes.StoreKey
	restakeKeeper keeper.Keeper
	accountKeeper *restaketestutil.MockAccountKeeper
	bankKeeper    *restaketestutil.MockBankKeeper
	stakingKeeper *restaketestutil.MockStakingKeeper

	queryClient types.QueryClient
	msgServer   types.MsgServer

	validKeys  []types.Key
	validLocks []types.Lock
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app.SetBech32AddressPrefixesAndBip44CoinTypeAndSeal(sdk.GetConfig())

	suite.validKeys = []types.Key{
		{
			Name:            "Key0",
			PoolAddress:     ValidPoolAddress1.String(),
			IsActive:        true,
			TotalPower:      sdk.NewInt(20),
			RewardPerPowers: sdk.NewDecCoins(sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(1, 1))),
			Remainders:      sdk.NewDecCoins(),
		},
		{
			Name:            "Key1",
			PoolAddress:     ValidPoolAddress2.String(),
			IsActive:        true,
			TotalPower:      sdk.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
	}

	suite.validLocks = []types.Lock{
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            "Key0",
			Amount:         sdk.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            "Key1",
			Amount:         sdk.NewInt(100),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress2.String(),
			Key:            "Key0",
			Amount:         sdk.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
	}

	key := sdk.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := restaketestutil.NewMockAccountKeeper(ctrl)
	suite.accountKeeper = accountKeeper

	bankKeeper := restaketestutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().
		SendCoins(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	suite.bankKeeper = bankKeeper

	stakingKeeper := restaketestutil.NewMockStakingKeeper(ctrl)
	suite.stakingKeeper = stakingKeeper

	suite.restakeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
	)
	suite.restakeKeeper.InitGenesis(suite.ctx, types.DefaultGenesisState())

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, encCfg.InterfaceRegistry)
	queryServer := keeper.Querier{
		Keeper: &suite.restakeKeeper,
	}

	types.RegisterInterfaces(encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, queryServer)
	queryClient := types.NewQueryClient(queryHelper)
	suite.queryClient = queryClient
	suite.msgServer = keeper.NewMsgServerImpl(&suite.restakeKeeper)
}
