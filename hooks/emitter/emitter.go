package emitter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmjson "github.com/cometbft/cometbft/libs/json"
	types1 "github.com/cometbft/cometbft/proto/tendermint/types"

	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	connectionkeeper "github.com/cosmos/ibc-go/v8/modules/core/03-connection/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bandprotocol/chain/v3/hooks/common"
	bandtsskeeper "github.com/bandprotocol/chain/v3/x/bandtss/keeper"
	feedskeeper "github.com/bandprotocol/chain/v3/x/feeds/keeper"
	feedstypes "github.com/bandprotocol/chain/v3/x/feeds/types"
	oraclekeeper "github.com/bandprotocol/chain/v3/x/oracle/keeper"
	oracletypes "github.com/bandprotocol/chain/v3/x/oracle/types"
	restakekeeper "github.com/bandprotocol/chain/v3/x/restake/keeper"
	restaketypes "github.com/bandprotocol/chain/v3/x/restake/types"
	tsskeeper "github.com/bandprotocol/chain/v3/x/tss/keeper"
)

// Hook uses Kafka functionality to act as an event producer for all events in the blockchains.
type Hook struct {
	cdc      codec.Codec
	txConfig client.TxConfig
	// Main Kafka writer instance.
	writer *kafka.Writer
	// Temporary variables that are reset on every block.
	accsInBlock    map[string]bool  // The accounts that need balance update at the end of block.
	accsInTx       map[string]bool  // The accounts related to the current processing transaction.
	msgs           []common.Message // The list of all messages to publish for this block.
	emitStartState bool             // If emitStartState is true will emit all non historical state to Kafka

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	mintKeeper    mintkeeper.Keeper
	distrKeeper   distrkeeper.Keeper
	govKeeper     *govkeeper.Keeper
	groupKeeper   groupkeeper.Keeper
	oracleKeeper  oraclekeeper.Keeper
	restakeKeeper restakekeeper.Keeper
	tssKeeper     *tsskeeper.Keeper
	bandtssKeeper bandtsskeeper.Keeper
	feedsKeeper   feedskeeper.Keeper
	icahostKeeper icahostkeeper.Keeper

	// ibc keeper
	clientKeeper     clientkeeper.Keeper
	connectionKeeper connectionkeeper.Keeper
	channelKeeper    channelkeeper.Keeper

	groupStoreKey storetypes.StoreKey
}

// NewHook creates an emitter hook instance that will be added in Band App.
func NewHook(
	cdc codec.Codec,
	txConfig client.TxConfig,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	govKeeper *govkeeper.Keeper,
	groupKeeper groupkeeper.Keeper,
	oracleKeeper oraclekeeper.Keeper,
	restakeKeeper restakekeeper.Keeper,
	feedsKeeper feedskeeper.Keeper,
	tssKeeper *tsskeeper.Keeper,
	bandtssKeeper bandtsskeeper.Keeper,
	icahostKeeper icahostkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	groupStorekey storetypes.StoreKey,
	kafkaURI string,
	emitStartState bool,
) *Hook {
	paths := strings.SplitN(kafkaURI, "@", 2)
	return &Hook{
		cdc:      cdc,
		txConfig: txConfig,
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:      paths[1:],
			Topic:        paths[0],
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 1 * time.Millisecond,
			// Async:    true, // TODO: We may be able to enable async mode on replay
		}),
		accountKeeper:    accountKeeper,
		bankKeeper:       bankKeeper,
		stakingKeeper:    stakingKeeper,
		mintKeeper:       mintKeeper,
		distrKeeper:      distrKeeper,
		govKeeper:        govKeeper,
		groupKeeper:      groupKeeper,
		oracleKeeper:     oracleKeeper,
		restakeKeeper:    restakeKeeper,
		feedsKeeper:      feedsKeeper,
		tssKeeper:        tssKeeper,
		bandtssKeeper:    bandtssKeeper,
		icahostKeeper:    icahostKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		channelKeeper:    channelKeeper,
		groupStoreKey:    groupStorekey,
		emitStartState:   emitStartState,
	}
}

