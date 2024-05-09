// Code generated by MockGen. DO NOT EDIT.
// Source: x/feeds/types/expected_keepers.go
//
// Generated by this command:
//
//	mockgen -source=x/feeds/types/expected_keepers.go -package testutil -destination x/feeds/testutil/expected_keepers_mocks.go
//

// Package testutil is a generated GoMock package.
package testutil

import (
	reflect "reflect"
	time "time"

	math "cosmossdk.io/math"
	types "github.com/bandprotocol/chain/v2/x/oracle/types"
	types0 "github.com/cosmos/cosmos-sdk/types"
	types1 "github.com/cosmos/cosmos-sdk/x/staking/types"
	gomock "go.uber.org/mock/gomock"
)

// MockOracleKeeper is a mock of OracleKeeper interface.
type MockOracleKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockOracleKeeperMockRecorder
}

// MockOracleKeeperMockRecorder is the mock recorder for MockOracleKeeper.
type MockOracleKeeperMockRecorder struct {
	mock *MockOracleKeeper
}

// NewMockOracleKeeper creates a new mock instance.
func NewMockOracleKeeper(ctrl *gomock.Controller) *MockOracleKeeper {
	mock := &MockOracleKeeper{ctrl: ctrl}
	mock.recorder = &MockOracleKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOracleKeeper) EXPECT() *MockOracleKeeperMockRecorder {
	return m.recorder
}

// GetValidatorStatus mocks base method.
func (m *MockOracleKeeper) GetValidatorStatus(ctx types0.Context, val types0.ValAddress) types.ValidatorStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidatorStatus", ctx, val)
	ret0, _ := ret[0].(types.ValidatorStatus)
	return ret0
}

// GetValidatorStatus indicates an expected call of GetValidatorStatus.
func (mr *MockOracleKeeperMockRecorder) GetValidatorStatus(ctx, val any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidatorStatus", reflect.TypeOf((*MockOracleKeeper)(nil).GetValidatorStatus), ctx, val)
}

// MissReport mocks base method.
func (m *MockOracleKeeper) MissReport(ctx types0.Context, val types0.ValAddress, requestTime time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "MissReport", ctx, val, requestTime)
}

// MissReport indicates an expected call of MissReport.
func (mr *MockOracleKeeperMockRecorder) MissReport(ctx, val, requestTime any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MissReport", reflect.TypeOf((*MockOracleKeeper)(nil).MissReport), ctx, val, requestTime)
}

// MockStakingKeeper is a mock of StakingKeeper interface.
type MockStakingKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockStakingKeeperMockRecorder
}

// MockStakingKeeperMockRecorder is the mock recorder for MockStakingKeeper.
type MockStakingKeeperMockRecorder struct {
	mock *MockStakingKeeper
}

// NewMockStakingKeeper creates a new mock instance.
func NewMockStakingKeeper(ctrl *gomock.Controller) *MockStakingKeeper {
	mock := &MockStakingKeeper{ctrl: ctrl}
	mock.recorder = &MockStakingKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStakingKeeper) EXPECT() *MockStakingKeeperMockRecorder {
	return m.recorder
}

// GetDelegation mocks base method.
func (m *MockStakingKeeper) GetDelegation(ctx types0.Context, delAddr types0.AccAddress, valAddr types0.ValAddress) (types1.Delegation, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDelegation", ctx, delAddr, valAddr)
	ret0, _ := ret[0].(types1.Delegation)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetDelegation indicates an expected call of GetDelegation.
func (mr *MockStakingKeeperMockRecorder) GetDelegation(ctx, delAddr, valAddr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDelegation", reflect.TypeOf((*MockStakingKeeper)(nil).GetDelegation), ctx, delAddr, valAddr)
}

// GetDelegatorBonded mocks base method.
func (m *MockStakingKeeper) GetDelegatorBonded(ctx types0.Context, delegator types0.AccAddress) math.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDelegatorBonded", ctx, delegator)
	ret0, _ := ret[0].(math.Int)
	return ret0
}

// GetDelegatorBonded indicates an expected call of GetDelegatorBonded.
func (mr *MockStakingKeeperMockRecorder) GetDelegatorBonded(ctx, delegator any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDelegatorBonded", reflect.TypeOf((*MockStakingKeeper)(nil).GetDelegatorBonded), ctx, delegator)
}

// GetValidator mocks base method.
func (m *MockStakingKeeper) GetValidator(ctx types0.Context, addr types0.ValAddress) (types1.Validator, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValidator", ctx, addr)
	ret0, _ := ret[0].(types1.Validator)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetValidator indicates an expected call of GetValidator.
func (mr *MockStakingKeeperMockRecorder) GetValidator(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValidator", reflect.TypeOf((*MockStakingKeeper)(nil).GetValidator), ctx, addr)
}

// IterateBondedValidatorsByPower mocks base method.
func (m *MockStakingKeeper) IterateBondedValidatorsByPower(ctx types0.Context, fn func(int64, types1.ValidatorI) bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "IterateBondedValidatorsByPower", ctx, fn)
}

// IterateBondedValidatorsByPower indicates an expected call of IterateBondedValidatorsByPower.
func (mr *MockStakingKeeperMockRecorder) IterateBondedValidatorsByPower(ctx, fn any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IterateBondedValidatorsByPower", reflect.TypeOf((*MockStakingKeeper)(nil).IterateBondedValidatorsByPower), ctx, fn)
}
