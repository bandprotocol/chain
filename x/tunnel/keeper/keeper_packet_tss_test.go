package keeper_test

import (
	"time"

	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	bandtesting "github.com/bandprotocol/chain/v3/testing"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	"github.com/bandprotocol/chain/v3/x/tunnel/types"
)

func (s *KeeperTestSuite) TestSendTSSPacket() {
	ctx, k := s.ctx, s.keeper

	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}
	packet := types.NewPacket(
		1,                    // tunnelID
		1,                    // sequence
		[]feedstypes.Price{}, // priceInfos[]
		sdk.NewCoins(),       // baseFee
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), // routeFee
		time.Now().Unix(),
	)

	// Mock the TSS keeper and set the state for checking later
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"chain-1",
		"0x1234567890abcdef",
		gomock.Any(),
		bandtesting.Alice.Address,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
	).DoAndReturn(func(
		ctx sdk.Context,
		tunnelID uint64,
		destinationChainID string,
		destinationContractAddr string,
		content tsstypes.Content,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error) {
		ctx.KVStore(s.storeKey).Set([]byte{0xff, 0xff}, []byte("test"))
		return bandtsstypes.SigningID(1), nil
	})

	// Send the TSS packet
	content, err := k.SendTSSPacket(ctx, &route, packet, bandtesting.Alice.Address)
	s.Require().NoError(err)

	packetReceipt, ok := content.(*types.TSSPacketReceipt)
	s.Require().True(ok)
	s.Require().Equal(bandtsstypes.SigningID(1), packetReceipt.SigningID)

	s.Require().Equal([]byte("test"), ctx.KVStore(s.storeKey).Get([]byte{0xff, 0xff}))
}

func (s *KeeperTestSuite) TestFailSendTSSPacket() {
	ctx, k := s.ctx, s.keeper

	route := types.TSSRoute{
		DestinationChainID:         "chain-1",
		DestinationContractAddress: "0x1234567890abcdef",
	}
	packet := types.NewPacket(
		1,                    // tunnelID
		1,                    // sequence
		[]feedstypes.Price{}, // priceInfos[]
		sdk.NewCoins(),       // baseFee
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))), // routeFee
		time.Now().Unix(),
	)

	// Mock the TSS keeper and set the state for checking later
	s.bandtssKeeper.EXPECT().CreateTunnelSigningRequest(
		gomock.Any(),
		uint64(1),
		"chain-1",
		"0x1234567890abcdef",
		gomock.Any(),
		bandtesting.Alice.Address,
		sdk.NewCoins(sdk.NewCoin("uband", sdkmath.NewInt(20))),
	).DoAndReturn(func(
		ctx sdk.Context,
		tunnelID uint64,
		destinationChainID string,
		destinationContractAddr string,
		content tsstypes.Content,
		sender sdk.AccAddress,
		feeLimit sdk.Coins,
	) (bandtsstypes.SigningID, error) {
		ctx.KVStore(s.storeKey).Set([]byte{0xff, 0xff}, []byte("test"))
		return bandtsstypes.SigningID(0), tsstypes.ErrInsufficientSigners
	})

	// Send the TSS packet
	content, err := k.SendTSSPacket(ctx, &route, packet, bandtesting.Alice.Address)
	s.Require().ErrorIs(err, tsstypes.ErrInsufficientSigners)
	s.Require().Nil(content)

	s.Require().False(ctx.KVStore(s.storeKey).Has([]byte{0xff, 0xff}))
}
