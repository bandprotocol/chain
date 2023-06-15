package testutil

import "github.com/bandprotocol/chain/v2/pkg/tss"

type TestCase struct {
	Name     string
	Group    Group
	Signings []Signing
}

func CopyTestCase(src TestCase) TestCase {
	return TestCase{
		Name:     src.Name,
		Group:    CopyGroup(src.Group),
		Signings: CopySignings(src.Signings),
	}
}

var (
	FakePrivKey = HexDecode("3b63e7ba7bcfd7ab89c329aa572e0be73681b9387aafe906cab6515c552761b1")
	FakePubKey  = HexDecode("0256d08999c2aae311c64396233508cde2101e234485dbda17078806aadb48b4cb")
	FakeSig     = HexDecode(
		"02f3ccc5cf138441e57a479856f50b6141435bdf37f57dad1ac0f3292694e96d0b189f73192795da234cdeb51c69821eeb982332da58e289ce1b9d9f8d27a3cd44",
	)
	FakeComplaintSig = HexDecode(
		"02f3ccc5cf138441e57a479856f50b6141435bdf37f57dad1ac0f3292694e96d0b02f3ccc5cf138441e57a479856f50b6141435bdf37f57dad1ac0f3292694e96d0b189f73192795da234cdeb51c69821eeb982332da58e289ce1b9d9f8d27a3cd44",
	)
	FakeLagrange = HexDecode(
		"0000000000000000000000000000000000000000000000000000000000000000",
	)
	TestCases = []TestCase{
		{
			"Group: 1 (Threshold: 2, Size: 2)",
			Group{
				ID:         1,
				DKGContext: HexDecode("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c"),
				Threshold:  2,
				PubKey:     HexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("83127264737dd61b4b7f8058a8418874f0e0e52ada48b39a497712a487096304"),
						OneTimeSig: HexDecode(
							"023d5cdddbdbe503590231e9a8096348cf27d93714021feaef91b3c09553723ba3c5d137db80b4642825e48c425450f14731e7cd3c2397abb4b2c70e65a70b062e",
						),
						A0PrivKey: HexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
						A0Sig: HexDecode(
							"033638414d6249831a89965f5f7fc59a77efc9335c4565cbd79f29f86b252d547a8aa2f99b06c196c7a81931b2a099ab1fcf998d115173e9def162b50180ddf2d9",
						),
						Coefficients: tss.Scalars{
							HexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
							HexDecode("2611e629e7043dbd32682e91c73292b087bb9f0747cd4bdd94e92093b6716504"),
						},
						CoefficientsCommit: tss.Points{
							HexDecode("039cff182d3b5653c215207c5b141983d6e784e51cc8088b7edfef6cba504573e3"),
							HexDecode("035b8a99ebc56c07b88404407046e6f0d5e5318a87b888ea25d6d12d8175b2d70c"),
						},
						KeySyms: tss.PublicKeys{
							HexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
						},
						SecretShares: tss.Scalars{
							HexDecode("fc93f14f4e3e4e15378e2c65ba1986494a3f54b7c135dd21d67a44435332eb71"),
						},
						EncSecretShares: tss.Scalars{
							HexDecode("d47a459f272be3d22e54af5a0a45ea8318e88f2c3c767962b2b5f9ba53d9922d"),
						},
						PrivKey: HexDecode("b248a8a2f6f1644b196402de4026d3b63db36529b2b365995f5b21eebf20acea"),
						PubKeySig: HexDecode(
							"02bf7d39a54f6d468ce71317e2d5cc87c34c4ef11ee2b6638f57b435dadd7a976520e65c8e296ff1570ad0bb4a5f18557126642e76cbda0f6ffd4a546ea4651ef8",
						),
						ComplaintSigs: tss.ComplaintSignatures{
							HexDecode(
								"02a55f7d417d1b51d91e6097473f00f528291aaa0dd11733e83eb85680ed5d4e36034946dba60574e576aef1c252e48db7c2c40f828efdb374ec8bd48ea36af06ac89fe3b8aef036713c547118f5a0adb8108dfe19b4067081f26a2fe27a87f60c0b",
							),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("e628ea45842af65d017c5c8c198f8c16741093b759f7c2145ec3b4a2c76942ad"),
						OneTimeSig: HexDecode(
							"031b25c792eebdefb217d64f632fa39b25b2ff1c3aed6889d82560aaf74daa397cc4ee436588e42749d1a24c831f285abafd52ca56f449f78fc77e8aa455e0143c",
						),
						A0PrivKey: HexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
						A0Sig: HexDecode(
							"036f96137c7d88f85a723f5933a697afd039032786f08bab2a223b1c0069f2ad07041319c804f1eb2e7d6b181a8beb791c2173451af5bbeab033e5b43590a5a9c3",
						),
						Coefficients: tss.Scalars{
							HexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
							HexDecode("42e8ead39b0d57a943cf5d7fba99da80a96eac0599bebfea0cfc5a775a6bae09"),
						},
						CoefficientsCommit: tss.Points{
							HexDecode("02786741d28ca0a66b628d6401d975da448fc08c15a1228eb7b65203c6bac5cedb"),
							HexDecode("023d61b24c8785efe8c7459dc706d95b197c0acb31697feb49fec2d3446dc36de4"),
						},
						KeySyms: tss.PublicKeys{
							HexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
						},
						SecretShares: tss.Scalars{
							HexDecode("dbc69d7d8fb753f3143e050a4d3fe01c35de8c5fe8937490dd9c5ccbf29567be"),
						},
						EncSecretShares: tss.Scalars{
							HexDecode("b3acf1cd68a4e9b00b0487fe9d6c44560487c6d463d410d1b9d81242f33c0e7a"),
						},
						PrivKey: HexDecode("1b4379a07902f9b18f9b8eefc1f340e8b42ed34fe4f6d125416e3e6cffc77eb6"),
						PubKeySig: HexDecode(
							"026604b13c5e604fa8bf0c8b2c4451469295aa465d7f4b18d1a6663548f5ffaccce75381360b70b1f9be78dbd634f3b27f47da41ec36c1dc611a4543a63163d14f",
						),
						ComplaintSigs: tss.ComplaintSignatures{
							HexDecode(
								"03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf103bb20bf99b70ae76cf2ef8779d0d88f8bb3eada6dd25f1663738f290ce9595b11110db1b2cbfc92e84076de48b1636a480fefcb1df6a4bdc4cea33d45b1851631",
							),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       1,
					Data:     []byte("data"),
					PubNonce: HexDecode("0305e39f4046be2f3c96c092ce90923e086adf1872a4766c476d2fe98bd4449e9d"),
					Commitment: HexDecode(
						"000000000000000103cd12d8f9abd0537d125fc6c998567bfd223cbdeb5ba66443f59731ff1a008aa203741dc9ba6f4876636424e02ab325dea615e262cc9b0e14404a1857b762cceba2000000000000000202234d901b8d6404b509e9926407d1a2749f456d18b159af647a65f3e907d61ef10250968aba50dbe4b8de4c1a2fa741dbee444ff48f0720b8cdb09ff33c47f34d45",
					),
					Sig: HexDecode(
						"0305e39f4046be2f3c96c092ce90923e086adf1872a4766c476d2fe98bd4449e9d48db67b34aedf4fe69d00c055a2cfa52b76baf5c0a68f79e1667fea4cae209b1",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"c51328a8409dab5115f9b081fdaa6f0271ac4482c0bcebf407c5734efdaccb9f",
							),
							PrivE: HexDecode(
								"dc835a4bd0c0e59b1aae9411a10325521b4140e2b02df9516ba8d06071c8d627",
							),
							BindingFactor: HexDecode(
								"ba578cde667fd3a7769a123a7e4baa0fe72fd323a541b95cff9c5eab86b7d645",
							),
							PrivNonce: HexDecode(
								"20c9a3904268432450487aae5fa48c14db670d50fb31417b1aeaf39ebe75e20a",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000002",
							),
							Sig: HexDecode(
								"03a6168ec50b9a3696bf690a26eae9842421f2fbf0b889eefb5eb4f7be21469e10eb8595956745253b5b2b47bb28dbbb44d4d36f91ab204d0b5e59c21c1058cb8d",
							),
						},
						{
							ID: 2,
							PrivD: HexDecode(
								"56b6a3783a58558ccc349fb3d8b33efd1184d38e80781b5e5ad6ece6067ae0cc",
							),
							PrivE: HexDecode(
								"2a9ac1414b227155b1d6ae81650f26fbf8a5807c713bf54cc80b1ed86277bde3",
							),
							BindingFactor: HexDecode(
								"cbe64511ff21632a6279ae22f31b237adac9cbf7a415b66e163e41871ec32441",
							),
							PrivNonce: HexDecode(
								"cd9eaea5226fe4ff2e05cb3897e56c9d688153fd0436085458019f587f277d26",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140",
							),
							Sig: HexDecode(
								"037d2ff91881cb049eff2532def4545a9bbaaaee4f66d24b8c559489a242aa82ae5d55d21de3a8cfc30ea4c44a31513f0c9d471cb10e914ace77e09b158abf7f65",
							),
						},
					},
				},
			},
		},
		{
			"Group: 2 (Threshold: 1, Size: 2)",
			Group{
				ID:         2,
				DKGContext: HexDecode("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c"),
				Threshold:  1,
				PubKey:     HexDecode("03534dfb533fedd09a97cbedeab70ae895399ed48be0ad7f789a705ec023dcf044"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("83127264737dd61b4b7f8058a8418874f0e0e52ada48b39a497712a487096304"),
						OneTimeSig: HexDecode(
							"023d5cdddbdbe503590231e9a8096348cf27d93714021feaef91b3c09553723ba3c5d137db80b4642825e48c425450f14731e7cd3c2397abb4b2c70e65a70b062e",
						),
						A0PrivKey: HexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
						A0Sig: HexDecode(
							"033638414d6249831a89965f5f7fc59a77efc9335c4565cbd79f29f86b252d547a8aa2f99b06c196c7a81931b2a099ab1fcf998d115173e9def162b50180ddf2d9",
						),
						Coefficients: tss.Scalars{
							HexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
						},
						CoefficientsCommit: tss.Points{
							HexDecode("039cff182d3b5653c215207c5b141983d6e784e51cc8088b7edfef6cba504573e3"),
						},
						KeySyms: tss.PublicKeys{
							HexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
						},
						SecretShares: tss.Scalars{
							HexDecode("b07024fb8035d29ad2bdcf422bb460e83ac816a9319b4566aca8031be6502169"),
						},
						EncSecretShares: tss.Scalars{
							HexDecode("8856794b59236857c98452367be0c5220971511dacdbe1a788e3b892e6f6c825"),
						},
						PrivKey: HexDecode("494dd7a574dfcee4a32c76ccbe5a66850c891a1cd12759d1bd75a6e3ae4399dd"),
						PubKeySig: HexDecode(
							"03a2e7b6e1b051079e45237b4982fbe9cb1993ab61cb274ff11d20a8f4c8b96482539b7b3e6d8e615c3b1c81ff9ef58cfbb294d3f75d61f4794d0021c83f4ceb42",
						),
						ComplaintSigs: tss.ComplaintSignatures{
							HexDecode(
								"02a55f7d417d1b51d91e6097473f00f528291aaa0dd11733e83eb85680ed5d4e36034946dba60574e576aef1c252e48db7c2c40f828efdb374ec8bd48ea36af06ac89fe3b8aef036713c547118f5a0adb8108dfe19b4067081f26a2fe27a87f60c0b",
							),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("e628ea45842af65d017c5c8c198f8c16741093b759f7c2145ec3b4a2c76942ad"),
						OneTimeSig: HexDecode(
							"031b25c792eebdefb217d64f632fa39b25b2ff1c3aed6889d82560aaf74daa397cc4ee436588e42749d1a24c831f285abafd52ca56f449f78fc77e8aa455e0143c",
						),
						A0PrivKey: HexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
						A0Sig: HexDecode(
							"036f96137c7d88f85a723f5933a697afd039032786f08bab2a223b1c0069f2ad07041319c804f1eb2e7d6b181a8beb791c2173451af5bbeab033e5b43590a5a9c3",
						),
						Coefficients: tss.Scalars{
							HexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
						},
						CoefficientsCommit: tss.Points{
							HexDecode("02786741d28ca0a66b628d6401d975da448fc08c15a1228eb7b65203c6bac5cedb"),
						},
						KeySyms: tss.PublicKeys{
							HexDecode("035db2a125a23300bef24e57883f547503ab2598a99ed07d65d482b4ea1ff8ed26"),
						},
						SecretShares: tss.Scalars{
							HexDecode("98ddb2a9f4a9fc49d06ea78a92a6059b8c6fe05a4ed4b4a6d0a002549829b9b5"),
						},
						EncSecretShares: tss.Scalars{
							HexDecode("70c406f9cd979206c7352a7ee2d269d55b191aceca1550e7acdbb7cb98d06071"),
						},
						PrivKey: HexDecode("494dd7a574dfcee4a32c76ccbe5a66850c891a1cd12759d1bd75a6e3ae4399dd"),
						PubKeySig: HexDecode(
							"023bfd9ed5683dce2b7ffae8882a019afdfa704936d1569b2fa1b4ff68f6b763139d578263e235745886e1b1e6f5c0bd5cc220eba009054a2738cb7203d6d7d7d2",
						),
						ComplaintSigs: tss.ComplaintSignatures{
							HexDecode(
								"03000b8376dd57146397b8b38edd9f4bd551dfd3d6955c1dbdad1115da1a8fcaf103bb20bf99b70ae76cf2ef8779d0d88f8bb3eada6dd25f1663738f290ce9595b11110db1b2cbfc92e84076de48b1636a480fefcb1df6a4bdc4cea33d45b1851631",
							),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       1,
					Data:     []byte("data"),
					PubNonce: HexDecode("02c8df48e5de29937cfaadf62c815a6f6c5af616d99a72f41fdcb1c2a9dd304526"),
					Commitment: HexDecode(
						"000000000000000103cd12d8f9abd0537d125fc6c998567bfd223cbdeb5ba66443f59731ff1a008aa203741dc9ba6f4876636424e02ab325dea615e262cc9b0e14404a1857b762cceba2",
					),
					Sig: HexDecode(
						"02c8df48e5de29937cfaadf62c815a6f6c5af616d99a72f41fdcb1c2a9dd30452678d1f12380abe66adb1121a35ab7cdeb62f5841a6b6da01f9978431116ed6b46",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"c51328a8409dab5115f9b081fdaa6f0271ac4482c0bcebf407c5734efdaccb9f",
							),
							PrivE: HexDecode(
								"dc835a4bd0c0e59b1aae9411a10325521b4140e2b02df9516ba8d06071c8d627",
							),
							BindingFactor: HexDecode(
								"30483e7023aa25fe2b5165444c29ea86de32cf9e8f33cfce33341ec3c77e30e0",
							),
							PrivNonce: HexDecode(
								"1bbbbdd8a1975c7ab82e7adb760a50ba054fbaa1dba719f190aea14d453540fd",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000001",
							),
							Sig: HexDecode(
								"02c8df48e5de29937cfaadf62c815a6f6c5af616d99a72f41fdcb1c2a9dd30452678d1f12380abe66adb1121a35ab7cdeb62f5841a6b6da01f9978431116ed6b46",
							),
						},
					},
				},
				{
					ID:       2,
					Data:     []byte("data"),
					PubNonce: HexDecode("028c7ab03ed162b678eca88d2b547f1f9eb2c729658789eb7d4ad8c74a019941f2"),
					Commitment: HexDecode(
						"000000000000000203cd12d8f9abd0537d125fc6c998567bfd223cbdeb5ba66443f59731ff1a008aa203741dc9ba6f4876636424e02ab325dea615e262cc9b0e14404a1857b762cceba2",
					),
					Sig: HexDecode(
						"028c7ab03ed162b678eca88d2b547f1f9eb2c729658789eb7d4ad8c74a019941f25186391d5e976e8fa084c470c6db1817745c81aced42611e36cdfa8f139ac9ae",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 2,
							PrivD: HexDecode(
								"c51328a8409dab5115f9b081fdaa6f0271ac4482c0bcebf407c5734efdaccb9f",
							),
							PrivE: HexDecode(
								"dc835a4bd0c0e59b1aae9411a10325521b4140e2b02df9516ba8d06071c8d627",
							),
							BindingFactor: HexDecode(
								"fa516f89148cb50781aac1bc1cb8b4ee363827178a32beb3987256b2c200e0cc",
							),
							PrivNonce: HexDecode(
								"543f446540e88eb7bbd8bbcc48fd09d2cba30ff34a9f443425f1f2a6095585a2",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000001",
							),
							Sig: HexDecode(
								"028c7ab03ed162b678eca88d2b547f1f9eb2c729658789eb7d4ad8c74a019941f25186391d5e976e8fa084c470c6db1817745c81aced42611e36cdfa8f139ac9ae",
							),
						},
					},
				},
			},
		},
	}
)
