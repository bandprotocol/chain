package keeper_test

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func (s *KeeperTestSuite) TestExportGenesis() {
	ctx, k := s.ctx, s.keeper

	addr1 := bandtesting.Alice.Address
	addr2 := bandtesting.Bob.Address

	data := types.GenesisState{
		Params: types.DefaultParams(),
		Groups: []types.Group{
			{
				ID:        1,
				Size_:     1,
				Threshold: 1,
				PubKey:    nil,
				Status:    types.GROUP_STATUS_ROUND_1,
			},
		},
		Members: []types.Member{
			{
				ID:          1,
				GroupID:     1,
				Address:     addr1.String(),
				PubKey:      nil,
				IsMalicious: false,
				IsActive:    true,
			},
		},
		DEs: []types.DEGenesis{
			{
				Address: addr1.String(),
				DE: types.DE{
					PubD: []byte("pubD"),
					PubE: []byte("pubE"),
				},
			},
			{
				Address: addr1.String(),
				DE: types.DE{
					PubD: []byte("pubD2"),
					PubE: []byte("pubE2"),
				},
			},
			{
				Address: addr2.String(),
				DE: types.DE{
					PubD: []byte("pubD3"),
					PubE: []byte("pubE3"),
				},
			},
		},
	}

	k.InitGenesis(ctx, data)

	exported := k.ExportGenesis(ctx)
	s.Require().Equal(data.Params, exported.Params)
	s.Require().Equal(uint64(1), k.GetGroupCount(ctx))

	sort.Slice(exported.DEs, func(i, j int) bool {
		if exported.DEs[i].Address != exported.DEs[j].Address {
			return exported.DEs[i].Address < exported.DEs[j].Address
		}
		return i < j
	})

	sort.Slice(data.DEs, func(i, j int) bool {
		if data.DEs[i].Address != data.DEs[j].Address {
			return data.DEs[i].Address < data.DEs[j].Address
		}
		return i < j
	})
	s.Require().Equal(data.DEs, exported.DEs)

	s.Require().Equal(types.DEQueue{Head: 0, Tail: 2}, k.GetDEQueue(ctx, addr1))
	existingDEs := []types.DE{}
	for i := 0; i < 2; i++ {
		de, err := k.GetDE(ctx, addr1, uint64(i))
		s.Require().NoError(err)
		existingDEs = append(existingDEs, de)
	}
	s.Require().Equal(existingDEs, []types.DE{
		{
			PubD: []byte("pubD"),
			PubE: []byte("pubE"),
		},
		{
			PubD: []byte("pubD2"),
			PubE: []byte("pubE2"),
		},
	})

	k.SetDEQueue(ctx, addr1, types.DEQueue{Head: 1, Tail: 2})

	exported = k.ExportGenesis(ctx)
	s.Require().Len(exported.DEs, 2)
	s.Require().Contains(exported.DEs, types.DEGenesis{
		Address: addr1.String(),
		DE: types.DE{
			PubD: []byte("pubD2"),
			PubE: []byte("pubE2"),
		},
	})
	s.Require().Contains(exported.DEs, types.DEGenesis{
		Address: addr2.String(),
		DE: types.DE{
			PubD: []byte("pubD3"),
			PubE: []byte("pubE3"),
		},
	})
}

func (s *KeeperTestSuite) TestGetDEsGenesis() {
	ctx, k := s.ctx, s.keeper

	address := sdk.MustAccAddressFromBech32("band1m5lq9u533qaya4q3nfyl6ulzqkpkhge9q8tpzs")
	before := k.GetDEsGenesis(ctx)
	de := types.DE{
		PubD: []byte("D"),
		PubE: []byte("E"),
	}

	// Set DE
	err := k.EnqueueDEs(ctx, address, []types.DE{de})
	s.Require().NoError(err)

	// Get des with address and index
	after := k.GetDEsGenesis(ctx)
	s.Require().Equal(len(before)+1, len(after))
	for _, q := range after {
		if q.Address == string(address) {
			expected := types.DEGenesis{Address: address.String(), DE: de}
			s.Require().Equal(expected, q)
		}
	}
}
