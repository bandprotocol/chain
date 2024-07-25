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

	"github.com/bandprotocol/chain/v2/x/restake/keeper"
	restaketestutil "github.com/bandprotocol/chain/v2/x/restake/testutil"
	"github.com/bandprotocol/chain/v2/x/restake/types"
)

var (
	// delegate power
	// - 1e18 -> address 1,2
	// - 10   -> address 3
	ValidAddress1 = sdk.AccAddress("1000000001")
	ValidAddress2 = sdk.AccAddress("1000000002")
	ValidAddress3 = sdk.AccAddress("1000000003")

	// status
	// - active   -> key 1,2,4
	// - inactive -> key 3
	// key 4 has total power as zero
	ValidKey1  = "key1"
	ValidKey2  = "key2"
	ValidKey3  = "key3"
	ValidKey4  = "key4"
	InvalidKey = "nonKey"

	ValidPoolAddress1 = sdk.AccAddress("2000000001")
	ValidPoolAddress2 = sdk.AccAddress("2000000002")
	ValidPoolAddress3 = sdk.AccAddress("2000000003")
	ValidPoolAddress4 = sdk.AccAddress("2000000004")

	RewarderAddress = sdk.AccAddress("3000000001")
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
	suite.validKeys = []types.Key{
		{
			Name:            ValidKey1,
			PoolAddress:     ValidPoolAddress1.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(20),
			RewardPerPowers: sdk.NewDecCoins(sdk.NewDecCoinFromDec("uband", sdkmath.LegacyNewDecWithPrec(1, 1))),
			Remainders:      nil,
		},
		{
			Name:            ValidKey2,
			PoolAddress:     ValidPoolAddress2.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            ValidKey3,
			PoolAddress:     ValidPoolAddress3.String(),
			IsActive:        false,
			TotalPower:      sdkmath.NewInt(100),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
		{
			Name:            ValidKey4,
			PoolAddress:     ValidPoolAddress4.String(),
			IsActive:        true,
			TotalPower:      sdkmath.NewInt(0),
			RewardPerPowers: nil,
			Remainders:      nil,
		},
	}

	suite.validLocks = []types.Lock{
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            ValidKey1,
			Amount:         sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            ValidKey2,
			Amount:         sdkmath.NewInt(100),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress1.String(),
			Key:            ValidKey3,
			Amount:         sdkmath.NewInt(50),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
		{
			LockerAddress:  ValidAddress2.String(),
			Key:            ValidKey1,
			Amount:         sdkmath.NewInt(10),
			PosRewardDebts: nil,
			NegRewardDebts: nil,
		},
	}

	suite.resetState()
}

func (suite *KeeperTestSuite) resetState() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	suite.storeKey = key
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig()

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	accountKeeper := restaketestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().
		GetAccount(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	accountKeeper.EXPECT().
		NewAccount(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	accountKeeper.EXPECT().
		SetAccount(gomock.Any(), gomock.Any()).
		Return().
		AnyTimes()
	suite.accountKeeper = accountKeeper

	bankKeeper := restaketestutil.NewMockBankKeeper(ctrl)
	bankKeeper.EXPECT().
		SendCoins(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	suite.bankKeeper = bankKeeper

	stakingKeeper := restaketestutil.NewMockStakingKeeper(ctrl)
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress1).
		Return(sdkmath.NewInt(1e18)).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress2).
		Return(sdkmath.NewInt(1e18)).
		AnyTimes()
	stakingKeeper.EXPECT().
		GetDelegatorBonded(gomock.Any(), ValidAddress3).
		Return(sdkmath.NewInt(10)).
		AnyTimes()
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

func (suite *KeeperTestSuite) setupState() {
	for _, key := range suite.validKeys {
		suite.restakeKeeper.SetKey(suite.ctx, key)
	}

	for _, lock := range suite.validLocks {
		suite.restakeKeeper.SetLock(suite.ctx, lock)
	}
}
