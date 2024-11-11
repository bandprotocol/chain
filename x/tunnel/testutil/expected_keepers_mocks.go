// Code generated by MockGen. DO NOT EDIT.
// Source: x/tunnel/types/expected_keepers.go
//
// Generated by this command:
//
//	mockgen -source=x/tunnel/types/expected_keepers.go -package testutil -destination x/tunnel/testutil/expected_keepers_mocks.go
//

// Package testutil is a generated GoMock package.
package testutil

import (
	context "context"
	reflect "reflect"

	types "github.com/bandprotocol/chain/v3/x/bandtss/types"
	types0 "github.com/bandprotocol/chain/v3/x/feeds/types"
	types1 "github.com/bandprotocol/chain/v3/x/tss/types"
	types2 "github.com/cosmos/cosmos-sdk/types"
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
func (m *MockAccountKeeper) GetAccount(ctx context.Context, addr types2.AccAddress) types2.AccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", ctx, addr)
	ret0, _ := ret[0].(types2.AccountI)
	return ret0
}

// GetAccount indicates an expected call of GetAccount.
func (mr *MockAccountKeeperMockRecorder) GetAccount(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockAccountKeeper)(nil).GetAccount), ctx, addr)
}

// GetModuleAccount mocks base method.
func (m *MockAccountKeeper) GetModuleAccount(ctx context.Context, name string) types2.ModuleAccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModuleAccount", ctx, name)
	ret0, _ := ret[0].(types2.ModuleAccountI)
	return ret0
}

// GetModuleAccount indicates an expected call of GetModuleAccount.
func (mr *MockAccountKeeperMockRecorder) GetModuleAccount(ctx, name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModuleAccount", reflect.TypeOf((*MockAccountKeeper)(nil).GetModuleAccount), ctx, name)
}

// GetModuleAddress mocks base method.
func (m *MockAccountKeeper) GetModuleAddress(name string) types2.AccAddress {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModuleAddress", name)
	ret0, _ := ret[0].(types2.AccAddress)
	return ret0
}

// GetModuleAddress indicates an expected call of GetModuleAddress.
func (mr *MockAccountKeeperMockRecorder) GetModuleAddress(name any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModuleAddress", reflect.TypeOf((*MockAccountKeeper)(nil).GetModuleAddress), name)
}

// NewAccount mocks base method.
func (m *MockAccountKeeper) NewAccount(ctx context.Context, account types2.AccountI) types2.AccountI {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewAccount", ctx, account)
	ret0, _ := ret[0].(types2.AccountI)
	return ret0
}

// NewAccount indicates an expected call of NewAccount.
func (mr *MockAccountKeeperMockRecorder) NewAccount(ctx, account any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewAccount", reflect.TypeOf((*MockAccountKeeper)(nil).NewAccount), ctx, account)
}

// SetAccount mocks base method.
func (m *MockAccountKeeper) SetAccount(ctx context.Context, account types2.AccountI) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetAccount", ctx, account)
}

// SetAccount indicates an expected call of SetAccount.
func (mr *MockAccountKeeperMockRecorder) SetAccount(ctx, account any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAccount", reflect.TypeOf((*MockAccountKeeper)(nil).SetAccount), ctx, account)
}

// SetModuleAccount mocks base method.
func (m *MockAccountKeeper) SetModuleAccount(ctx context.Context, moduleAccount types2.ModuleAccountI) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetModuleAccount", ctx, moduleAccount)
}

// SetModuleAccount indicates an expected call of SetModuleAccount.
func (mr *MockAccountKeeperMockRecorder) SetModuleAccount(ctx, moduleAccount any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetModuleAccount", reflect.TypeOf((*MockAccountKeeper)(nil).SetModuleAccount), ctx, moduleAccount)
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
func (m *MockBankKeeper) GetAllBalances(ctx context.Context, addr types2.AccAddress) types2.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllBalances", ctx, addr)
	ret0, _ := ret[0].(types2.Coins)
	return ret0
}

// GetAllBalances indicates an expected call of GetAllBalances.
func (mr *MockBankKeeperMockRecorder) GetAllBalances(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllBalances", reflect.TypeOf((*MockBankKeeper)(nil).GetAllBalances), ctx, addr)
}

// SendCoinsFromAccountToModule mocks base method.
func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr types2.AccAddress, recipientModule string, amt types2.Coins) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCoinsFromAccountToModule", ctx, senderAddr, recipientModule, amt)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCoinsFromAccountToModule indicates an expected call of SendCoinsFromAccountToModule.
func (mr *MockBankKeeperMockRecorder) SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCoinsFromAccountToModule", reflect.TypeOf((*MockBankKeeper)(nil).SendCoinsFromAccountToModule), ctx, senderAddr, recipientModule, amt)
}

