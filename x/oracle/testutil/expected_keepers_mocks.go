// Code generated by MockGen. DO NOT EDIT.
// Source: x/oracle/types/expected_keepers.go
//
// Generated by this command:
//
//	mockgen -source=x/oracle/types/expected_keepers.go -package testutil -destination x/oracle/testutil/expected_keepers_mocks.go
//

// Package testutil is a generated GoMock package.
package testutil

import (
	context "context"
	reflect "reflect"
	time "time"

	math "cosmossdk.io/math"
	types "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	types0 "github.com/cosmos/cosmos-sdk/x/staking/types"
	types1 "github.com/cosmos/ibc-go/modules/capability/types"
	gomock "go.uber.org/mock/gomock"
)

// MockAccountKeeper is a mock of AccountKeeper interface.
type MockAccountKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockAccountKeeperMockRecorder
	isgomock struct{}
}

// MockAccountKeeperMockRecorder is the mock recorder for MockAccountKeeper.
type MockAccountKeeperMockRecorder struct {
	mock *MockAccountKeeper
}

// NewMockAccountKeeper creates a new mock instance.
func NewMockAccountKeeper(ctrl *gomock.Controller) *MockAccountKeeper {
	mock := &MockAccountKeeper{ctrl: ctrl}
	mock.recorder = &MockAccountKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccountKeeper) EXPECT() *MockAccountKeeperMockRecorder {
	return m.recorder
}

// GetAccount mocks base method.
func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr types.AccAddress) types.AccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", ctx, addr)
	ret0, _ := ret[0].(types.AccountI)
	return ret0
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockAccountKeeperMockRecorder) GetAccount(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockAccountKeeper)(nil).GetAccount), ctx, addr)
}

// GetModuleAccount mocks base method.
func (m *MockAccountKeeper) GetModuleAccount(ctx context.Context, moduleName string) types.ModuleAccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModuleAccount", ctx, moduleName)
	ret0, _ := ret[0].(types.ModuleAccountI)
	return ret0
}

// GetModuleAccount indicates an expected call of GetModuleAccount.
func (mr *MockAccountKeeperMockRecorder) GetModuleAccount(ctx, moduleName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModuleAccount", reflect.TypeOf((*MockAccountKeeper)(nil).GetModuleAccount), ctx, moduleName)
}

// MockBankKeeper is a mock of BankKeeper interface.
type MockBankKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockBankKeeperMockRecorder
	isgomock struct{}
}

// MockBankKeeperMockRecorder is the mock recorder for MockBankKeeper.
type MockBankKeeperMockRecorder struct {
	mock *MockBankKeeper
}

// NewMockBankKeeper creates a new mock instance.
func NewMockBankKeeper(ctrl *gomock.Controller) *MockBankKeeper {
	mock := &MockBankKeeper{ctrl: ctrl}
	mock.recorder = &MockBankKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBankKeeper) EXPECT() *MockBankKeeperMockRecorder {
	return m.recorder
}

// GetAllBalances mocks base method.
func (m *MockBankKeeper) GetAllBalances(ctx context.Context, addr types.AccAddress) types.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllBalances", ctx, addr)
	ret0, _ := ret[0].(types.Coins)
	return ret0
}

// GetAllBalances indicates an expected call of GetAllBalances.
func (mr *MockBankKeeperMockRecorder) GetAllBalances(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllBalances", reflect.TypeOf((*MockBankKeeper)(nil).GetAllBalances), ctx, addr)
}

// SendCoins mocks base method.
func (m *MockBankKeeper) SendCoins(ctx context.Context, from, to types.AccAddress, amt types.Coins) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCoins", ctx, from, to, amt)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCoins indicates an expected call of SendCoins.
func (mr *MockBankKeeperMockRecorder) SendCoins(ctx, from, to, amt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCoins", reflect.TypeOf((*MockBankKeeper)(nil).SendCoins), ctx, from, to, amt)
}

// SendCoinsFromModuleToModule mocks base method.
func (m *MockBankKeeper) SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt types.Coins) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCoinsFromModuleToModule", ctx, senderModule, recipientModule, amt)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCoinsFromModuleToModule indicates an expected call of SendCoinsFromModuleToModule.
func (mr *MockBankKeeperMockRecorder) SendCoinsFromModuleToModule(ctx, senderModule, recipientModule, amt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCoinsFromModuleToModule", reflect.TypeOf((*MockBankKeeper)(nil).SendCoinsFromModuleToModule), ctx, senderModule, recipientModule, amt)
}

// SpendableCoins mocks base method.
func (m *MockBankKeeper) SpendableCoins(ctx context.Context, addr types.AccAddress) types.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SpendableCoins", ctx, addr)
	ret0, _ := ret[0].(types.Coins)
	return ret0
}

