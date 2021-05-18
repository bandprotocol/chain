package proof

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmversion "github.com/tendermint/tendermint/proto/tendermint/version"
	"github.com/tendermint/tendermint/types"
)

/*
{
	"block_id": {
		"hash": "253E2EF603743B5CD0C7E8B8591082190398A16032ED2FF096F854033D106F4E",
		"parts": {
			"total": 1,
			"hash": "2EF12771EF64B0E04CA9C7C7DACAA483C09644D00786880620B25DBCED6A8637"
		}
	},
	"block": {
		"header": {
		"version": {
			"block": "11"
		},
		"chain_id": "bandchain",
		"height": "50000",
		"time": "2021-05-14T12:17:10.432169362Z",
		"last_block_id": {
			"hash": "7FC3C34D33E706A9743B155463F350A83D8C1DA3306D80C9E8D20151E590BD35",
			"parts": {
				"total": 1,
				"hash": "E5F569E83BEF3C1BFBADE8C0E400E5DF53EEDD2A51E8E26999F45EF600854AFB"
			}
		},
		"last_commit_hash": "7233A6E661B397CA76B52C6D789D071EF2479DFEEC3E65F362A0C87D2055EB73",
		"data_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
		"validators_hash": "4D9A1C01B7F48AA0B12B8F9BF558D4FB4D534EFD2C55E337F749A42C48D2D5B3",
		"next_validators_hash": "4D9A1C01B7F48AA0B12B8F9BF558D4FB4D534EFD2C55E337F749A42C48D2D5B3",
		"consensus_hash": "7A4E64D2A9B5CE22B75E956E2025274DC03B6846CCD7725152A6BC0D0727E33E",
		"app_hash": "87E354BB1C3E363A43CAD782F331C25C3228A131A30F56CBDEE9D899AE035741",
		"last_results_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
		"evidence_hash": "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
		"proposer_address": "F0C23921727D869745C4F9703CF33996B1D2B715"
	}
}
*/
func TestBlockHeaderMerkleParts(t *testing.T) {
	// Copy block header Merkle Part here
	header := types.Header{
		Version: tmversion.Consensus{Block: 11},
		ChainID: "bandchain",
		Height:  50000,
		Time:    parseTime("2021-05-14T12:17:10.432169362Z"),
		LastBlockID: types.BlockID{
			Hash: hexToBytes("7FC3C34D33E706A9743B155463F350A83D8C1DA3306D80C9E8D20151E590BD35"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("E5F569E83BEF3C1BFBADE8C0E400E5DF53EEDD2A51E8E26999F45EF600854AFB"),
			},
		},
		LastCommitHash:     hexToBytes("7233A6E661B397CA76B52C6D789D071EF2479DFEEC3E65F362A0C87D2055EB73"),
		DataHash:           hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		ValidatorsHash:     hexToBytes("4D9A1C01B7F48AA0B12B8F9BF558D4FB4D534EFD2C55E337F749A42C48D2D5B3"),
		NextValidatorsHash: hexToBytes("4D9A1C01B7F48AA0B12B8F9BF558D4FB4D534EFD2C55E337F749A42C48D2D5B3"),
		ConsensusHash:      hexToBytes("7A4E64D2A9B5CE22B75E956E2025274DC03B6846CCD7725152A6BC0D0727E33E"),
		AppHash:            hexToBytes("87E354BB1C3E363A43CAD782F331C25C3228A131A30F56CBDEE9D899AE035741"),
		LastResultsHash:    hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		EvidenceHash:       hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		ProposerAddress:    hexToBytes("F0C23921727D869745C4F9703CF33996B1D2B715"),
	}
	blockMerkleParts := GetBlockHeaderMerkleParts(&header)
	expectBlockHash := hexToBytes("253E2EF603743B5CD0C7E8B8591082190398A16032ED2FF096F854033D106F4E")
	appHash := tmbytes.HexBytes(hexToBytes("87E354BB1C3E363A43CAD782F331C25C3228A131A30F56CBDEE9D899AE035741"))

	// Verify code
	blockHash := innerHash(
		innerHash(
			innerHash(
				blockMerkleParts.VersionAndChainIDHash,
				innerHash(
					leafHash(cdcEncode(header.Height)),
					leafHash(encodeTime(header.Time)),
				),
			),
			blockMerkleParts.LastBlockIDAndOther,
		),
		innerHash(
			innerHash(
				blockMerkleParts.NextValidatorHashAndConsensusHash,
				innerHash(
					leafHash(cdcEncode(appHash)),
					blockMerkleParts.LastResultsHash,
				),
			),
			blockMerkleParts.EvidenceAndProposerHash,
		),
	)
	require.Equal(t, expectBlockHash, blockHash)
}
