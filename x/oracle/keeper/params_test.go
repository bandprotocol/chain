package keeper_test

import (
	"fmt"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

func (suite *KeeperTestSuite) TestGetSetParams() {
	ctx := suite.ctx
	k := suite.oracleKeeper
	require := suite.Require()

	expectedParams := types.Params{
		MaxRawRequestCount:      1,
		MaxAskCount:             10,
		MaxCalldataSize:         256,
		MaxReportDataSize:       512,
		ExpirationBlockCount:    30,
		BaseOwasmGas:            50000,
		PerValidatorRequestGas:  3000,
		SamplingTryCount:        3,
		OracleRewardPercentage:  50,
		InactivePenaltyDuration: 1000,
		IBCRequestEnabled:       true,
	}
	err := k.SetParams(ctx, expectedParams)
	require.NoError(err)
	require.Equal(expectedParams, k.GetParams(ctx))

	expectedParams = types.Params{
		MaxRawRequestCount:      2,
		MaxAskCount:             20,
		MaxCalldataSize:         512,
		MaxReportDataSize:       256,
		ExpirationBlockCount:    40,
		BaseOwasmGas:            150000,
		PerValidatorRequestGas:  30000,
		SamplingTryCount:        5,
		OracleRewardPercentage:  80,
		InactivePenaltyDuration: 10000,
		IBCRequestEnabled:       false,
	}
	err = k.SetParams(ctx, expectedParams)
	require.NoError(err)
	require.Equal(expectedParams, k.GetParams(ctx))

	expectedParams = types.Params{
		MaxRawRequestCount:      2,
		MaxAskCount:             20,
		MaxCalldataSize:         512,
		MaxReportDataSize:       256,
		ExpirationBlockCount:    40,
		BaseOwasmGas:            0,
		PerValidatorRequestGas:  0,
		SamplingTryCount:        5,
		OracleRewardPercentage:  0,
		InactivePenaltyDuration: 0,
		IBCRequestEnabled:       false,
	}
	err = k.SetParams(ctx, expectedParams)
	require.NoError(err)
	require.Equal(expectedParams, k.GetParams(ctx))

	expectedParams = types.Params{
		MaxRawRequestCount:      0,
		MaxAskCount:             20,
		MaxCalldataSize:         512,
		MaxReportDataSize:       256,
		ExpirationBlockCount:    40,
		BaseOwasmGas:            150000,
		PerValidatorRequestGas:  30000,
		SamplingTryCount:        5,
		OracleRewardPercentage:  80,
		InactivePenaltyDuration: 10000,
		IBCRequestEnabled:       false,
	}
	err = k.SetParams(ctx, expectedParams)
	require.EqualError(fmt.Errorf("max raw request count must be positive: 0"), err.Error())
}
