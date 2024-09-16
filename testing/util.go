package testing

import (
	"fmt"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	bankkeeper "github.com/bandprotocol/chain/v3/x/bank/keeper"
)

// ParseTime is a helper function to parse from number to time.Time with UTC locale.
func ParseTime(t int64) time.Time {
	return time.Unix(t, 0).UTC()
}

type GasRecord struct {
	Gas        storetypes.Gas
	Descriptor string
}

// GasMeterWrapper wrap gas meter for testing purpose
type GasMeterWrapper struct {
	storetypes.GasMeter
	GasRecords []GasRecord
}

func (m *GasMeterWrapper) ConsumeGas(amount storetypes.Gas, descriptor string) {
	m.GasRecords = append(m.GasRecords, GasRecord{amount, descriptor})
	m.GasMeter.ConsumeGas(amount, descriptor)
}

func (m *GasMeterWrapper) CountRecord(amount storetypes.Gas, descriptor string) int {
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
func NewGasMeterWrapper(meter storetypes.GasMeter) *GasMeterWrapper {
	return &GasMeterWrapper{meter, nil}
}

func MustGetBalances(ctx sdk.Context, bankKeeper bankkeeper.WrappedBankKeeper, address sdk.AccAddress) sdk.Coins {
	balancesRes, err := bankKeeper.AllBalances(
		ctx,
		banktypes.NewQueryAllBalancesRequest(address, &query.PageRequest{}, false),
	)
	if err != nil {
		panic(err)
	}

	return balancesRes.Balances
}

func CheckBalances(
	t *testing.T,
	ctx sdk.Context,
	bankKeeper bankkeeper.WrappedBankKeeper,
	address sdk.AccAddress,
	expected sdk.Coins,
) {
	balancesRes, err := bankKeeper.AllBalances(
		ctx,
		banktypes.NewQueryAllBalancesRequest(address, &query.PageRequest{}, false),
	)
	require.NoError(t, err)

	require.True(t, expected.Equal(balancesRes.Balances))
}

// CheckErrorf checks whether
// - error type is wrapped inside the given error
// - error match given message string combined with error type
func CheckErrorf(t *testing.T, err error, errType error, msg string, a ...interface{}) {
	require.ErrorIs(t, err, errType)
	formattedMsg := fmt.Sprintf(msg, a...)
	require.EqualError(t, err, fmt.Sprintf("%s: %s", formattedMsg, errType))
}
