package ibctesting

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/testing/mock"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmprotoversion "github.com/tendermint/tendermint/proto/tendermint/version"
	tmtypes "github.com/tendermint/tendermint/types"
	tmversion "github.com/tendermint/tendermint/version"

	bandapp "github.com/bandprotocol/chain/v2/app"
	"github.com/bandprotocol/chain/v2/testing/testapp"
	"github.com/bandprotocol/chain/v2/x/oracle/types"
)

var (
	valSize uint64 = 2
)

// TestChain is a testing struct that wraps a simapp with the last TM Header, the current ABCI
// header and the validators of the TestChain. It also contains a field called ChainID. This
// is the clientID that *other* chains use to refer to this TestChain. The SenderAccount
// is used for delivering transactions through the application state.
// NOTE: the actual application uses an empty chain-id for ease of testing.
type TestChain struct {
	t *testing.T

	Coordinator   *Coordinator
	App           *testapp.TestingApp
	ChainID       string
	LastHeader    *ibctmtypes.Header // header for last block height committed
	CurrentHeader tmproto.Header     // header for current block height
	// QueryServer   types.QueryServer
	TxConfig client.TxConfig
	Codec    codec.BinaryMarshaler

	Vals     *tmtypes.ValidatorSet
	Signers  []tmtypes.PrivValidator
	Treasury sdk.AccAddress

	senderPrivKey cryptotypes.PrivKey
	SenderAccount authtypes.AccountI

	senders map[string]*authtypes.BaseAccount
}

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on
// counterparty chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the Coordinator in order to ensure synchrony between chains.
// Each update of any chain increments the block header time for all chains by 5 seconds.
func NewTestChain(t *testing.T, coord *Coordinator, chainID string) *TestChain {
	// TODO: change sender
	signers := make([]tmtypes.PrivValidator, valSize)
	validators := make([]*tmtypes.Validator, valSize)
	genesisAccount := make([]authtypes.GenesisAccount, valSize)
	balances := make([]banktypes.Balance, valSize)
	senders := make(map[string]*authtypes.BaseAccount)

	for i := uint64(0); i < valSize; i++ {
		// generate validator private/public key
		privVal := mock.PV{testapp.Validators[i].PrivKey}
		tmPub, err := cryptocodec.ToTmPubKeyInterface(testapp.Validators[i].PubKey)
		require.NoError(t, err)

		// create validator set with two validators
		validators[i] = tmtypes.NewValidator(tmPub, 1)

		signers[i] = privVal

		senders[testapp.Validators[i].Address.String()] = authtypes.NewBaseAccount(testapp.Validators[i].PubKey.Address().Bytes(), testapp.Validators[i].PubKey, i, 0)
		genesisAccount[i] = senders[testapp.Validators[i].Address.String()]
		balances[i] = banktypes.Balance{
			Address: genesisAccount[i].GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin("uband", sdk.NewInt(100000000000000))),
		}
	}

	valSet := tmtypes.NewValidatorSet(validators)

	app := testapp.SetupWithGenesisValSet(t, valSet, genesisAccount, balances...)
	ctx := app.NewContext(false, tmproto.Header{Height: app.LastBlockHeight()})
	vals := app.StakingKeeper.GetAllValidators(ctx)
	for _, v := range vals {
		app.OracleKeeper.Activate(ctx, v.GetOperator())
	}

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := bandapp.MakeEncodingConfig().TxConfig

	// create an account to send transactions from
	chain := &TestChain{
		t:             t,
		Coordinator:   coord,
		ChainID:       chainID,
		App:           app,
		CurrentHeader: header,
		// QueryServer:   app.GetIBCKeeper(),
		TxConfig:      txConfig,
		Codec:         app.AppCodec(),
		Vals:          valSet,
		Signers:       signers,
		senderPrivKey: testapp.Validators[0].PrivKey,
		SenderAccount: genesisAccount[0],
		Treasury:      testapp.Treasury.Address,
		senders:       senders,
	}

	coord.CommitBlock(chain)

	return chain
}

// GetContext returns the current context for the application.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.GetBaseApp().NewContext(false, chain.CurrentHeader)
}

// // GetSimApp returns the SimApp to allow usage ofnon-interface fields.
// // CONTRACT: This function should not be called by third parties implementing
// // their own SimApp.
// func (chain *TestChain) GetSimApp() *simapp.SimApp {
// 	app, ok := chain.App.(*simapp.SimApp)
// 	require.True(chain.t, ok)