// AddAccountsInBlock adds the given accounts to the list of accounts to update balances end-of-block.
func (h *Hook) AddAccountsInBlock(accs ...string) {
	for _, acc := range accs {
		h.accsInBlock[acc] = true
	}
}

// AddAccountsInTx adds the given accounts to the list of accounts to track related account in transaction.
func (h *Hook) AddAccountsInTx(accs ...string) {
	for _, acc := range accs {
		h.accsInTx[acc] = true
	}
}

// Write adds the given key-value pair to the list of messages to publish during Commit.
func (h *Hook) Write(key string, val common.JsDict) {
	h.msgs = append(h.msgs, common.Message{Key: key, Value: val})
}

// FlushMessages publishes all pending messages to Kafka. Blocks until completion.
func (h *Hook) FlushMessages() {
	kafkaMsgs := make([]kafka.Message, len(h.msgs))
	for idx, msg := range h.msgs {
		res, _ := json.Marshal(msg.Value) // Error must always be nil.
		kafkaMsgs[idx] = kafka.Message{Key: []byte(msg.Key), Value: res}
	}
	err := h.writer.WriteMessages(context.Background(), kafkaMsgs...)
	if err != nil {
		panic(err)
	}
}

// AfterInitChain specify actions need to do after chain initialization (app.Hook interface).
func (h *Hook) AfterInitChain(ctx sdk.Context, req *abci.RequestInitChain, res *abci.ResponseInitChain) {
	var genesisState map[string]json.RawMessage
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	// Auth module
	var bankGenesis banktypes.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[banktypes.ModuleName], &bankGenesis)

	for _, account := range bankGenesis.Balances {
		h.Write("SET_ACCOUNT", common.JsDict{
			"address": account.Address,
			"balance": account.GetCoins().String(),
		})
	}
	// GenUtil module for create validator genesis transactions.
	var genutilState genutiltypes.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[genutiltypes.ModuleName], &genutilState)
	for _, genTx := range genutilState.GenTxs {
		var tx sdk.Tx
		tx, err := h.txConfig.TxJSONDecoder()(genTx)
		if err != nil {
			panic(err)
		}
		for _, msg := range tx.GetMsgs() {
			if msg, ok := msg.(*stakingtypes.MsgCreateValidator); ok {
				h.handleMsgCreateValidator(ctx, msg, make(common.JsDict))
			}
		}
	}

	// Staking module
	var stakingState stakingtypes.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[stakingtypes.ModuleName], &stakingState)
	for _, val := range stakingState.Validators {
		h.emitSetValidator(ctx, MustParseValAddress(val.GetOperator()))
	}

	for _, del := range stakingState.Delegations {
		valAddr, _ := sdk.ValAddressFromBech32(del.ValidatorAddress)
		delAddr, _ := sdk.AccAddressFromBech32(del.DelegatorAddress)
		h.emitDelegation(ctx, valAddr, delAddr)
	}

	for _, unbonding := range stakingState.UnbondingDelegations {
		for _, entry := range unbonding.Entries {
			h.Write("NEW_UNBONDING_DELEGATION", common.JsDict{
				"delegator_address": unbonding.DelegatorAddress,
				"operator_address":  unbonding.ValidatorAddress,
				"completion_time":   entry.CompletionTime.UnixNano(),
				"amount":            entry.Balance,
			})
		}
	}

	for _, redelegate := range stakingState.Redelegations {
		for _, entry := range redelegate.Entries {
			h.Write("NEW_REDELEGATION", common.JsDict{
				"delegator_address":    redelegate.DelegatorAddress,
				"operator_src_address": redelegate.ValidatorSrcAddress,
				"operator_dst_address": redelegate.ValidatorDstAddress,
				"completion_time":      entry.CompletionTime.UnixNano(),
				"amount":               entry.InitialBalance,
			})
		}
	}

	// Gov module
	var govState govv1.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[govtypes.ModuleName], &govState)
	for _, proposal := range govState.Proposals {
		msgs, _ := proposal.GetMsgs()
		switch subMsg := msgs[0].(type) {
		case *govv1.MsgExecLegacyContent:
			content := subMsg.Content.GetCachedValue().(govv1beta1.Content)
			h.Write("NEW_PROPOSAL", common.JsDict{
				"id":               proposal.Id,
				"proposer":         nil,
				"type":             content.ProposalType(),
				"title":            content.GetTitle(),
				"description":      content.GetDescription(),
				"proposal_route":   content.ProposalRoute(),
				"status":           int(proposal.Status),
				"submit_time":      common.TimeToNano(proposal.SubmitTime),
				"deposit_end_time": common.TimeToNano(proposal.DepositEndTime),
				"total_deposit":    sdk.NewCoins(proposal.TotalDeposit...).String(),
				"voting_time":      common.TimeToNano(proposal.VotingStartTime),
				"voting_end_time":  common.TimeToNano(proposal.VotingEndTime),
				"content":          content,
			})
		case sdk.Msg:
			h.Write("NEW_PROPOSAL", common.JsDict{
				"id":               proposal.Id,
				"proposer":         nil,
				"type":             sdk.MsgTypeURL(subMsg),
				"title":            proposal.Title,
				"description":      proposal.Summary,
				"proposal_route":   sdk.MsgTypeURL(subMsg),
				"status":           int(proposal.Status),
				"submit_time":      common.TimeToNano(proposal.SubmitTime),
				"deposit_end_time": common.TimeToNano(proposal.DepositEndTime),
				"total_deposit":    sdk.NewCoins(proposal.TotalDeposit...).String(),
				"voting_time":      common.TimeToNano(proposal.VotingStartTime),
				"voting_end_time":  common.TimeToNano(proposal.VotingEndTime),
				"content":          subMsg,
			})
		default:
			break
		}
	}
	for _, deposit := range govState.Deposits {
		h.Write("SET_DEPOSIT", common.JsDict{
			"proposal_id": deposit.ProposalId,
			"depositor":   deposit.Depositor,
			"amount":      sdk.NewCoins(deposit.Amount...).String(),
			"tx_hash":     nil,
		})
	}
	for _, vote := range govState.Votes {
		setVoteWeighted := common.JsDict{
			"proposal_id": vote.ProposalId,
			"voter":       vote.Voter,
			"tx_hash":     nil,
		}
		h.emitSetVoteWeighted(setVoteWeighted, vote.Options)
	}

	// TSS module
	h.handleInitTssModule(ctx)

	// Bandtss module
	h.handleInitBandtssModule(ctx)

	// Oracle module
	var oracleState oracletypes.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[oracletypes.ModuleName], &oracleState)
	for idx, ds := range oracleState.DataSources {
		h.Write("NEW_DATA_SOURCE", common.JsDict{
			"id":          oracletypes.DataSourceID(idx + 1),
			"name":        ds.Name,
			"description": ds.Description,
			"owner":       ds.Owner,
			"executable":  h.oracleKeeper.GetFile(ds.Filename),
			"fee":         ds.Fee.String(),
			"treasury":    ds.Treasury,
			"tx_hash":     nil,
		})
	}
	for idx, os := range oracleState.OracleScripts {
		h.emitSetOracleScript(oracletypes.OracleScriptID(idx+1), os, nil)
	}

	// Feeds module
	var feedsState feedstypes.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[feedstypes.ModuleName], &feedsState)
	for _, vote := range feedsState.Votes {
		for _, signal := range vote.Signals {
			h.emitSetFeedsVoterSignal(ctx, vote.Voter, signal)
		}
	}

	signalTotalPowers := h.feedsKeeper.CalculateNewSignalTotalPowers(ctx)
	for _, stp := range signalTotalPowers {
		h.emitSetFeedsSignalTotalPower(stp)
	}

	// Restake module
	var restakeState restaketypes.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[restaketypes.ModuleName], &restakeState)
	for _, vault := range restakeState.Vaults {
		h.updateRestakeVault(ctx, vault.Key)
	}
	for _, lock := range restakeState.Locks {
		h.updateRestakeLock(ctx, lock.StakerAddress, lock.Key, nil)
	}
	for _, stake := range restakeState.Stakes {
		h.updateRestakeStake(ctx, stake.StakerAddress)
	}

	var authzState authz.GenesisState
	h.cdc.MustUnmarshalJSON(genesisState[authz.ModuleName], &authzState)
	for _, authz := range authzState.Authorization {
		authorization := authz.Authorization
		switch authorization.GetTypeUrl() {
		case sdk.MsgTypeURL(&oracletypes.MsgReportData{}):
			acc, _ := sdk.AccAddressFromBech32(authz.Granter)
			val := sdk.ValAddress(acc).String()
			h.Write("SET_REPORTER", common.JsDict{
				"reporter":  authz.Grantee,
				"validator": val,
			})
		}
	}

	h.Write("COMMIT", common.JsDict{"height": 0})
	h.FlushMessages()
}

