package proof

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MultiStoreProof stores a compact of other Cosmos-SDK modules' storage hash in multistore to
// compute (in combination with oracle store hash) Tendermint's application state hash at a given block.
//                         ________________[AppHash]_______________
//                        /                                        \
//             _______[I10]______                          _______[I11]________
//            /                  \                        /                    \
//       __[I6]__             __[I7]__                __[I8]__              __[I9]__
//      /         \          /         \            /          \           /         \
//    [I0]       [I1]     [I2]        [I3]        [I4]        [I5]       [C]         [D]
//   /   \      /   \    /    \      /    \      /    \      /    \
// [0]   [1]  [2]   [3] [4]   [5]  [6]    [7]  [8]    [9]  [A]    [B]
// [0] - auth   [1] - bank     [2] - capability  [3] - dist    [4] - evidence
// [5] - gov    [6] - ibchost  [7] - ibctransfer [8] - mint    [9] - oracle
// [A] - params [B] - slashing [C] - staking     [D] - upgrade
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [9] In fact, only
// [8], [I5], [I9], and [I10] are needed in order to compute [AppHash].
type MultiStoreProof struct {
	AuthToIbcTransferStoresMerkleHash tmbytes.HexBytes `json:"auth_to_ibc_transfer_stores_Merkle_hash"`
	MintStoreMerkleHash               tmbytes.HexBytes `json:"mint_store_merkle_hash"`
	OracleIAVLStateHash               tmbytes.HexBytes `json:"oracle_iavl_State_hash"`
	ParamsToSlashStoresMerkleHash     tmbytes.HexBytes `json:"params_to_slash_stores_merkle_hash"`
	StakingToUpgradeStoresMerkleHash  tmbytes.HexBytes `json:"staking_to_upgrade_stores_merkle_hash"`
}

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	AuthToIbcTransferStoresMerkleHash common.Hash
	MintStoreMerkleHash               common.Hash
	OracleIAVLStateHash               common.Hash
	ParamsToSlashStoresMerkleHash     common.Hash
	StakingToUpgradeStoresMerkleHash  common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		AuthToIbcTransferStoresMerkleHash: common.BytesToHash(m.AuthToIbcTransferStoresMerkleHash),
		MintStoreMerkleHash:               common.BytesToHash(m.MintStoreMerkleHash),
		OracleIAVLStateHash:               common.BytesToHash(m.OracleIAVLStateHash),
		ParamsToSlashStoresMerkleHash:     common.BytesToHash(m.ParamsToSlashStoresMerkleHash),
		StakingToUpgradeStoresMerkleHash:  common.BytesToHash(m.StakingToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		AuthToIbcTransferStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		MintStoreMerkleHash:               tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		OracleIAVLStateHash:               tmbytes.HexBytes(multiStoreEp.Value),
		ParamsToSlashStoresMerkleHash:     tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		StakingToUpgradeStoresMerkleHash:  tmbytes.HexBytes(multiStoreEp.Path[2].Suffix),
	}
}
