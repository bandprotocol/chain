package testutil

import (
	"github.com/bandprotocol/chain/v2/pkg/tss"
)

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
	FalsePrivKey   = HexDecode("3b63e7ba7bcfd7ab89c329aa572e0be73681b9387aafe906cab6515c552761b1")
	FalsePubKey    = HexDecode("0256d08999c2aae311c64396233508cde2101e234485dbda17078806aadb48b4cb")
	FalseSignature = HexDecode(
		"02f3ccc5cf138441e57a479856f50b6141435bdf37f57dad1ac0f3292694e96d0b189f73192795da234cdeb51c69821eeb982332da58e289ce1b9d9f8d27a3cd44",
	)
	FalseComplaintSignature = HexDecode(
		"02f3ccc5cf138441e57a479856f50b6141435bdf37f57dad1ac0f3292694e96d0b02f3ccc5cf138441e57a479856f50b6141435bdf37f57dad1ac0f3292694e96d0b189f73192795da234cdeb51c69821eeb982332da58e289ce1b9d9f8d27a3cd44",
	)
	FalseChallenge = HexDecode("97e42d6dad095e552b6467e24f35740172233fbd0c2d078066009a85206e6e93")
	FalseLagrange  = HexDecode(
		"0000000000000000000000000000000000000000000000000000000000000000",
	)
	TestCases = []TestCase{
		{
			"Group: 1 (Threshold: 2, Size: 2)",
			Group{
				ID:         1,
				DKGContext: HexDecode("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c"),
				Threshold:  2,
				PubKey:     HexDecode("0260aa1c85288f77aeaba5d02e984d987b16dd7f6722544574a03d175b48d8b83b"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("0acd7f54b5d7f11148441830ecb968bb53d2485259a6b030970064724c21bb93"),
						OneTimeSignature: HexDecode(
							"0263595fa982b85c54cfe7a090684b98ad1104e8547f4427df8f2cb4520aa8f7aae598dea6ea0ea26dcc76c6df602b689cf9e94268e433fe857861dffe4b7234df",
						),
						A0PrivKey: HexDecode("3149ffc8d1fb31890c493fdb9b85fb7fb83090527fe80df27a4f8c09c2c3de11"),
						A0Signature: HexDecode(
							"02d679023006d616140c78cca2ddfdd79c168a56e5837abde46a60b98da3594db262174e676905efcfe4891efdcd23387c22a69e7e360a34e413335be391beacdc",
						),
						Coefficients: tss.Scalars{
							HexDecode("3149ffc8d1fb31890c493fdb9b85fb7fb83090527fe80df27a4f8c09c2c3de11"),
							HexDecode("e60eb5ac73febf973c30704e1ed0037b7f2f53b44c5f9041b55a3d85999a3fe3"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("02d14f38cea1abe9a3406ed84d937cb202a72f0b1742d922f1983462587e31b782"),
							HexDecode("02c3d07f122b5015e134c9e19fa72966dab8e8279966ae1882f8f0b5093756d419"),
						},
						KeySyms: tss.Points{
							HexDecode("02075b96bb760670b253b0632d3360e5ef789638c55fe9f605e7a2bc2dffae365a"),
						},
						SecretShares: tss.Scalars{
							HexDecode("fd676b21b9f8b0b784aa2077d9260277fbe05ad4695e8e3a2531a88825c21c96"),
						},
						EncSecretShares: tss.EncSecretShares{
							tss.EncSecretShare{
								Value: HexDecode("bb8257070c0d68e56e64289bff96525f67ad1f180eba8bc55a2e66319958e2de"),
								Nonce: HexDecode("cb234d564ddfdc4c15eccd9ce25a509d"),
							},
						},
						PrivKey: HexDecode("0c2ba7a1236807693a68b6115754119336e59d8890f3295d7e460bdb3902ff58"),
						PubKeySignature: HexDecode(
							"03ac6f6f47b3e8625d0ea1a7cea832655c95b679dadefef8ca4de2cdb66640496d48b72a3ca150c1b8a3715c501e9e92723f4262f0c8268868ad091eb22a769931",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode(
								"026d8cecb7d3eac4a729f0f9e50bbb690f01b20ad678409cf79376944c078b3f6c02979db8b033adc0717b669ccbff913e681659245d97b67f66133513976ca1f95070fbe5f06e141ad77c93070dd20a275c1246c1ef9ab62a4e5048b8baec8ea8a4",
							),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("5903c6a0c4063b70a4bbc2864b2f4a26ed51d75c8550c5754dda2fe0d1f0c04b"),
						OneTimeSignature: HexDecode(
							"037cccd123c9eaf857e619fe0f6e5c55d213a827301f324c6496510f0ad639c7f57c328093f056f9d5b12277b53375cbe7c4ae3541d3d4f28ca9fc2ac2ddad1a6c",
						),
						A0PrivKey: HexDecode("c90e95106aba6b0d5808f77b5bc5ff9678a59bb2e42ebaca793f2b4ae0e635e0"),
						A0Signature: HexDecode(
							"036a2d7e3e6c6c70502b8d187776f5e1b69b98e3a0e7ec92894f078ddf118a376920d51617c8a41587c798a5d21c9a1ede169d54eab05f489ddd334c93a87b290a",
						),
						Coefficients: tss.Scalars{
							HexDecode("c90e95106aba6b0d5808f77b5bc5ff9678a59bb2e42ebaca793f2b4ae0e635e0"),
							HexDecode("2bc45d1b72b3ab3b99e60e6c413812fefc3dd79c3f0e10d65501d41a9c2b2e06"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("0278a9bd9ed98306ac26f35f9783278abc12796fd0b993ee6ab85bd59f20c7eac9"),
							HexDecode("024da1fd7050eb2c554e6d05168213051d40a09f017627355088dfde6fb09e423d"),
						},
						KeySyms: tss.Points{
							HexDecode("02075b96bb760670b253b0632d3360e5ef789638c55fe9f605e7a2bc2dffae365a"),
						},
						SecretShares: tss.Scalars{
							HexDecode("f4d2f22bdd6e1648f1ef05e79cfe129574e3734f233ccba0ce40ff657d1163e6"),
						},
						EncSecretShares: tss.EncSecretShares{
							tss.EncSecretShare{
								Value: HexDecode("3ed7c20ee24a58e975650d757fa0ac01ff266dd0ddddbc41074d4f1d2ee7156d"),
								Nonce: HexDecode("3f3d151aef2e69037c729a8c490a38bb"),
							},
						},
						PrivKey: HexDecode("1dfeba690a1a723c107f34cbb75c280ef7a3ebf26d182a39c8cfbeee9e922c00"),
						PubKeySignature: HexDecode(
							"03c3befda42b4b28d5058842e45f46de7d4e4c79fa5cd36ea84af71514712842ac3e0b418701980d7601e31eeea9e683267ca1b35fb4abe9e877cbe041e6afbba3",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode(
								"02a0958e29cb00571f3ed832b146edfaf5c67bc887748cc911abc4b9375da2f0fe020760dbce0abca8daa81a14b9701368b05d5d606f02d3fc9574c28eb761d627af8bb7b85624d5e1d78db08359d124ca57322adbc80f4fbb9930f9f1ed1838668c",
							),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       1,
					Data:     []byte("data"),
					PubNonce: HexDecode("02d447778a1a2cd2a55ceb47d6bd3f01587d079d6ddadbfff5d6956ca9b7ca0e31"),
					Commitment: HexDecode(
						"000000000000000103385f1ec9f2154f088e74b04e4f15e51738c872067a92b3255dfb34489d98b90e026a2cd9c8d61fd283a4cff50afe5b15ccb573c006ffdcc5bb0663e57da3082558000000000000000202ba1c259cae61c92932162d0f8e7abf64a74b7e50eb2299cea2fca8f3154110080341a30384b1cb5d4184bfaec322c169bac46b89c1252031b1807ae3560bd2c2e9",
					),
					Signature: HexDecode(
						"02d447778a1a2cd2a55ceb47d6bd3f01587d079d6ddadbfff5d6956ca9b7ca0e317074433c8adbfb338cd69a343fc1155ce60d4f5e276975ba8bdb8ae8f803ea23",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"de6aedbe8ba688dd6d342881eb1e67c3476e825106477360148e2858a5eb565c",
							),
							PrivE: HexDecode(
								"3ff4fb2beac0cee0ab230829a5ae0881310046282e79c978ca22f44897ea434a",
							),
							BindingFactor: HexDecode(
								"1e2427a0bc6c3834c50525d0ae243d120920951528242430627b54e13aeaa5a6",
							),
							PrivNonce: HexDecode(
								"aede07cfee6c2228e976da951140d3ccd13eda2cb9a46928383d4554d252f438",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000002",
							),
							Signature: HexDecode(
								"03ead532525ed81efa287760719b8068a008eb4df26c8600d04ee0c0b007dc6319e15c1618f6e19cfd5cc2e75aef743483e103e303d4fb8abaca456d2040fc00c2",
							),
						},
						{
							ID: 2,
							PrivD: HexDecode(
								"69b349465d7b23ad7a6a02af6848b3f61d1d2f8f21ae2d3dd075be0491ca8987",
							),
							PrivE: HexDecode(
								"bbe729556f9c1dae4e7d3b70a1008db49e29fd388896c697fb6b3796c5b25e1f",
							),
							BindingFactor: HexDecode(
								"18bbd327bf86730736be4b1bcbd0aad28aa69c9e65060f3685f5d5da7e12011b",
							),
							PrivNonce: HexDecode(
								"2e85c46f9ac03c8a355daf4071a40372ea7abd65711d3530e9dfbb23c407ebda",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140",
							),
							Signature: HexDecode(
								"02002550ad6f8d5ae09cbdfc4bcc232e9a55cf1b365c19438a86de4c14c1d41a8a8f182d2393fa5e363013b2d9504ce0d7bfb8494101b68b3b81687c55873e2aa2",
							),
						},
					},
				},
				{
					ID:       2,
					Data:     []byte("data"),
					PubNonce: HexDecode("038778d7fcff5031106334e35ec7b5253602e558e373abce8ae26dd852816afa56"),
					Commitment: HexDecode(
						"0000000000000001029338ec269dda11b2f4e8aade43490952a03a3bf95083b6b2fdfac0b539c1af5c03a142e008766562190aadc219df154b39a263cbc666417ba310dabb6f27784162000000000000000203d8cca0aa79c7a8cef078e5af3bb278f7d4e11fa2c59a761aff3c53bece84bb5d028f88bccea2df03a2c9e33f35792a9592c87cfc150f2bd4d2f3cf65193fadf01e",
					),
					Signature: HexDecode(
						"038778d7fcff5031106334e35ec7b5253602e558e373abce8ae26dd852816afa56f2df592b90703aecbd438722cb2287db1233b59feae2e097666922a591d7f2da",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"50864fe87972706de3edbd63ae3b53dfb15e9763800a1affd3d679ad1f3f8202",
							),
							PrivE: HexDecode(
								"6fe909176f15181663e78e272fae08756e2670e279d6727d9621c5038ef6ba1c",
							),
							BindingFactor: HexDecode(
								"18190219c71e064b23b504aa17f0bcec5367df260806644a6ff6c6eb695d444b",
							),
							PrivNonce: HexDecode(
								"735a0131cd1a9668e3eacf9c2d0fe55812804ce7286eb6b8b85533f8a397f342",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000002",
							),
							Signature: HexDecode(
								"031ab5312c34cce0d8b94df401225c01cce3bddff93442a1101f4369ef42c31fc952106877fe098c94920e4712da8cabe5a2a5e09ce1f3ff32f0dc8324816554b3",
							),
						},
						{
							ID: 2,
							PrivD: HexDecode(
								"47a19b095e55e752eb01a3d640cba19db5004cababb571ba6942b457fb19db9d",
							),
							PrivE: HexDecode(
								"9500aa9d765cbab6c7ac323efefac6ae491d3e8369f3c39ebe355d33ca772bdd",
							),
							BindingFactor: HexDecode(
								"e17aeb5ff5b87aa135ca410a9885473aff4b28313db0173d6a14a5bf9199c04d",
							),
							PrivNonce: HexDecode(
								"85de24d4bb6855353919adb97697a5c4259f8d4dc816117b8449d85948e329ce",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140",
							),
							Signature: HexDecode(
								"02ba40075664b86857f6df7ed8ed2a5ed16f9a7ab4bfb3c7a8b5281acbf5ecadfda0cef0b39266ae582b35400ff095dbf56f8dd50308eee164758c9f8110729e27",
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
				PubKey:     HexDecode("02a37461c1621d12f2c436b98ffe95d6ff0fedc102e8b5b35a08c96b889cb448fd"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("b688151a692b3742303d15a820f48df74f169e1f335c0eec32a9062485fcec11"),
						OneTimeSignature: HexDecode(
							"026905d1d2af2fa42f7118b1c5d3ddf65e562485aa4bced9ef16ddb41dab4962520b1a37f255e033ecf2b499b80fe603da131f225e2c514c1facfcd60151f6ab62",
						),
						A0PrivKey: HexDecode("75455ebc9f9f6aa3471da39a85a1283df7451d4e9b3c5c807bb10dd0c14bb051"),
						A0Signature: HexDecode(
							"02cc374c42e50675c8942de9dd6644a0920f78c9bde40f9020337f42d8db918eacc5e5ab6fc039404098c2e12998ecfedb3d7fa8c5949f1409114a096ec347f957",
						),
						Coefficients: tss.Scalars{
							HexDecode("75455ebc9f9f6aa3471da39a85a1283df7451d4e9b3c5c807bb10dd0c14bb051"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("0262f4888feb37b89c9744444093f6c00b0e196087c7c28671bb8247a32eef704f"),
						},
						KeySyms: tss.Points{
							HexDecode("036f9072ece9b94f22a5547acd536e25c609fb065f276104444df535ea78288198"),
						},
						SecretShares: tss.Scalars{
							HexDecode("75455ebc9f9f6aa3471da39a85a1283df7451d4e9b3c5c807bb10dd0c14bb051"),
						},
						EncSecretShares: tss.EncSecretShares{
							tss.EncSecretShare{
								Value: HexDecode("fa00ec78f1dd3a82b12abebaa084251bde4879e295cab318ae9dcef42cd54c95"),
								Nonce: HexDecode("3f3d151aef2e69037c729a8c490a38bb"),
							},
						},
						PrivKey: HexDecode("97f0af1189e626c1bb22a0611c8947567b6fb5c29a543669a8e877168b806cbb"),
						PubKeySignature: HexDecode(
							"0398ba0bae0d08502c7112f12d153e3044779ea1c7b4be2931655c906f2e515c4b65904c694643fa9ebe18b28c0852dce2aca58712477a1631651096fefe4ef2a3",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode(
								"03f8d441370179701150e6c87c4e12b9549e8501b6bee9456ed20001a292f495e1034b8a8f569c8a1a3b0a25c963a85d526dafe4c01ce71a7d15db445098a21fdb7bff5bb38f970fd3d6f33161a1eec7056e3b5e9bbebc9f8b617750ce545ee0173b",
							),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("f11005581a9438273cddb8c00cab5e137b530d30ec874593050d194a328c3152"),
						OneTimeSignature: HexDecode(
							"025a6024eb03400c5bb2758170194b561ec888dc90b0f343cfb193088877d4a6e8f7b1927258296bf023b2cda5b1443fc6b902cd371f3d0484b01211eb09b5968b",
						),
						A0PrivKey: HexDecode("22ab5054ea46bc1e7404fcc696e81f18842a9873ff17d9e92d376945ca34bc6a"),
						A0Signature: HexDecode(
							"039a483f79423309cae3c4b05b03327c6921411579647cdc2a80a6079ec375dc0b41142f756b007b087faf736609ac670e710003264fbce2b8b2cccc0ba35609f0",
						),
						Coefficients: tss.Scalars{
							HexDecode("22ab5054ea46bc1e7404fcc696e81f18842a9873ff17d9e92d376945ca34bc6a"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("0369f21b7c2841750b8ba29e8559147643f8f49c41b54be2f4f7c453c84b0bc929"),
						},
						KeySyms: tss.Points{
							HexDecode("036f9072ece9b94f22a5547acd536e25c609fb065f276104444df535ea78288198"),
						},
						SecretShares: tss.Scalars{
							HexDecode("22ab5054ea46bc1e7404fcc696e81f18842a9873ff17d9e92d376945ca34bc6a"),
						},
						EncSecretShares: tss.EncSecretShares{
							tss.EncSecretShare{
								Value: HexDecode("57e9ea849d1a4efba136a950eda2d0119b753a89a42e98b71d0de49249baa137"),
								Nonce: HexDecode("c1e4d34f382090758f29da4c0a52173f"),
							},
						},
						PrivKey: HexDecode("97f0af1189e626c1bb22a0611c8947567b6fb5c29a543669a8e877168b806cbb"),
						PubKeySignature: HexDecode(
							"02aac6b21e1e05461544f933a2c4b5bc3e37380462f855ebaf62cb3e867514a9df098c3ac9c120d7a4c1f2ce1e0cdc109da41792122440c0a2669d084d96c90441",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode(
								"021ab11f3eae2807ead59d5e8cc65d0841cb324de5988be2768ea38116ac130af50399487cefdecbcf378edb1533d41c2640fa35f9e8e98f862ec7108d135e243aea0f1bf8672686f3c8972de0eaeeb989558c1fd01bea58e7c6cfeaa1f143513b55",
							),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       3,
					Data:     []byte("data"),
					PubNonce: HexDecode("024a00a111b8231ee0b4995f1f97280a3b22f2db298bbc31b1ef07adfd286bf955"),
					Commitment: HexDecode(
						"000000000000000102a1233c2d1ef055f8ff047518d35e64f2b1b938b37ba3093599f5fe41ce18080402228a0cece369591d357e6c89f29eb779b755912418fd389f94b9060b17ca49d2",
					),
					Signature: HexDecode(
						"024a00a111b8231ee0b4995f1f97280a3b22f2db298bbc31b1ef07adfd286bf9551c5000d5b8811a82eb25e139727d3bbbc04affcc36a88178b7cba24db58647d4",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"ff92bb0c646c6aa0bfeda96b76d92a3636fa3481ea8acde486def423a76bf7ff",
							),
							PrivE: HexDecode(
								"ab1ccde9bfa9dfba4a6b07549c9e6ed67703675d3593f7ff93e34a8c48a9ad56",
							),
							BindingFactor: HexDecode(
								"c2209639bb4beb23c553e593174a22c90806d81ba710077a8f78607a40923bf1",
							),
							PrivNonce: HexDecode(
								"41eafa30986f64ef8b7f842916cfd832ee6fda2bd5a54a588f9233d6274e80d7",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000001",
							),
							Signature: HexDecode(
								"024a00a111b8231ee0b4995f1f97280a3b22f2db298bbc31b1ef07adfd286bf9551c5000d5b8811a82eb25e139727d3bbbc04affcc36a88178b7cba24db58647d4",
							),
						},
					},
				},
				{
					ID:       4,
					Data:     []byte("data"),
					PubNonce: HexDecode("036b790561a9c631033a9f8a18e2e200f713394be93ff51131d47f43755a181a2e"),
					Commitment: HexDecode(
						"0000000000000002023533bf77c50242fe743daba7292644f9a71c34ae9b9d7b8a22526af45828d5f4034df4a62cf3d8e0267dc2392651ca7b544e70abdcfc770b07d1fcb1ededef34bc",
					),
					Signature: HexDecode(
						"036b790561a9c631033a9f8a18e2e200f713394be93ff51131d47f43755a181a2e8d0b88adbbe74eb40af9f00f2d9a43853f82334c30b7bfc2e13e951612292439",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 2,
							PrivD: HexDecode(
								"43cb56ddf0a36b601480ae041d4aca6f504215993191a605a70645110f64f975",
							),
							PrivE: HexDecode(
								"097bbd65bb9b82f138eb99c569714e692fe1d18fe256cf37cec83a969892ec27",
							),
							BindingFactor: HexDecode(
								"218c8a95e82d00d970f919d99fca9ebaf0833e0fd592533336b1568530c841f7",
							),
							PrivNonce: HexDecode(
								"afb2c697eef9fa1f0efd479ec9b56b79d050dfee71b8fb7ccbeb7b4d87d9de3a",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000001",
							),
							Signature: HexDecode(
								"036b790561a9c631033a9f8a18e2e200f713394be93ff51131d47f43755a181a2e8d0b88adbbe74eb40af9f00f2d9a43853f82334c30b7bfc2e13e951612292439",
							),
						},
					},
				},
			},
		},
	}
)
