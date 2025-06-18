package v3

import (
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"

	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/app/upgrades"
	bandtsstypes "github.com/bandprotocol/chain/v3/x/bandtss/types"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	globalfeetypes "github.com/bandprotocol/chain/v3/x/globalfee/types"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	rollingseedtypes "github.com/bandprotocol/chain/v3/x/rollingseed/types"
	tsstypes "github.com/bandprotocol/chain/v3/x/tss/types"
	tunneltypes "github.com/bandprotocol/chain/v3/x/tunnel/types"
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
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			globalfeetypes.StoreKey,
			consensusparamtypes.StoreKey,
			crisistypes.StoreKey,
			ibcfeetypes.StoreKey,
			restaketypes.StoreKey,
			feedstypes.StoreKey,
			rollingseedtypes.StoreKey,
			bandtsstypes.StoreKey,
			tsstypes.StoreKey,
			tunneltypes.StoreKey,
		},
	},
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
	sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
	sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
	sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
	sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
	sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
	// Change: add messages from Oracle module
	sdk.MsgTypeURL(&oracletypes.MsgCreateDataSource{}),
	sdk.MsgTypeURL(&oracletypes.MsgCreateOracleScript{}),
	sdk.MsgTypeURL(&oracletypes.MsgEditDataSource{}),
	sdk.MsgTypeURL(&oracletypes.MsgEditOracleScript{}),
	sdk.MsgTypeURL(&oracletypes.MsgRequestData{}),
}