// AfterBeginBlock specify actions need to do after begin block period (app.Hook interface).
func (h *Hook) AfterBeginBlock(ctx sdk.Context, req *abci.RequestFinalizeBlock, events []abci.Event) {
	h.accsInBlock = make(map[string]bool)
	h.accsInTx = make(map[string]bool)
	h.msgs = []common.Message{}
	if h.emitStartState {
		// TODO: fast-sync mode need to bring it back
		// h.emitStartState = false
		// h.emitNonHistoricalState(ctx)
	} else {
		for _, val := range req.DecidedLastCommit.Votes {
			validator, _ := h.stakingKeeper.ValidatorByConsAddr(ctx, val.Validator.Address)
			conAddr, _ := validator.GetConsAddr()
			h.Write("NEW_VALIDATOR_VOTE", common.JsDict{
				"consensus_address": sdk.ConsAddress(conAddr).String(),
				"block_height":      req.Height - 1,
				"voted":             val.BlockIdFlag == types1.BlockIDFlagCommit,
			})
			h.emitUpdateValidatorRewardAndAccumulatedCommission(ctx, MustParseValAddress(validator.GetOperator()))
		}
	}
	totalSupply := make([]string, 0)
	h.bankKeeper.IterateTotalSupply(ctx, func(coin sdk.Coin) bool {
		totalSupply = append(totalSupply, coin.String())
		return true
	})
	minter, _ := h.mintKeeper.Minter.Get(ctx)

	h.Write("NEW_BLOCK", common.JsDict{
		"height":    ctx.BlockHeight(),
		"timestamp": ctx.BlockTime().UnixNano(),
		// "proposer":  sdk.ConsAddress(req.Header.GetProposerAddress()).String(),
		"proposer":  sdk.ConsAddress(ctx.BlockHeader().ProposerAddress).String(),
		"hash":      ctx.HeaderHash(),
		"inflation": minter.Inflation.String(),
		"supply":    totalSupply,
	})

	eventQuerier := NewEventQuerier(events)
	for i, event := range events {
		h.handleBeginBlockEndBlockEvent(ctx, event, i, eventQuerier)
	}
}

