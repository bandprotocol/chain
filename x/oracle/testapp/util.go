package testapp

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (m *GasMeterWrapper) GasConsumed() sdk.Gas {
	return m.GasMeter.GasConsumed()
}

func (m *GasMeterWrapper) GasConsumedToLimit() sdk.Gas {
	return m.GasMeter.GasConsumedToLimit()
}

func (m *GasMeterWrapper) Limit() sdk.Gas {
	return m.GasMeter.Limit()
}

func (m *GasMeterWrapper) ConsumeGas(amount sdk.Gas, descriptor string) {
	m.GasRecords = append(m.GasRecords, GasRecord{amount, descriptor})
	m.GasMeter.ConsumeGas(amount, descriptor)
}

func (m *GasMeterWrapper) IsPastLimit() bool {
	return m.GasMeter.IsPastLimit()
}

func (m *GasMeterWrapper) IsOutOfGas() bool {
	return m.GasMeter.IsOutOfGas()
}

func (m *GasMeterWrapper) String() string {
	return m.GasMeter.String()
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
