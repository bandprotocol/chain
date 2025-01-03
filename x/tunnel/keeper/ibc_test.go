package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v3/x/tunnel/keeper"
)

func (s *KeeperTestSuite) TestPortIDForTunnel() {
	expPortID := "tunnel.1"
	portID := keeper.PortIDForTunnel(1)
	s.Require().Equal(expPortID, portID)
}

func (s *KeeperTestSuite) TestIsValidPortID() {
	testCases := []struct {
		name      string
		portID    string
		expResult bool
	}{
		{
			name:      "valid portID",
			portID:    "tunnel.1",
			expResult: true,
		},
		{
			name:      "without dot",
			portID:    "tunnel1",
			expResult: false,
		},
		{
			name:      "invalid prefix",
			portID:    "tun.1",
			expResult: false,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.name), func() {
			valid := keeper.IsValidPortID(tc.portID)
			s.Require().Equal(tc.expResult, valid)
		})
	}
}
