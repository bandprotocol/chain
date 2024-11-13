package proof

// TODO: fix test for merkle proof
// import (
// 	"testing"

// 	"github.com/stretchr/testify/require"

// 	"github.com/cometbft/cometbft/crypto/tmhash"

// 	ics23 "github.com/cosmos/ics23/go"

// 	storetypes "cosmossdk.io/store/types"
// )

// /*
// query at localhost:26657/abci_query?path="/store/oracle/key"&data=0xc000000000000000&prove=true
// {
//   "jsonrpc": "2.0",
//   "id": -1,
//   "result": {
//     "response": {
//       "code": 0,
//       "log": "",
//       "info": "",
//       "index": "0",
//       "key": "wAAAAAAAAAA=",
//       "value": null,
//       "proofOps": {
//         "ops": [
//           {
//             "type": "ics23:iavl",
//             "key": "wAAAAAAAAAA=",
//             "data": "ErUDCgjAAAAAAAAAABLeAQoBBhIdCBAQEBiAAiCABChkMNCGA0ADSEZQgOCllrsRWAEaCwgBGAEgASoDAAICIikIARIlAgQ6IO7X4GoH3/B39qoS7kwXqpew2/xCytzm0c+1hXCrE18wICIrCAESBAQIOiAaISAqCGxacs/wly0s3f0cm3A5hWPEqt5K3ZkZukgMN8Jh8yIpCAESJQYQOiBq45JAWFuyTlcQl3041dlnafT8p/xriWG6aYZVV2cOXCAiKggBEiYIHIAEIL8475jthvcKtmoUWgCwMuysVcVJ8A0LBt29Ppz6JFA4IBrHAQoB8BIGb3JhY2xlGgsIARgBIAEqAwACAiIrCAESBAIEOiAaISDATq3d36m92dob4c0uJUMWi53SOBVeGZi57Tlv4scCRiIpCAESJQQIOiCobCbU/uzI8AHz+5KI8jmmceLGnVLfV8D67kSpM9sLbSAiKQgBEiUGEDogauOSQFhbsk5XEJd9ONXZZ2n0/Kf8a4lhummGVVdnDlwgIioIARImCByABCC/OO+Y7Yb3CrZqFFoAsDLsrFXFSfANCwbdvT6c+iRQOCA="
//           },
//           {
//             "type": "ics23:simple",
//             "key": "b3JhY2xl",
//             "data": "CtcBCgZvcmFjbGUSIGc/gPOe8DTh4zWomosnmmV1wYko2vDCkkSY2bBCyBqwGgkIARgBIAEqAQAiJwgBEgEBGiAxNQ4L0hzy/N+VrkNUjx9H/v8IJH93ZhhXEPjufSivtyInCAESAQEaIP0NvxoPsBcWmTkGZ2ul74DYHqh4tDZivlB3QV0k8F/TIicIARIBARogtps/fryeI7RxDCD2zsLuFvtPY2nRPc6ZZTBTAVNUyG4iJQgBEiEBPUI4cA6YgD+R99n7HJ85590HWfl1TOQyVGZqghIsTRE="
//           }
//         ]
//       },
//       "height": "256",
//       "codespace": ""
//     }
//   }
// }
// */

// func TestGetMultiStoreProof(t *testing.T) {
// 	key := []byte("oracle")
// 	data := base64ToBytes(
// 		"CtcBCgZvcmFjbGUSIGc/gPOe8DTh4zWomosnmmV1wYko2vDCkkSY2bBCyBqwGgkIARgBIAEqAQAiJwgBEgEBGiAxNQ4L0hzy/N+VrkNUjx9H/v8IJH93ZhhXEPjufSivtyInCAESAQEaIP0NvxoPsBcWmTkGZ2ul74DYHqh4tDZivlB3QV0k8F/TIicIARIBARogtps/fryeI7RxDCD2zsLuFvtPY2nRPc6ZZTBTAVNUyG4iJQgBEiEBPUI4cA6YgD+R99n7HJ85590HWfl1TOQyVGZqghIsTRE=",
// 	)

// 	var multistoreOps storetypes.CommitmentOp
// 	proof := &ics23.CommitmentProof{}
// 	err := proof.Unmarshal(data)
// 	require.Nil(t, err)

// 	multistoreOps = storetypes.NewSimpleMerkleCommitmentOp(key, proof)
// 	multistoreEp := multistoreOps.Proof.GetExist()
// 	require.NotNil(t, multistoreEp)

// 	var expectAppHash []byte
// 	expectAppHash, err = multistoreEp.Calculate()

// 	require.Nil(t, err)

// 	m := GetMultiStoreProof(multistoreEp)

// 	prefix := []byte{}
// 	prefix = append(prefix, 6)      // key length
// 	prefix = append(prefix, key...) // key to result of request #1
// 	prefix = append(prefix, 32)     // size of result hash must be 32

// 	apphash := innerHash(
// 		m.AuthToIcahostStoresMerkleHash,
// 		innerHash(
// 			innerHash(
// 				innerHash(
// 					innerHash(
// 						m.MintStoreMerkleHash,
// 						leafHash(append(prefix, tmhash.Sum(m.OracleIAVLStateHash)...)),
// 					),
// 					m.ParamsToRollingseedStoresMerkleHash,
// 				),
// 				m.SlashingToTssStoresMerkleHash,
// 			),
// 			m.UpgradeStoreMerkleHash,
// 		),
// 	)

// 	require.Equal(t, expectAppHash, apphash)
// }
