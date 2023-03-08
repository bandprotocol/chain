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
//	                      _________________[N16]_________________                             __[N17]___
//	                     /                                        \                          /          \
//	          _______[N12]______                          _______[N13]________             [N14]        [N15]
//	         /                  \                        /                    \           /     \     /      \
//	    __[N8]__             __[N9]__                __[N10]__              __[N11]__    [G]   [H]   [I]    [J]
//	   /         \          /         \            /          \            /         \
//	  [N0]       [N1]      [N2]       [N3]       [N4]        [N5]         [N6]       [N7]
//	/     \     /    \    /    \     /     \     /     \     /     \     /     \     /    \
// [0]   [1]  [2]   [3] [4]   [5]  [6]    [7]  [8]    [9]  [A]    [B]  [C]    [D]  [E]   [F]
//
// [0] - acc (auth) [1] - authz    [2] - bank      [3] - capability [4] - crisis    [5] - dist
// [6] - evidence   [7] - feegrant [8] - globalfee [9] - gov        [A] - group     [B] - ibccore
// [C] - icahost    [D] - mint     [E] - oracle    [F] - params     [G] - slashing  [H] - staking
// [I] - transfer   [J] - upgrade
// Notice that NOT all leaves of the Merkle tree are needed in order to compute the Merkle
// root hash, since we only want to validate the correctness of [D] In fact, only
// [F], [N6], [N10], [N12], and [N17] are needed in order to compute [AppHash].

// MultiStoreProofEthereum is an Ethereum version of MultiStoreProof for solidity ABI-encoding.
type MultiStoreProofEthereum struct {
	OracleIAVLStateHash                common.Hash
	ParamsStoreMerkleHash              common.Hash
	IcahostToMintStoresMerkleHash      common.Hash
	GlobalfeeToIbccoreStoresMerkleHash common.Hash
	AuthToFeegrantStoresMerkleHash     common.Hash
	SlashingToUpgradeStoresMerkleHash  common.Hash
}

func (m *MultiStoreProof) encodeToEthFormat() MultiStoreProofEthereum {
	return MultiStoreProofEthereum{
		OracleIAVLStateHash:                common.BytesToHash(m.OracleIAVLStateHash),
		ParamsStoreMerkleHash:              common.BytesToHash(m.ParamsStoreMerkleHash),
		IcahostToMintStoresMerkleHash:      common.BytesToHash(m.IcahostToMintStoresMerkleHash),
		GlobalfeeToIbccoreStoresMerkleHash: common.BytesToHash(m.GlobalfeeToIbccoreStoresMerkleHash),
		AuthToFeegrantStoresMerkleHash:     common.BytesToHash(m.AuthToFeegrantStoresMerkleHash),
		SlashingToUpgradeStoresMerkleHash:  common.BytesToHash(m.SlashingToUpgradeStoresMerkleHash),
	}
}

// GetMultiStoreProof compacts Multi store proof from Tendermint to MultiStoreProof version.
func GetMultiStoreProof(multiStoreEp *ics23.ExistenceProof) *MultiStoreProof {
	return &MultiStoreProof{
		OracleIAVLStateHash:                tmbytes.HexBytes(multiStoreEp.Value),
		ParamsStoreMerkleHash:              tmbytes.HexBytes(multiStoreEp.Path[0].Suffix),
		IcahostToMintStoresMerkleHash:      tmbytes.HexBytes(multiStoreEp.Path[1].Prefix[1:]),
		GlobalfeeToIbccoreStoresMerkleHash: tmbytes.HexBytes(multiStoreEp.Path[2].Prefix[1:]),
		AuthToFeegrantStoresMerkleHash:     tmbytes.HexBytes(multiStoreEp.Path[3].Prefix[1:]),
		SlashingToUpgradeStoresMerkleHash:  tmbytes.HexBytes(multiStoreEp.Path[4].Suffix),
	}
}
