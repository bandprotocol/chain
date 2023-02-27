package proof

import (
	ics23 "github.com/confio/ics23/go"
	"github.com/ethereum/go-ethereum/common"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// MultiStoreProof stores a compact of other Cosmos-SDK modules' storage hash in multistore to
// compute (in combination with oracle store hash) Tendermint's application state hash at a given block.
//
//	                                           ________________[AppHash]_________________________
//	                                          /                                                  \
//	                      _________________[N15]_________________                             __[N16]___
//	                     /                                        \                          /          \
//	          _______[N12]______                          _______[N13]________             [N14]        [I]
//	         /                  \                        /                    \           /     \
//	    __[N8]__             __[N9]__                __[N10]__              __[N11]__    [G]   [H]
//	   /         \          /         \            /          \            /         \
//	  [N0]       [N1]      [N2]       [N3]       [N4]        [N5]         [N6]       [N7]
//	/     \     /    \    /    \     /     \     /     \     /     \     /     \     /    \
//
// [0]   [1]  [2]   [3] [4]   [5]  [6]    [7]  [8]    [9]  [A]    [B]  [C]    [D]  [E]   [F]
//
// [0] - acc (auth) [1] - authz    [2] - bank     [3] - capability [4] - crisis   [5] - dist
// [6] - evidence   [7] - feegrant [8] - gov      [9] - group      [A] - ibccore    [B] - icahost
// [C] - mint       [D] - oracle   [E] - params   [F] - slashing   [G] - staking    [H] - transfer [I] - upgrade
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [D] In fact, only
// [C], [N7], [N10], [N12], and [N16] are needed in order to compute [AppHash].

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIavlStateHash              common.Hash
	MintStoreMerkleHash              common.Hash
	ParamsToSlashingStoresMerkleHash common.Hash
	GovToIcahostStoresMerkleHash     common.Hash
	AuthToFeeGrantStoresMerkleHash   common.Hash
	StakingToUpgradeStoresMerkleHash common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIavlStateHash:              common.BytesToHash(m.OracleIavlStateHash),
		MintStoreMerkleHash:              common.BytesToHash(m.MintStoreMerkleHash),
		ParamsToSlashingStoresMerkleHash: common.BytesToHash(m.ParamsToSlashingStoresMerkleHash),
		GovToIcahostStoresMerkleHash:     common.BytesToHash(m.GovToIcahostStoresMerkleHash),
		AuthToFeeGrantStoresMerkleHash:   common.BytesToHash(m.AuthToFeeGrantStoresMerkleHash),
		StakingToUpgradeStoresMerkleHash: common.BytesToHash(m.StakingToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) *MultiStoreProof {
	return &MultiStoreProof{
		OracleIavlStateHash:              tmbytes.HexBytes(multiStoreEp.Value),
		MintStoreMerkleHash:              tmbytes.HexBytes(multiStoreEp.Path[0].Prefix[1:]),
		ParamsToSlashingStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[1].Suffix),
		GovToIcahostStoresMerkleHash:     tmbytes.HexBytes(multiStoreEp.Path[2].Prefix[1:]),
		AuthToFeeGrantStoresMerkleHash:   tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		StakingToUpgradeStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
