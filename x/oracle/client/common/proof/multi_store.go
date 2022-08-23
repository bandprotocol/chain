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
//                         _________________[I14]_________________                     __[I15]__
//                        /                                        \				 /          \
//             _______[I12]______                          _______[I13]________    [G]         [H]
//            /                  \                        /                    \
//       __[I8]__             __[I9]__                __[I10]__              __[I11]__
//      /         \          /         \            /          \            /         \
//    [I0]       [I1]     [I2]        [I3]        [I4]        [I5]        [I6]       [I7]
//   /   \      /   \    /    \      /    \      /    \      /    \      /    \     /    \
// [0]   [1]  [2]   [3] [4]   [5]  [6]    [7]  [8]    [9]  [A]    [B]  [C]    [D]  [E]   [F]
// [0] - acc (auth) [1] - authz    [2] - bank     [3] - capability [4] - crisis   [5] - dist
// [6] - evidence   [7] - feegrant [8] - gov      [9] - ibccore    [A] - icahost  [B] - mint
// [C] - oracle     [D] - params   [E] - slashing [F] - staking    [G] - transfer [H] - upgrade
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [C] In fact, only
// [D], [I7], [I10], [I12], and [I15] are needed in order to compute [AppHash].
type MultiStoreProof struct {
	OracleIAVLStateHash               tmbytes.HexBytes `json:"oracle_iavl_state_hash"`
	ParamsStoreMerkleHash             tmbytes.HexBytes `json:"params_store_merkle_hash"`
	SlashingToStakingStoresMerkleHash tmbytes.HexBytes `json:"slashing_to_staking_stores_merkle_hash"`
	GovToMintStoresMerkleHash         tmbytes.HexBytes `json:"gov_to_mint_stores_merkle_hash"`
	AuthToFeegrantStoresMerkleHash    tmbytes.HexBytes `json:"auth_to_fee_grant_stores_merkle_hash"`
	TransferToUpgradeStoresMerkleHash tmbytes.HexBytes `json:"transfer_to_upgrade_stores_merkle_hash"`
}

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash               common.Hash
	ParamsStoreMerkleHash             common.Hash
	SlashingToStakingStoresMerkleHash common.Hash
	GovToMintStoresMerkleHash         common.Hash
	AuthToFeegrantStoresMerkleHash    common.Hash
	TransferToUpgradeStoresMerkleHash common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:               common.BytesToHash(m.OracleIAVLStateHash),
		ParamsStoreMerkleHash:             common.BytesToHash(m.ParamsStoreMerkleHash),
		SlashingToStakingStoresMerkleHash: common.BytesToHash(m.SlashingToStakingStoresMerkleHash),
		GovToMintStoresMerkleHash:         common.BytesToHash(m.GovToMintStoresMerkleHash),
		AuthToFeegrantStoresMerkleHash:    common.BytesToHash(m.AuthToFeegrantStoresMerkleHash),
		TransferToUpgradeStoresMerkleHash: common.BytesToHash(m.TransferToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) MultiStoreProof {
	return MultiStoreProof{
		OracleIAVLStateHash:               tmbytes.HexBytes(multiStoreEp.Value),
		ParamsStoreMerkleHash:             tmbytes.HexBytes(multiStoreEp.Path[0].Suffix),
		SlashingToStakingStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		GovToMintStoresMerkleHash:         tmbytes.HexBytes(multiStoreEp.Path[2].Prefix[1:]),
		AuthToFeegrantStoresMerkleHash:    tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		TransferToUpgradeStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