// SendCoinsFromModuleToAccount mocks base method.
func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr types2.AccAddress, amt types2.Coins) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCoinsFromModuleToAccount", ctx, senderModule, recipientAddr, amt)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCoinsFromModuleToAccount indicates an expected call of SendCoinsFromModuleToAccount.
func (mr *MockBankKeeperMockRecorder) SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, amt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCoinsFromModuleToAccount", reflect.TypeOf((*MockBankKeeper)(nil).SendCoinsFromModuleToAccount), ctx, senderModule, recipientAddr, amt)
}

// SpendableCoins mocks base method.
func (m *MockBankKeeper) SpendableCoins(ctx context.Context, addr types2.AccAddress) types2.Coins {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SpendableCoins", ctx, addr)
	ret0, _ := ret[0].(types2.Coins)
	return ret0
}

// SpendableCoins indicates an expected call of SpendableCoins.
func (mr *MockBankKeeperMockRecorder) SpendableCoins(ctx, addr any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SpendableCoins", reflect.TypeOf((*MockBankKeeper)(nil).SpendableCoins), ctx, addr)
}

// MockFeedsKeeper is a mock of FeedsKeeper interface.
type MockFeedsKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockFeedsKeeperMockRecorder
	isgomock struct{}
}

// MockFeedsKeeperMockRecorder is the mock recorder for MockFeedsKeeper.
type MockFeedsKeeperMockRecorder struct {
	mock *MockFeedsKeeper
}

// NewMockFeedsKeeper creates a new mock instance.
func NewMockFeedsKeeper(ctrl *gomock.Controller) *MockFeedsKeeper {
	mock := &MockFeedsKeeper{ctrl: ctrl}
	mock.recorder = &MockFeedsKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFeedsKeeper) EXPECT() *MockFeedsKeeperMockRecorder {
	return m.recorder
}

// GetAllPrices mocks base method.
func (m *MockFeedsKeeper) GetAllPrices(ctx types2.Context) []types0.Price {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllPrices", ctx)
	ret0, _ := ret[0].([]types0.Price)
	return ret0
}

// GetAllPrices indicates an expected call of GetAllPrices.
func (mr *MockFeedsKeeperMockRecorder) GetAllPrices(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllPrices", reflect.TypeOf((*MockFeedsKeeper)(nil).GetAllPrices), ctx)
}

// GetPrices mocks base method.
func (m *MockFeedsKeeper) GetPrices(ctx types2.Context, signalIDs []string) []types0.Price {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPrices", ctx, signalIDs)
	ret0, _ := ret[0].([]types0.Price)
	return ret0
}

// GetPrices indicates an expected call of GetPrices.
func (mr *MockFeedsKeeperMockRecorder) GetPrices(ctx, signalIDs any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPrices", reflect.TypeOf((*MockFeedsKeeper)(nil).GetPrices), ctx, signalIDs)
}

// MockBandtssKeeper is a mock of BandtssKeeper interface.
type MockBandtssKeeper struct {
	ctrl     *gomock.Controller
	recorder *MockBandtssKeeperMockRecorder
	isgomock struct{}
}

// MockBandtssKeeperMockRecorder is the mock recorder for MockBandtssKeeper.
type MockBandtssKeeperMockRecorder struct {
	mock *MockBandtssKeeper
}

// NewMockBandtssKeeper creates a new mock instance.
func NewMockBandtssKeeper(ctrl *gomock.Controller) *MockBandtssKeeper {
	mock := &MockBandtssKeeper{ctrl: ctrl}
	mock.recorder = &MockBandtssKeeperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBandtssKeeper) EXPECT() *MockBandtssKeeperMockRecorder {
	return m.recorder
}

// CreateTunnelSigningRequest mocks base method.
func (m *MockBandtssKeeper) CreateTunnelSigningRequest(ctx types2.Context, tunnelID uint64, destinationContractAddr, destinationChainID string, content types1.Content, sender types2.AccAddress, feeLimit types2.Coins) (types.SigningID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTunnelSigningRequest", ctx, tunnelID, destinationContractAddr, destinationChainID, content, sender, feeLimit)
	ret0, _ := ret[0].(types.SigningID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTunnelSigningRequest indicates an expected call of CreateTunnelSigningRequest.
func (mr *MockBandtssKeeperMockRecorder) CreateTunnelSigningRequest(ctx, tunnelID, destinationContractAddr, destinationChainID, content, sender, feeLimit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTunnelSigningRequest", reflect.TypeOf((*MockBandtssKeeper)(nil).CreateTunnelSigningRequest), ctx, tunnelID, destinationContractAddr, destinationChainID, content, sender, feeLimit)
}

// GetParams mocks base method.
func (m *MockBandtssKeeper) GetParams(ctx types2.Context) types.Params {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetParams", ctx)
	ret0, _ := ret[0].(types.Params)
	return ret0
}

// GetParams indicates an expected call of GetParams.
func (mr *MockBandtssKeeperMockRecorder) GetParams(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetParams", reflect.TypeOf((*MockBandtssKeeper)(nil).GetParams), ctx)
}