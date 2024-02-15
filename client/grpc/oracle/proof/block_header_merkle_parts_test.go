package proof

import (
	"testing"

	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	tmversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

/*
	{
		block_id: {
			hash: "3489F21785ACE1CE4214CB2B57F3A98DC0B7377D1BA1E1180B6E199E33B0FC5A",
			parts: {
				total: 1,
				hash: "6BF91EFBA26A4CD86EBBD0E54DCFC9BD2C790859CFA96215661A47E4921A6301"
			}
		},
		block: {
			header: {
				version: {
					block: "11"
				},
				chain_id: "bandchain",
				height: "25000",
				time: "2021-08-25T00:05:31.290650376Z",
				last_block_id: {
					hash: "622A4600128DECC6C42E471F06F00C654785485D5AB4437556F41743DC4684C8",
					parts: {
						total: 1,
						hash: "733EDAE763A4635509BE9E55E06A2CBF726056A0898B6B4D3AF74683ECCF3475"
					}
				},
				last_commit_hash: "021C8BBD047747AE943C5F7991B6848DE371B313FF3C15E5B2EDA94DD834BB42",
				data_hash: "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
				validators_hash: "113928F64B9B2A1F1A58D87C93147D822CF2069309B55D47717700D4074A43B6",
				next_validators_hash: "113928F64B9B2A1F1A58D87C93147D822CF2069309B55D47717700D4074A43B6",
				consensus_hash: "188E4357E7B1201E6C2B418759CB8246FAB30CF2FFA87433E21690B7BC8BC88C",
				app_hash: "37D2CA95F226A7AFE3C41DE288F8158B737E78C4B733B1CCB0061D3236E926BE",
				last_results_hash: "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
				evidence_hash: "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
				proposer_address: "F23391B5DBF982E37FB7DADEA64AAE21CAE4C172"
			}
		}
	}
*/
func TestBlockHeaderMerkleParts(t *testing.T) {
	// Copy block header Merkle Part here
	header := types.Header{
		Version: tmversion.Consensus{Block: 11},
		ChainID: "bandchain",
		Height:  25000,
		Time:    parseTime("2021-08-25T00:05:31.290650376Z"),
		LastBlockID: types.BlockID{
			Hash: hexToBytes("622A4600128DECC6C42E471F06F00C654785485D5AB4437556F41743DC4684C8"),
			PartSetHeader: types.PartSetHeader{
				Total: 1,
				Hash:  hexToBytes("733EDAE763A4635509BE9E55E06A2CBF726056A0898B6B4D3AF74683ECCF3475"),
			},
		},
		LastCommitHash:     hexToBytes("021C8BBD047747AE943C5F7991B6848DE371B313FF3C15E5B2EDA94DD834BB42"),
		DataHash:           hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		ValidatorsHash:     hexToBytes("113928F64B9B2A1F1A58D87C93147D822CF2069309B55D47717700D4074A43B6"),
		NextValidatorsHash: hexToBytes("113928F64B9B2A1F1A58D87C93147D822CF2069309B55D47717700D4074A43B6"),
		ConsensusHash:      hexToBytes("188E4357E7B1201E6C2B418759CB8246FAB30CF2FFA87433E21690B7BC8BC88C"),
		AppHash:            hexToBytes("37D2CA95F226A7AFE3C41DE288F8158B737E78C4B733B1CCB0061D3236E926BE"),
		LastResultsHash:    hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		EvidenceHash:       hexToBytes("E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"),
		ProposerAddress:    hexToBytes("F23391B5DBF982E37FB7DADEA64AAE21CAE4C172"),
	}
	blockMerkleParts := GetBlockHeaderMerkleParts(&header)
	expectBlockHash := hexToBytes("3489F21785ACE1CE4214CB2B57F3A98DC0B7377D1BA1E1180B6E199E33B0FC5A")
	appHash := tmbytes.HexBytes(hexToBytes("37D2CA95F226A7AFE3C41DE288F8158B737E78C4B733B1CCB0061D3236E926BE"))

	// Verify code
	blockHash := innerHash(
		innerHash(
			innerHash(
				blockMerkleParts.VersionAndChainIdHash,
				innerHash(
					leafHash(cdcEncode(header.Height)),
					leafHash(encodeTime(header.Time)),
				),
			),
			blockMerkleParts.LastBlockIdAndOther,
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