// 	return app
// }

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProof(key []byte) ([]byte, clienttypes.Height) {
	return chain.QueryProofAtHeight(key, chain.App.LastBlockHeight())
}

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProofAtHeight(key []byte, height int64) ([]byte, clienttypes.Height) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", host.StoreKey),
		Height: height - 1,
		Data:   key,
		Prove:  true,
	})

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.t, err)

	proof, err := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)
	require.NoError(chain.t, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height)+1)
}

// QueryUpgradeProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryUpgradeProof(key []byte, height uint64) ([]byte, clienttypes.Height) {
	res := chain.App.Query(abci.RequestQuery{
		Path:   "store/upgrade/key",
		Height: int64(height - 1),
		Data:   key,
		Prove:  true,
	})

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.t, err)

	proof, err := chain.App.AppCodec().MarshalBinaryBare(&merkleProof)
	require.NoError(chain.t, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height+1))
}

// QueryConsensusStateProof performs an abci query for a consensus state
// stored on the given clientID. The proof and consensusHeight are returned.
func (chain *TestChain) QueryConsensusStateProof(clientID string) ([]byte, clienttypes.Height) {
	clientState := chain.GetClientState(clientID)

	consensusHeight := clientState.GetLatestHeight().(clienttypes.Height)
	consensusKey := host.FullConsensusStateKey(clientID, consensusHeight)
	proofConsensus, _ := chain.QueryProof(consensusKey)

	return proofConsensus, consensusHeight
}

// NextBlock sets the last header to the current header and increments the current header to be
// at the next block height. It does not update the time as that is handled by the Coordinator.
//
// CONTRACT: this function must only be called after app.Commit() occurs
func (chain *TestChain) NextBlock() {
	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentTMClientHeader()

	// increment the current header
	chain.CurrentHeader = tmproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.Vals.Hash(),
	}
	chain.App.BeginBlock(abci.RequestBeginBlock{Header: chain.CurrentHeader, Hash: chain.CurrentHeader.AppHash})
}

// sendMsgs delivers a transaction through the application without returning the result.
func (chain *TestChain) sendMsgs(msgs ...sdk.Msg) error {
	_, err := chain.SendMsgs(msgs...)
	return err
}

// SendMsgs delivers a transaction through the application. It updates the senders sequence
// number and updates the TestChain's headers. It returns the result and error if one
// occurred.
func (chain *TestChain) SendMsgs(msgs ...sdk.Msg) (*sdk.Result, error) {

	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	_, r, err := testapp.SignAndDeliver(
		chain.t,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		msgs,
		chain.ChainID,
		[]uint64{chain.SenderAccount.GetAccountNumber()},
		[]uint64{chain.SenderAccount.GetSequence()},
		chain.senderPrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)

	chain.Coordinator.IncrementTime()

	return r, nil
}

// SendMsgs delivers a transaction through the application. It updates the senders sequence
// number and updates the TestChain's headers. It returns the result and error if one
// occurred.
func (chain *TestChain) SendReport(rid types.RequestID, rawReps []types.RawReport, sender testapp.Account) (*sdk.Result, error) {
	senderAccount := chain.senders[sender.Address.String()]

	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	_, r, err := testapp.SignAndDeliver(
		chain.t,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		chain.GetContext().BlockHeader(),
		[]sdk.Msg{types.NewMsgReportData(rid, rawReps, sender.ValAddress, sender.Address)},
		chain.ChainID,
		[]uint64{senderAccount.GetAccountNumber()},
		[]uint64{senderAccount.GetSequence()},
		sender.PrivKey,
	)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	senderAccount.SetSequence(senderAccount.GetSequence() + 1)

	chain.Coordinator.IncrementTime()

	return r, nil
}

// GetClientState retrieves the client state for the provided clientID. The client is
// expected to exist otherwise testing will fail.
func (chain *TestChain) GetClientState(clientID string) exported.ClientState {
	clientState, found := chain.App.GetIBCKeeper().ClientKeeper.GetClientState(chain.GetContext(), clientID)
	require.True(chain.t, found)

	return clientState
}

// GetConsensusState retrieves the consensus state for the provided clientID and height.
// It will return a success boolean depending on if consensus state exists or not.
func (chain *TestChain) GetConsensusState(clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return chain.App.GetIBCKeeper().ClientKeeper.GetClientConsensusState(chain.GetContext(), clientID, height)
}

// GetValsAtHeight will return the validator set of the chain at a given height. It will return
// a success boolean depending on if the validator set exists or not at that height.
func (chain *TestChain) GetValsAtHeight(height int64) (*tmtypes.ValidatorSet, bool) {
	histInfo, ok := chain.App.GetStakingKeeper().GetHistoricalInfo(chain.GetContext(), height)
	if !ok {
		return nil, false
	}

	valSet := stakingtypes.Validators(histInfo.Valset)

	tmValidators, err := teststaking.ToTmValidators(valSet)
	if err != nil {
		panic(err)
	}
	return tmtypes.NewValidatorSet(tmValidators), true
}

