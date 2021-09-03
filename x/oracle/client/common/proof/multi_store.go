package proof

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MultiStoreProof stores a compact of other Cosmos-SDK modules' storage hash in multistore to
// compute (in combination with oracle store hash) Tendermint's application state hash at a given block.
//                                              ________________[AppHash]_________________
//                                             /                                          \
//                         _________________[I14]_________________                        [G]
//                        /                                        \
//             _______[I12]______                          _______[I13]________
//            /                  \                        /                    \
//       __[I8]__             __[I9]__                __[I10]__              __[I11]__
//      /         \          /         \            /          \            /         \
//    [I0]       [I1]     [I2]        [I3]        [I4]        [I5]        [I6]       [I7]
//   /   \      /   \    /    \      /    \      /    \      /    \      /    \     /    \
// [0]   [1]  [2]   [3] [4]   [5]  [6]    [7]  [8]    [9]  [A]    [B]  [C]    [D]  [E]   [F]
// [0] - auth     [1] - authz    [2] - bank    [3] - capability [4] - crisis  [5] - dist
// [6] - evidence [7] - feegrant [8] - gov     [9] - ibccore    [A] - mint    [B] - oracle
// [C] - params   [D] - slashing [E] - staking [F] - transfer   [G] - upgrade
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [B] In fact, only
// [A], [I4], [I11], [I12], and [G] are needed in order to compute [AppHash].
type MultiStoreProof struct {
	AuthToFeeGrantStoresMerkleHash   tmbytes.HexBytes `json:"auth_to_fee_grant_stores_Merkle_hash"`
	GovToIbcCoreStoresMerkleHash     tmbytes.HexBytes `json:"gov_to_ibc_core_stores_merkle_hash"`
	MintStoreMerkleHash              tmbytes.HexBytes `json:"mint_store_merkle_hash"`
	OracleIAVLStateHash              tmbytes.HexBytes `json:"oracle_iavl_State_hash"`
	ParamsToTransferStoresMerkleHash tmbytes.HexBytes `json:"params_to_transfer_stores_merkle_hash"`
	UpgradeStoreMerkleHash           tmbytes.HexBytes `json:"upgrade_store_merkle_hash"`
}

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	AuthToFeeGrantStoresMerkleHash   common.Hash
	GovToIbcCoreStoresMerkleHash     common.Hash
	MintStoreMerkleHash              common.Hash
	OracleIAVLStateHash              common.Hash
	ParamsToTransferStoresMerkleHash common.Hash
	UpgradeStoreMerkleHash           common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		AuthToFeeGrantStoresMerkleHash:   common.BytesToHash(m.AuthToFeeGrantStoresMerkleHash),
		GovToIbcCoreStoresMerkleHash:     common.BytesToHash(m.GovToIbcCoreStoresMerkleHash),
		MintStoreMerkleHash:              common.BytesToHash(m.MintStoreMerkleHash),
		OracleIAVLStateHash:              common.BytesToHash(m.OracleIAVLStateHash),
		ParamsToTransferStoresMerkleHash: common.BytesToHash(m.ParamsToTransferStoresMerkleHash),
		UpgradeStoreMerkleHash:           common.BytesToHash(m.UpgradeStoreMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		AuthToFeeGrantStoresMerkleHash:   tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		GovToIbcCoreStoresMerkleHash:     tmbytes.HexBytes(multiStoreEp.Path[1].Prefix[1:]),
		MintStoreMerkleHash:              tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		OracleIAVLStateHash:              tmbytes.HexBytes(multiStoreEp.Value),
		ParamsToTransferStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[2].Suffix),
		UpgradeStoreMerkleHash:           tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
