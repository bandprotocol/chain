package feechecker

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	bandtsskeeper "github.com/bandprotocol/chain/v2/x/bandtss/keeper"
	bandtsstypes "github.com/bandprotocol/chain/v2/x/bandtss/types"
	"github.com/bandprotocol/chain/v2/x/globalfee/keeper"
	oraclekeeper "github.com/bandprotocol/chain/v2/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v2/x/oracle/types"
	tsskeeper "github.com/bandprotocol/chain/v2/x/tss/keeper"
	tsstypes "github.com/bandprotocol/chain/v2/x/tss/types"
)

type FeeChecker struct {
	AuthzKeeper     *authzkeeper.Keeper
	OracleKeeper    *oraclekeeper.Keeper
	GlobalfeeKeeper *keeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
	TSSKeeper       *tsskeeper.Keeper
	BandTSSKeeper   *bandtsskeeper.Keeper

	TSSMsgServer       tsstypes.MsgServer
	TSSMemberMsgServer bandtsstypes.MsgServer
}

func NewFeeChecker(
	authzKeeper *authzkeeper.Keeper,
	oracleKeeper *oraclekeeper.Keeper,
	globalfeeKeeper *keeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
	tssKeeper *tsskeeper.Keeper,
	BandTSSKeeper *bandtsskeeper.Keeper,
) FeeChecker {
	tssMsgServer := tsskeeper.NewMsgServerImpl(tssKeeper)
	tssMemberMsgServer := bandtsskeeper.NewMsgServerImpl(BandTSSKeeper)

	return FeeChecker{
		AuthzKeeper:        authzKeeper,
		OracleKeeper:       oracleKeeper,
		GlobalfeeKeeper:    globalfeeKeeper,
		StakingKeeper:      stakingKeeper,
		TSSKeeper:          tssKeeper,
		BandTSSKeeper:      BandTSSKeeper,
		TSSMsgServer:       tssMsgServer,
		TSSMemberMsgServer: tssMemberMsgServer,
	}
}

func (fc FeeChecker) CheckTxFeeWithMinGasPrices(
	ctx sdk.Context,
	tx sdk.Tx,
) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.ErrTxDecode.Wrap("Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet minimum-gas-prices and globalFees,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() {
		// Check if this tx should be free or not
		isBypassMinFeeTx := fc.IsBypassMinFeeTx(ctx, tx)
		if isBypassMinFeeTx {
			return sdk.Coins{}, int64(math.MaxInt64), nil
		}

		minGasPrices := getMinGasPrices(ctx)
		globalMinGasPrices, err := fc.GetGlobalMinGasPrices(ctx)
		if err != nil {
			return nil, 0, err
		}

		allGasPrices := CombinedGasPricesRequirement(minGasPrices, globalMinGasPrices)

		// Calculate all fees from all gas prices
		gas := feeTx.GetGas()
		var allFees sdk.Coins
		if !allGasPrices.IsZero() {
			glDec := sdk.NewDec(int64(gas))
			for _, gp := range allGasPrices {
				if !gp.IsZero() {
					fee := gp.Amount.Mul(glDec)
					allFees = append(allFees, sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt()))
				}
			}
		}

		if !allFees.IsZero() && !feeCoins.IsAnyGTE(allFees) {
			return nil, 0, sdkerrors.ErrInsufficientFee.Wrapf(
				"insufficient fees; got: %s required: %s",
				feeCoins,
				allFees,
			)
		}
	}

	priority := getTxPriority(feeCoins, int64(gas), fc.GetBondDenom(ctx))
	return feeCoins, priority, nil
}

func (fc FeeChecker) IsBypassMinFeeTx(ctx sdk.Context, tx sdk.Tx) bool {
	newCtx, _ := ctx.CacheContext()

	// Check if all messages are free
	for _, msg := range tx.GetMsgs() {
		if !fc.IsBypassMinFeeMsg(newCtx, msg) {
			return false
		}
	}

	return true
}

func (fc FeeChecker) IsBypassMinFeeMsg(ctx sdk.Context, msg sdk.Msg) bool {
	switch msg := msg.(type) {
	case *oracletypes.MsgReportData:
		if err := checkValidMsgReport(ctx, fc.OracleKeeper, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitDKGRound1:
		if _, err := fc.TSSMsgServer.SubmitDKGRound1(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitDKGRound2:
		if _, err := fc.TSSMsgServer.SubmitDKGRound2(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgConfirm:
		if _, err := fc.TSSMsgServer.Confirm(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgComplain:
		if _, err := fc.TSSMsgServer.Complain(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitDEs:
		if _, err := fc.TSSMsgServer.SubmitDEs(ctx, msg); err != nil {
			return false
		}
	case *tsstypes.MsgSubmitSignature:
		if _, err := fc.TSSMsgServer.SubmitSignature(ctx, msg); err != nil {
			return false
		}
	case *bandtsstypes.MsgHealthCheck:
		if _, err := fc.TSSMemberMsgServer.HealthCheck(ctx, msg); err != nil {
			return false
		}
	case *authz.MsgExec:
		msgs, err := msg.GetMessages()
		if err != nil {
			return false
		}

		grantee, err := sdk.AccAddressFromBech32(msg.Grantee)
		if err != nil {
			return false
		}

		for _, m := range msgs {
			// Check if this grantee have authorization for the message.
			cap, _ := fc.AuthzKeeper.GetAuthorization(
				ctx,
				grantee,
				m.GetSigners()[0],
				sdk.MsgTypeURL(m),
			)
			if cap == nil {
				return false
			}

			// Check if this message should be free or not.
			if !fc.IsBypassMinFeeMsg(ctx, m) {
				return false
			}
		}
	default:
		return false
	}

	return true
}

func (fc FeeChecker) GetGlobalMinGasPrices(ctx sdk.Context) (sdk.DecCoins, error) {
	var (
		globalMinGasPrices sdk.DecCoins
		err                error
	)

	globalMinGasPrices = fc.GlobalfeeKeeper.GetParams(ctx).MinimumGasPrices
	// global fee is empty set, set global fee to 0uband (bondDenom)
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices, err = fc.DefaultZeroGlobalFee(ctx)
	}

	return globalMinGasPrices.Sort(), err
}

func (fc FeeChecker) DefaultZeroGlobalFee(ctx sdk.Context) ([]sdk.DecCoin, error) {
	bondDenom := fc.GetBondDenom(ctx)
	if bondDenom == "" {
		return nil, sdkerrors.ErrNotFound.Wrap("empty staking bond denomination")
	}

	return []sdk.DecCoin{sdk.NewDecCoinFromDec(bondDenom, sdk.NewDec(0))}, nil
}

func (fc FeeChecker) GetBondDenom(ctx sdk.Context) string {
	return fc.StakingKeeper.BondDenom(ctx)
}