// AfterDeliverTx specify actions need to do after transaction has been processed (app.Hook interface).
func (h *Hook) AfterDeliverTx(ctx sdk.Context, tx sdk.Tx, res *abci.ExecTxResult) {
	if ctx.BlockHeight() == 0 {
		return
	}
	h.accsInTx = make(map[string]bool)
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return
	}
	memoTx, ok := tx.(sdk.TxWithMemo)
	if !ok {
		return
	}

	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return
	}

	rawTx, _ := h.txConfig.TxEncoder()(tx)
	txHash := tmhash.Sum(rawTx)

	signers, _ := sigTx.GetSigners()

	txDict := common.JsDict{
		"hash":         txHash,
		"block_height": ctx.BlockHeight(),
		"gas_used":     res.GasUsed,
		"gas_limit":    feeTx.GetGas(),
		"gas_fee":      feeTx.GetFee().String(),
		"sender":       sdk.AccAddress(signers[0]).String(),
		"success":      res.IsOK(),
		"memo":         memoTx.GetMemo(),
		"fee_payer":    sdk.AccAddress(feeTx.FeeGranter()).String(),
	}
	if !res.IsOK() {
		txDict["err_msg"] = fmt.Sprintf("Error on codespace %s: with code %d", res.Codespace, res.Code)
	}

	// NOTE: We add txDict to the list of pending Kafka messages here, but it will still be
	// mutated in the loop below as we know the messages won't get flushed until ABCI Commit.
	h.Write("NEW_TRANSACTION", txDict)
	messages := []map[string]interface{}{}
	txMsgs := tx.GetMsgs()
	// Tx events not used yet
	_, events := splitTxEvents(len(txMsgs), res.Events)
	for i, msg := range txMsgs {
		detail := make(common.JsDict)
		DecodeMsg(msg, detail)
		if res.IsOK() {
			h.handleMsg(ctx, txHash, msg, events[i], detail)
		}
		messages = append(messages, common.JsDict{
			"msg":  detail,
			"type": sdk.MsgTypeURL(msg),
		})
	}
	addrs := make([]string, len(signers))
	for idx, signer := range signers {
		addrs[idx] = sdk.AccAddress(signer).String()
	}
	h.AddAccountsInTx(addrs...)
	relatedAccounts := make([]string, 0, len(h.accsInBlock))
	for acc := range h.accsInTx {
		relatedAccounts = append(relatedAccounts, acc)
	}

	h.AddAccountsInBlock(relatedAccounts...)
	h.Write("SET_RELATED_TRANSACTION", common.JsDict{
		"hash":             txHash,
		"related_accounts": relatedAccounts,
	})
	txDict["messages"] = messages
}

