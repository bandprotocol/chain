package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/suite"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	"github.com/bandprotocol/chain/v2/x/tunnel/keeper"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
)

// Keeper of the x/tunnel store
type KeeperTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	feedsKeeper keeper.Keeper
	queryClient types.QueryClient
	msgSrvr     types.MsgServer
	authority   sdk.AccAddress
}

func (s *KeeperTestSuite) SetupTest() {
	app, ctx := bandtesting.CreateTestApp(s.T(), true)
	s.ctx = ctx
	s.feedsKeeper = app.TunnelKeeper

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServer(app.TunnelKeeper))
	queryClient := types.NewQueryClient(queryHelper)
	s.queryClient = queryClient
	s.msgSrvr = keeper.NewMsgServerImpl(app.TunnelKeeper)
	s.authority = authtypes.NewModuleAddress(govtypes.ModuleName)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