// GetAcknowledgement retrieves an acknowledgement for the provided packet. If the
// acknowledgement does not exist then testing will fail.
func (chain *TestChain) GetAcknowledgement(packet exported.PacketI) []byte {
	ack, found := chain.App.GetIBCKeeper().ChannelKeeper.GetPacketAcknowledgement(chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	require.True(chain.t, found)

	return ack
}

// GetPrefix returns the prefix for used by a chain in connection creation
func (chain *TestChain) GetPrefix() commitmenttypes.MerklePrefix {
	return commitmenttypes.NewMerklePrefix(chain.App.GetIBCKeeper().ConnectionKeeper.GetCommitmentPrefix().Bytes())
}

// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
// light client on the source chain.
func (chain *TestChain) ConstructUpdateTMClientHeader(counterparty *TestChain, clientID string) (*ibctmtypes.Header, error) {
	return chain.ConstructUpdateTMClientHeaderWithTrustedHeight(counterparty, clientID, clienttypes.ZeroHeight())
}

// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
// light client on the source chain.
func (chain *TestChain) ConstructUpdateTMClientHeaderWithTrustedHeight(counterparty *TestChain, clientID string, trustedHeight clienttypes.Height) (*ibctmtypes.Header, error) {
	header := counterparty.LastHeader
	// Relayer must query for LatestHeight on client to get TrustedHeight if the trusted height is not set
	if trustedHeight.IsZero() {
		trustedHeight = chain.GetClientState(clientID).GetLatestHeight().(clienttypes.Height)
	}
	var (
		tmTrustedVals *tmtypes.ValidatorSet
		ok            bool
	)
	// Once we get TrustedHeight from client, we must query the validators from the counterparty chain
	// If the LatestHeight == LastHeader.Height, then TrustedValidators are current validators
	// If LatestHeight < LastHeader.Height, we can query the historical validator set from HistoricalInfo
	if trustedHeight == counterparty.LastHeader.GetHeight() {
		tmTrustedVals = counterparty.Vals
	} else {
		// NOTE: We need to get validators from counterparty at height: trustedHeight+1
		// since the last trusted validators for a header at height h
		// is the NextValidators at h+1 committed to in header h by
		// NextValidatorsHash
		tmTrustedVals, ok = counterparty.GetValsAtHeight(int64(trustedHeight.RevisionHeight + 1))
		if !ok {
			return nil, sdkerrors.Wrapf(ibctmtypes.ErrInvalidHeaderHeight, "could not retrieve trusted validators at trustedHeight: %d", trustedHeight)
		}
	}
	// inject trusted fields into last header
	// for now assume revision number is 0
	header.TrustedHeight = trustedHeight

	trustedVals, err := tmTrustedVals.ToProto()
	if err != nil {
		return nil, err
	}
	header.TrustedValidators = trustedVals

	return header, nil

}

// ExpireClient fast forwards the chain's block time by the provided amount of time which will
// expire any clients with a trusting period less than or equal to this amount of time.
func (chain *TestChain) ExpireClient(amount time.Duration) {
	chain.Coordinator.IncrementTimeBy(amount)
}

// CurrentTMClientHeader creates a TM header using the current header parameters
// on the chain. The trusted fields in the header are set to nil.
func (chain *TestChain) CurrentTMClientHeader() *ibctmtypes.Header {
	return chain.CreateTMClientHeader(chain.ChainID, chain.CurrentHeader.Height, clienttypes.Height{}, chain.CurrentHeader.Time, chain.Vals, nil, chain.Signers)
}

// CreateTMClientHeader creates a TM header to update the TM client. Args are passed in to allow
// caller flexibility to use params that differ from the chain.
func (chain *TestChain) CreateTMClientHeader(chainID string, blockHeight int64, trustedHeight clienttypes.Height, timestamp time.Time, tmValSet, tmTrustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) *ibctmtypes.Header {
	var (
		valSet      *tmproto.ValidatorSet
		trustedVals *tmproto.ValidatorSet
	)
	require.NotNil(chain.t, tmValSet)

	vsetHash := tmValSet.Hash()

	tmHeader := tmtypes.Header{
		Version:            tmprotoversion.Consensus{Block: tmversion.BlockProtocol, App: 2},
		ChainID:            chainID,
		Height:             blockHeight,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), 10_000, make([]byte, tmhash.Size)),
		LastCommitHash:     chain.App.LastCommitID().Hash,
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: vsetHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            chain.CurrentHeader.AppHash,
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    tmValSet.Proposer.Address, //nolint:staticcheck
	}
	hhash := tmHeader.Hash()
	blockID := MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := tmtypes.NewVoteSet(chainID, blockHeight, 1, tmproto.PrecommitType, tmValSet)

	commit, err := tmtypes.MakeCommit(blockID, blockHeight, 1, voteSet, signers, timestamp)
	require.NoError(chain.t, err)

	signedHeader := &tmproto.SignedHeader{
		Header: tmHeader.ToProto(),
		Commit: commit.ToProto(),
	}

	if tmValSet != nil {
		valSet, err = tmValSet.ToProto()
		if err != nil {
			panic(err)
		}
	}

	if tmTrustedVals != nil {
		trustedVals, err = tmTrustedVals.ToProto()
		if err != nil {
			panic(err)
		}
	}

	// The trusted fields may be nil. They may be filled before relaying messages to a client.
	// The relayer is responsible for querying client and injecting appropriate trusted fields.
	return &ibctmtypes.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      valSet,
		TrustedHeight:     trustedHeight,
		TrustedValidators: trustedVals,
	}
}

