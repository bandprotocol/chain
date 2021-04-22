package testapp

import (
	"testing"
	"time"

	bandbankkeeper "github.com/GeoDB-Limited/odin-core/x/bank/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

// ParseTime is a helper function to parse from number to time.Time with UTC locale.
func ParseTime(t int64) time.Time {
	return time.Unix(t, 0).UTC()
}

type GasRecord struct {
	Gas        sdk.Gas
	Descriptor string
}

// GasMeterWrapper wrap gas meter for testing purpose
type GasMeterWrapper struct {
	sdk.GasMeter
	GasRecords []GasRecord
}

func (m *GasMeterWrapper) ConsumeGas(amount sdk.Gas, descriptor string) {
	m.GasRecords = append(m.GasRecords, GasRecord{amount, descriptor})
	m.GasMeter.ConsumeGas(amount, descriptor)
}

func (m *GasMeterWrapper) CountRecord(amount sdk.Gas, descriptor string) int {
	count := 0
	for _, r := range m.GasRecords {
		if r.Gas == amount && r.Descriptor == descriptor {
			count++
		}
	}

	return count
}

func (m *GasMeterWrapper) CountDescriptor(descriptor string) int {
	count := 0
	for _, r := range m.GasRecords {
		if r.Descriptor == descriptor {
			count++
		}
	}

	return count
}

// NewGasMeterWrapper to wrap gas meters for testing purposes
func NewGasMeterWrapper(meter sdk.GasMeter) *GasMeterWrapper {
	return &GasMeterWrapper{meter, nil}
}

func MustGetBalances(ctx sdk.Context, bankKeeper *bandbankkeeper.WrappedBankKeeper, address sdk.AccAddress) sdk.Coins {
	balancesRes, err := bankKeeper.AllBalances(sdk.WrapSDKContext(ctx), banktypes.NewQueryAllBalancesRequest(address, &query.PageRequest{}))
	if err != nil {
		panic(err)
	}

	return balancesRes.Balances
}

func CheckBalances(
	t *testing.T,
	ctx sdk.Context,
	bankKeeper *bandbankkeeper.WrappedBankKeeper,
	address sdk.AccAddress,
	expected sdk.Coins,
) {
	balancesRes, err := bankKeeper.AllBalances(sdk.WrapSDKContext(ctx), banktypes.NewQueryAllBalancesRequest(address, &query.PageRequest{}))
	require.NoError(t, err)

	require.True(t, expected.IsEqual(balancesRes.Balances))
}
