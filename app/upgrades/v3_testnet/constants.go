package v3

import (
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
)

// UpgradeName defines the on-chain upgrade name.
const (
	UpgradeName = "v3"

	// BlockMaxBytes is the max bytes for a block, 3mb
	BlockMaxBytes = int64(3000000)

	// BlockMaxGas is the max gas allowed in a block
	BlockMaxGas = int64(50000000)
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades:        storetypes.StoreUpgrades{},
}

// TODO: Update ICA Allow messages
var ICAAllowMessages = []string{
	sdk.MsgTypeURL(&authz.MsgExec{}),
	sdk.MsgTypeURL(&authz.MsgGrant{}),
	sdk.MsgTypeURL(&authz.MsgRevoke{}),
	sdk.MsgTypeURL(&banktypes.MsgSend{}),
	sdk.MsgTypeURL(&banktypes.MsgMultiSend{}),
	sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
	sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
	sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
	sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
	sdk.MsgTypeURL(&feegrant.MsgGrantAllowance{}),
	sdk.MsgTypeURL(&feegrant.MsgRevokeAllowance{}),
	sdk.MsgTypeURL(&govv1beta1.MsgVoteWeighted{}),
	sdk.MsgTypeURL(&govv1beta1.MsgSubmitProposal{}),
	sdk.MsgTypeURL(&govv1beta1.MsgDeposit{}),
	sdk.MsgTypeURL(&govv1beta1.MsgVote{}),
	// Change: add messages from Oracle module
	sdk.MsgTypeURL(&oracletypes.MsgActivate{}),
	sdk.MsgTypeURL(&oracletypes.MsgCreateDataSource{}),
	sdk.MsgTypeURL(&oracletypes.MsgCreateOracleScript{}),
	sdk.MsgTypeURL(&oracletypes.MsgEditDataSource{}),
	sdk.MsgTypeURL(&oracletypes.MsgEditOracleScript{}),
	sdk.MsgTypeURL(&oracletypes.MsgReportData{}),
	sdk.MsgTypeURL(&oracletypes.MsgRequestData{}),

	sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
	sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
	sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
	sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
}
