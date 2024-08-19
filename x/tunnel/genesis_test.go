package tunnel_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/bandprotocol/chain/v2/x/tunnel"
	"github.com/bandprotocol/chain/v2/x/tunnel/testutil"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func TestValidateGenesis(t *testing.T) {
	// Create a valid genesis state
	validGenesisState := &types.GenesisState{
		PortID:      types.PortID,
		Params:      types.DefaultParams(),
		TunnelCount: 1,
		Tunnels: []types.Tunnel{
			{ID: 1},
		},
	}

	// Test with valid genesis state
	err := tunnel.ValidateGenesis(validGenesisState)
	require.NoError(t, err)

	// Test with invalid tunnel count
	invalidGenesisState := &types.GenesisState{
		Params:      types.DefaultParams(),
		TunnelCount: 2,
		Tunnels: []types.Tunnel{
			{ID: 1},
		},
	}
	err = tunnel.ValidateGenesis(invalidGenesisState)
	require.Error(t, err)

	// Test with invalid tunnel IDs
	invalidGenesisState = &types.GenesisState{
		Params:      types.DefaultParams(),
		TunnelCount: 1,
		Tunnels: []types.Tunnel{
			{ID: 2},
		},
	}
	err = tunnel.ValidateGenesis(invalidGenesisState)
	require.Error(t, err)
}

func TestInitExportGenesis(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	// Mock the account keeper
	s.MockAccountKeeper.EXPECT().
		GetModuleAccount(ctx, gomock.Any()).
		Return(authtypes.AccountI(&authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{Address: "test"},
		})).
		AnyTimes()
	s.MockAccountKeeper.EXPECT().GetModuleAddress(types.ModuleName).Return(sdk.AccAddress{}).AnyTimes()
	s.MockAccountKeeper.EXPECT().SetModuleAccount(ctx, gomock.Any()).AnyTimes()
	s.MockBankKeeper.EXPECT().GetAllBalances(ctx, gomock.Any()).Return(sdk.Coins{}).AnyTimes()
	s.MockScopedKeeper.EXPECT().GetCapability(ctx, gomock.Any()).Return(nil, true).AnyTimes()

	// Create a valid genesis state
	genesisState := &types.GenesisState{
		PortID:      types.PortID,
		Params:      types.DefaultParams(),
		TunnelCount: 1,
		Tunnels: []types.Tunnel{
			{ID: 1},
		},
	}

	// Initialize the genesis state
	tunnel.InitGenesis(ctx, k, genesisState)

	// Export the genesis state
	exportedGenesisState := tunnel.ExportGenesis(ctx, k)

	// Verify the exported state matches the initialized state
	require.Equal(t, genesisState, exportedGenesisState)
}