// AfterEndBlock specify actions need to do after end block period (app.Hook interface).
func (h *Hook) AfterEndBlock(ctx sdk.Context, events []abci.Event) {
	// update group proposals when voting period is end
	timeBytes := sdk.FormatTimeBytes(ctx.BlockTime().UTC())
	lenTimeByte := byte(len(timeBytes))
	prefix := []byte{groupkeeper.ProposalsByVotingPeriodEndPrefix}

	iterator := ctx.KVStore(h.groupStoreKey).
		Iterator(prefix, storetypes.PrefixEndBytes(append(append(prefix, lenTimeByte), timeBytes...)))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := splitKeyWithTime(iterator.Key())
		h.doUpdateGroupProposal(ctx, proposalID)
	}

	eventQuerier := NewEventQuerier(events)
	for i, event := range events {
		h.handleBeginBlockEndBlockEvent(ctx, event, i, eventQuerier)
	}

	// Emit all new current prices at every endblock.
	prices := h.feedsKeeper.GetAllCurrentPrices(ctx)
	if len(prices) > 0 {
		h.emitSetFeedsPrices(ctx, prices)
	}

	// Update balances of all affected accounts on this block.
	// Index 0 is message NEW_BLOCK, we insert SET_ACCOUNT messages right after it.
	modifiedMsgs := []common.Message{h.msgs[0]}
	for accStr := range h.accsInBlock {
		acc, _ := sdk.AccAddressFromBech32(accStr)
		modifiedMsgs = append(modifiedMsgs, common.Message{
			Key: "SET_ACCOUNT",
			Value: common.JsDict{
				"address": acc,
				"balance": h.bankKeeper.GetAllBalances(ctx, acc).String(),
			},
		})
	}

	h.msgs = append(modifiedMsgs, h.msgs[1:]...)
	h.Write("COMMIT", common.JsDict{"height": ctx.BlockHeight()})
}

func (h *Hook) RequestSearch(
	req *oracletypes.QueryRequestSearchRequest,
) (*oracletypes.QueryRequestSearchResponse, bool, error) {
	return nil, false, nil
}

func (h *Hook) RequestPrice(
	req *oracletypes.QueryRequestPriceRequest,
) (*oracletypes.QueryRequestPriceResponse, bool, error) {
	return nil, false, nil
}

// BeforeCommit specify actions need to do before commit block (app.Hook interface).
func (h *Hook) BeforeCommit() {
	h.FlushMessages()
}
