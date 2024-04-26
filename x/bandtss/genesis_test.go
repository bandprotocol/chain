package bandtss_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/bandtss"
	"github.com/bandprotocol/chain/v2/x/bandtss/testutil"
	"github.com/bandprotocol/chain/v2/x/bandtss/types"
)

func TestExportGenesis(t *testing.T) {
	s := testutil.NewTestSuite(t)
	ctx, k := s.Ctx, s.Keeper

	data := types.GenesisState{
		Params: types.DefaultParams(),
		Members: []types.Member{
			{Address: testapp.Alice.Address.String(), Since: ctx.BlockTime(), IsActive: true, LastActive: ctx.BlockTime()},
		},
		CurrentGroupID: tss.GroupID(1),
		SigningCount:   1,
		Signings: []types.Signing{
			{ID: types.SigningID(1), Requester: testapp.Alice.Address.String(), CurrentGroupSigningID: tss.SigningID(3)},
		},
		SigningIDMappings: []types.SigningIDMappingGenesis{
			{SigningID: tss.SigningID(1), BandtssSigningID: types.SigningID(3)},
		},
		Replacement: types.Replacement{
			SigningID:      tss.SigningID(1),
			CurrentGroupID: tss.GroupID(1),
			NewGroupID:     tss.GroupID(2),
			Status:         types.REPLACEMENT_STATUS_WAITING_SIGNING,
		},
	}

	bandtss.InitGenesis(ctx, k, &data)

	exported := bandtss.ExportGenesis(ctx, k)
	require.Equal(t, data.Params, exported.Params)
}