// MakeBlockID copied unimported test functions from tmtypes to use them here
func MakeBlockID(hash []byte, partSetSize uint32, partSetHash []byte) tmtypes.BlockID {
	return tmtypes.BlockID{
		Hash: hash,
		PartSetHeader: tmtypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// CreateSortedSignerArray takes two PrivValidators, and the corresponding Validator structs
// (including voting power). It returns a signer array of PrivValidators that matches the
// sorting of ValidatorSet.
// The sorting is first by .VotingPower (descending), with secondary index of .Address (ascending).
func CreateSortedSignerArray(altPrivVal, suitePrivVal tmtypes.PrivValidator,
	altVal, suiteVal *tmtypes.Validator) []tmtypes.PrivValidator {

	switch {
	case altVal.VotingPower > suiteVal.VotingPower:
		return []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
	case altVal.VotingPower < suiteVal.VotingPower:
		return []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	default:
		if bytes.Compare(altVal.Address, suiteVal.Address) == -1 {
			return []tmtypes.PrivValidator{altPrivVal, suitePrivVal}
		}
		return []tmtypes.PrivValidator{suitePrivVal, altPrivVal}
	}
}

// CreatePortCapability binds and claims a capability for the given portID if it does not
// already exist. This function will fail testing on any resulting error.
// NOTE: only creation of a capbility for a transfer or mock port is supported
// Other applications must bind to the port in InitGenesis or modify this code.
func (chain *TestChain) CreatePortCapability(scopedKeeper capabilitykeeper.ScopedKeeper, portID string) {
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), host.PortPath(portID))
	if !ok {
		// create capability using the IBC capability keeper
		cap, err := chain.App.GetScopedIBCKeeper().NewCapability(chain.GetContext(), host.PortPath(portID))
		require.NoError(chain.t, err)

		// claim capability using the scopedKeeper
		err = scopedKeeper.ClaimCapability(chain.GetContext(), cap, host.PortPath(portID))
		require.NoError(chain.t, err)
	}

	chain.App.Commit()

	chain.NextBlock()
}

// GetPortCapability returns the port capability for the given portID. The capability must
// exist, otherwise testing will fail.
func (chain *TestChain) GetPortCapability(portID string) *capabilitytypes.Capability {
	cap, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), host.PortPath(portID))
	require.True(chain.t, ok)

	return cap
}

// CreateChannelCapability binds and claims a capability for the given portID and channelID
// if it does not already exist. This function will fail testing on any resulting error. The
// scoped keeper passed in will claim the new capability.
func (chain *TestChain) CreateChannelCapability(scopedKeeper capabilitykeeper.ScopedKeeper, portID, channelID string) {
	capName := host.ChannelCapabilityPath(portID, channelID)
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), capName)
	if !ok {
		cap, err := chain.App.GetScopedIBCKeeper().NewCapability(chain.GetContext(), capName)
		require.NoError(chain.t, err)
		err = scopedKeeper.ClaimCapability(chain.GetContext(), cap, capName)
		require.NoError(chain.t, err)
	}

	chain.App.Commit()

	chain.NextBlock()
}

// GetChannelCapability returns the channel capability for the given portID and channelID.
// The capability must exist, otherwise testing will fail.
func (chain *TestChain) GetChannelCapability(portID, channelID string) *capabilitytypes.Capability {
	cap, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), host.ChannelCapabilityPath(portID, channelID))
	require.True(chain.t, ok)

	return cap
}