// SpendableCoins indicates an expected call of SpendableCoins.
func (mr *MockBankKeeperMockRecorder) SpendableCoins(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SpendableCoins", reflect.TypeOf((*MockBankKeeper)(nil).SpendableCoins), ctx, addr)
}

// MockStakingKeeper is a mock of StakingKeeper interface.
type MockStakingKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockStakingKeeperMockRecorder
	isgomock struct{}
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

// IterateBondedValidatorsByPower mocks base method.
func (m *MockStakingKeeper) IterateBondedValidatorsByPower(arg0 context.Context, arg1 func(int64, types0.ValidatorI) bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IterateBondedValidatorsByPower", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// IterateBondedValidatorsByPower indicates an expected call of IterateBondedValidatorsByPower.
func (mr *MockStakingKeeperMockRecorder) IterateBondedValidatorsByPower(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IterateBondedValidatorsByPower", reflect.TypeOf((*MockStakingKeeper)(nil).IterateBondedValidatorsByPower), arg0, arg1)
}

// Validator mocks base method.
func (m *MockStakingKeeper) Validator(arg0 context.Context, arg1 types.ValAddress) (types0.ValidatorI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validator", arg0, arg1)
	ret0, _ := ret[0].(types0.ValidatorI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Validator indicates an expected call of Validator.
func (mr *MockStakingKeeperMockRecorder) Validator(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validator", reflect.TypeOf((*MockStakingKeeper)(nil).Validator), arg0, arg1)
}

// ValidatorByConsAddr mocks base method.
func (m *MockStakingKeeper) ValidatorByConsAddr(arg0 context.Context, arg1 types.ConsAddress) (types0.ValidatorI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatorByConsAddr", arg0, arg1)
	ret0, _ := ret[0].(types0.ValidatorI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidatorByConsAddr indicates an expected call of ValidatorByConsAddr.
func (mr *MockStakingKeeperMockRecorder) ValidatorByConsAddr(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatorByConsAddr", reflect.TypeOf((*MockStakingKeeper)(nil).ValidatorByConsAddr), arg0, arg1)
}

// MockDistrKeeper is a mock of DistrKeeper interface.
type MockDistrKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockDistrKeeperMockRecorder
	isgomock struct{}
}

// MockDistrKeeperMockRecorder is the mock recorder for MockDistrKeeper.
type MockDistrKeeperMockRecorder struct {
	mock *MockDistrKeeper
}

// NewMockDistrKeeper creates a new mock instance.
func NewMockDistrKeeper(ctrl *gomock.Controller) *MockDistrKeeper {
	mock := &MockDistrKeeper{ctrl: ctrl}
	mock.recorder = &MockDistrKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDistrKeeper) EXPECT() *MockDistrKeeperMockRecorder {
	return m.recorder
}

// AllocateTokensToValidator mocks base method.
func (m *MockDistrKeeper) AllocateTokensToValidator(ctx context.Context, val types0.ValidatorI, tokens types.DecCoins) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AllocateTokensToValidator", ctx, val, tokens)
	ret0, _ := ret[0].(error)
	return ret0
}

// AllocateTokensToValidator indicates an expected call of AllocateTokensToValidator.
func (mr *MockDistrKeeperMockRecorder) AllocateTokensToValidator(ctx, val, tokens any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AllocateTokensToValidator", reflect.TypeOf((*MockDistrKeeper)(nil).AllocateTokensToValidator), ctx, val, tokens)
}

// FundCommunityPool mocks base method.
func (m *MockDistrKeeper) FundCommunityPool(ctx context.Context, amount types.Coins, sender types.AccAddress) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FundCommunityPool", ctx, amount, sender)
	ret0, _ := ret[0].(error)
	return ret0
}

// FundCommunityPool indicates an expected call of FundCommunityPool.
func (mr *MockDistrKeeperMockRecorder) FundCommunityPool(ctx, amount, sender any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FundCommunityPool", reflect.TypeOf((*MockDistrKeeper)(nil).FundCommunityPool), ctx, amount, sender)
}

// GetCommunityTax mocks base method.
func (m *MockDistrKeeper) GetCommunityTax(ctx context.Context) (math.LegacyDec, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCommunityTax", ctx)
	ret0, _ := ret[0].(math.LegacyDec)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCommunityTax indicates an expected call of GetCommunityTax.
func (mr *MockDistrKeeperMockRecorder) GetCommunityTax(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCommunityTax", reflect.TypeOf((*MockDistrKeeper)(nil).GetCommunityTax), ctx)
}

// MockPortKeeper is a mock of PortKeeper interface.
type MockPortKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockPortKeeperMockRecorder
	isgomock struct{}
}

// MockPortKeeperMockRecorder is the mock recorder for MockPortKeeper.
type MockPortKeeperMockRecorder struct {
	mock *MockPortKeeper
}

// NewMockPortKeeper creates a new mock instance.
func NewMockPortKeeper(ctrl *gomock.Controller) *MockPortKeeper {
	mock := &MockPortKeeper{ctrl: ctrl}
	mock.recorder = &MockPortKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPortKeeper) EXPECT() *MockPortKeeperMockRecorder {
	return m.recorder
}

// BindPort mocks base method.
func (m *MockPortKeeper) BindPort(ctx types.Context, portID string) *types1.Capability {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BindPort", ctx, portID)
	ret0, _ := ret[0].(*types1.Capability)
	return ret0
}

// BindPort indicates an expected call of BindPort.
func (mr *MockPortKeeperMockRecorder) BindPort(ctx, portID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BindPort", reflect.TypeOf((*MockPortKeeper)(nil).BindPort), ctx, portID)
}

// MockAuthzKeeper is a mock of AuthzKeeper interface.
type MockAuthzKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockAuthzKeeperMockRecorder
	isgomock struct{}
}

// MockAuthzKeeperMockRecorder is the mock recorder for MockAuthzKeeper.
type MockAuthzKeeperMockRecorder struct {
	mock *MockAuthzKeeper
}

// NewMockAuthzKeeper creates a new mock instance.
func NewMockAuthzKeeper(ctrl *gomock.Controller) *MockAuthzKeeper {
	mock := &MockAuthzKeeper{ctrl: ctrl}
	mock.recorder = &MockAuthzKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAuthzKeeper) EXPECT() *MockAuthzKeeperMockRecorder {
	return m.recorder
}

// DeleteGrant mocks base method.
func (m *MockAuthzKeeper) DeleteGrant(ctx context.Context, grantee, granter types.AccAddress, msgType string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteGrant", ctx, grantee, granter, msgType)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteGrant indicates an expected call of DeleteGrant.
func (mr *MockAuthzKeeperMockRecorder) DeleteGrant(ctx, grantee, granter, msgType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteGrant", reflect.TypeOf((*MockAuthzKeeper)(nil).DeleteGrant), ctx, grantee, granter, msgType)
}

// DispatchActions mocks base method.
func (m *MockAuthzKeeper) DispatchActions(ctx context.Context, grantee types.AccAddress, msgs []types.Msg) ([][]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DispatchActions", ctx, grantee, msgs)
	ret0, _ := ret[0].([][]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DispatchActions indicates an expected call of DispatchActions.
func (mr *MockAuthzKeeperMockRecorder) DispatchActions(ctx, grantee, msgs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DispatchActions", reflect.TypeOf((*MockAuthzKeeper)(nil).DispatchActions), ctx, grantee, msgs)
}

// GetAuthorization mocks base method.
func (m *MockAuthzKeeper) GetAuthorization(ctx context.Context, grantee, granter types.AccAddress, msgType string) (authz.Authorization, *time.Time) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuthorization", ctx, grantee, granter, msgType)
	ret0, _ := ret[0].(authz.Authorization)
	ret1, _ := ret[1].(*time.Time)
	return ret0, ret1
}

// GetAuthorization indicates an expected call of GetAuthorization.
func (mr *MockAuthzKeeperMockRecorder) GetAuthorization(ctx, grantee, granter, msgType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuthorization", reflect.TypeOf((*MockAuthzKeeper)(nil).GetAuthorization), ctx, grantee, granter, msgType)
}

// GetAuthorizations mocks base method.
func (m *MockAuthzKeeper) GetAuthorizations(ctx context.Context, grantee, granter types.AccAddress) ([]authz.Authorization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAuthorizations", ctx, grantee, granter)
	ret0, _ := ret[0].([]authz.Authorization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAuthorizations indicates an expected call of GetAuthorizations.
func (mr *MockAuthzKeeperMockRecorder) GetAuthorizations(ctx, grantee, granter any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAuthorizations", reflect.TypeOf((*MockAuthzKeeper)(nil).GetAuthorizations), ctx, grantee, granter)
}

// GranterGrants mocks base method.
func (m *MockAuthzKeeper) GranterGrants(ctx context.Context, req *authz.QueryGranterGrantsRequest) (*authz.QueryGranterGrantsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GranterGrants", ctx, req)
	ret0, _ := ret[0].(*authz.QueryGranterGrantsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GranterGrants indicates an expected call of GranterGrants.
func (mr *MockAuthzKeeperMockRecorder) GranterGrants(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GranterGrants", reflect.TypeOf((*MockAuthzKeeper)(nil).GranterGrants), ctx, req)
}

// SaveGrant mocks base method.
func (m *MockAuthzKeeper) SaveGrant(ctx context.Context, grantee, granter types.AccAddress, authorization authz.Authorization, expiration *time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveGrant", ctx, grantee, granter, authorization, expiration)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveGrant indicates an expected call of SaveGrant.
func (mr *MockAuthzKeeperMockRecorder) SaveGrant(ctx, grantee, granter, authorization, expiration any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveGrant", reflect.TypeOf((*MockAuthzKeeper)(nil).SaveGrant), ctx, grantee, granter, authorization, expiration)
}
