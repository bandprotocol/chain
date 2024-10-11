package benchmark

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	bandtesting "github.com/bandprotocol/chain/v2/testing"
	feedstypes "github.com/bandprotocol/chain/v2/x/feeds/types"
	"github.com/bandprotocol/chain/v2/x/tunnel/types"
)

func BenchmarkTunnelABCI(b *testing.B) {
	testcases := []struct {
		numTunnels  int
		numSignals  int
		maxSignals  int
		encoderType types.Encoder
	}{
		{1, 1, 1, types.ENCODER_FIXED_POINT_ABI},
		{1, 100, 100, types.ENCODER_FIXED_POINT_ABI},
		{10, 10, 100, types.ENCODER_FIXED_POINT_ABI},
		{10, 100, 100, types.ENCODER_FIXED_POINT_ABI},
		{100, 100, 100, types.ENCODER_FIXED_POINT_ABI},
		{1, 1, 1, types.ENCODER_TICK_ABI},
		{1, 100, 100, types.ENCODER_TICK_ABI},
		{10, 10, 100, types.ENCODER_TICK_ABI},
		{10, 100, 100, types.ENCODER_TICK_ABI},
		{100, 100, 100, types.ENCODER_TICK_ABI},
	}

	for _, tc := range testcases {
		f := testBenchmarkTunnel(tc.numTunnels, tc.numSignals, tc.maxSignals, tc.encoderType)
		benchmarkTestName := fmt.Sprintf(
			"TunnelABCI_%dTunnel_%dSignals_%dMaxSignal_%sEncoderType",
			tc.numTunnels, tc.numSignals, tc.maxSignals, tc.encoderType,
		)
		b.Run(benchmarkTestName, f)
	}
}

// testBenchmarkTunnel is a helper function to benchmark tunnel endblock process.
func testBenchmarkTunnel(numTunnels, numSignals, maxSignals int, encoder types.Encoder) func(b *testing.B) {
	return func(b *testing.B) {
		require.GreaterOrEqual(b, maxSignals, numSignals)
		require.NotEqual(b, types.ENCODER_UNSPECIFIED, encoder)

		ba := InitializeBenchmarkApp(b, -1)

		// set minDeposit to 1
		params := ba.TunnelKeeper.GetParams(ba.Ctx)
		params.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin("uband", 1))
		err := ba.TunnelKeeper.SetParams(ba.Ctx, params)
		require.NoError(b, err)

		var globalSignalDeviations []types.SignalDeviation
		for i := 1; i <= maxSignals; i++ {
			globalSignalDeviations = append(globalSignalDeviations, types.SignalDeviation{
				SignalID: fmt.Sprintf("test%d", i), SoftDeviationBPS: 100, HardDeviationBPS: 200,
			})
		}

		// create tunnel; for each tunnel, randomly pick signals from global signalInfos.
		for i := 1; i <= numTunnels; i++ {
			signalIdx := rand.Perm(maxSignals)
			var signalDeviations []types.SignalDeviation
			for j := 0; j < numSignals; j++ {
				signalDeviations = append(signalDeviations, globalSignalDeviations[signalIdx[j]])
			}

			err := createNewTunnels(ba, &types.TSSRoute{}, signalDeviations, encoder)
			require.NoError(b, err)
		}

		setupFeedsPrice(ba, globalSignalDeviations)

		b.ResetTimer()
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			ba.CallBeginBlock()
			err := shiftFeedsPrice(ba, globalSignalDeviations, 10500)
			require.NoError(b, err)

			tunnels := []types.Tunnel{}
			for j := 1; j <= numTunnels; j++ {
				tunnel := ba.TunnelKeeper.MustGetTunnel(ba.Ctx, uint64(j))
				tunnels = append(tunnels, tunnel)
			}

			b.StartTimer()
			ba.CallEndBlock()
			b.StopTimer()

			ba.Commit()

			// check result
			for j := 1; j <= numTunnels; j++ {
				newTunnel := ba.TunnelKeeper.MustGetTunnel(ba.Ctx, uint64(j))
				require.Equal(b, tunnels[j-1].Sequence+1, newTunnel.Sequence)
				require.True(b, newTunnel.IsActive)
			}
		}
	}
}

// createNewTunnels creates new tunnels with given signalInfos and encoder.
func createNewTunnels(
	ba *BenchmarkApp,
	route types.RouteI,
	signalDeviations []types.SignalDeviation,
	encoder types.Encoder,
) error {
	creator := bandtesting.Alice.Address
	tunnel, err := ba.TunnelKeeper.AddTunnel(
		ba.Ctx, route, encoder, signalDeviations, 1000, creator,
	)
	if err != nil {
		return err
	}

	depositor := bandtesting.Bob.Address
	minDeposit := ba.TunnelKeeper.GetParams(ba.Ctx).MinDeposit
	if err := ba.TunnelKeeper.AddDeposit(ba.Ctx, tunnel.ID, depositor, minDeposit); err != nil {
		return err
	}

	if err := ba.TunnelKeeper.ActivateTunnel(ba.Ctx, tunnel.ID); err != nil {
		return err
	}

	depositAmt := sdk.NewCoins(sdk.NewInt64Coin("uband", 50000))
	if err := ba.BankKeeper.SendCoins(
		ba.Ctx,
		bandtesting.Validators[0].Address,
		sdk.MustAccAddressFromBech32(tunnel.FeePayer),
		depositAmt,
	); err != nil {
		return err
	}

	return nil
}

// setupFeedsPrice sets up feeds and prices for benchmarking.
func setupFeedsPrice(ba *BenchmarkApp, signalDeviations []types.SignalDeviation) {
	var feeds []feedstypes.Feed
	var prices []feedstypes.Price
	for i, sd := range signalDeviations {
		feeds = append(feeds, feedstypes.Feed{
			SignalID: sd.SignalID, Power: 1000, Interval: 1000,
		})
		prices = append(prices, feedstypes.Price{
			SignalID: sd.SignalID, Price: uint64(i+1) * 1000, Timestamp: ba.Ctx.BlockTime().Unix(),
		})
	}

	ba.FeedsKeeper.SetCurrentFeeds(ba.Ctx, feeds)
	ba.FeedsKeeper.SetPrices(ba.Ctx, prices)
}

// shiftFeedsPrice shifts current feeds price by given multiplier.
func shiftFeedsPrice(ba *BenchmarkApp, signalDeviations []types.SignalDeviation, mltpyBps uint64) error {
	for _, sd := range signalDeviations {
		p, err := ba.FeedsKeeper.GetPrice(ba.Ctx, sd.SignalID)
		if err != nil {
			return err
		}

		p.Price = p.Price * mltpyBps / 10000
		ba.FeedsKeeper.SetPrice(ba.Ctx, p)
	}

	return nil
}
