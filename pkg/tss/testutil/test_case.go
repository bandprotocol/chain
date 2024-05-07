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
	FalsePrivKey        = HexDecode("3b63e7ba7bcfd7ab89c329aa572e0be73681b9387aafe906cab6515c552761b1")
	FalseEncSecretShare = HexDecode(
		"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	)
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
							HexDecode(
								"bb8257070c0d68e56e64289bff96525f67ad1f180eba8bc55a2e66319958e2decb234d564ddfdc4c15eccd9ce25a509d",
							),
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
							HexDecode(
								"3ed7c20ee24a58e975650d757fa0ac01ff266dd0ddddbc41074d4f1d2ee7156d3f3d151aef2e69037c729a8c490a38bb",
							),
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
			"Group: 2 (Threshold: 2, Size: 3)",
			Group{
				ID:         2,
				DKGContext: HexDecode("a1cdd234702bbdbd8a4fa9fc17f2a83d569f553ae4bd1755985e5039532d108c"),
				Threshold:  2,
				PubKey:     HexDecode("0251d3a86d64f1a578f4e2cdca2cc2374578c5b5c450dc61c4edc24e89d94017ac"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("cd956cec2a3e763d8635581452b86f4a21f59656f1c6b59a6b0114a29c01c367"),
						OneTimeSignature: HexDecode(
							"0256826ba9c85c45d275737fecb4ccd8645902815edd7db48a0e8bc87e4c479bb01f767cb4bf77d09de015b754153350ae0324657463858f4ad560a06959c85d54",
						),
						A0PrivKey: HexDecode("32ad23c573d211b2266cd5e710351abe00ead1cd55563bf863ed606174d0fde1"),
						A0Signature: HexDecode(
							"0210d598daa4e9874d1323a199c9b60bc063ff9fecf9d76cb2d5b33feb04d8f77fe315e7b20f4957d9ca57fac0144ba4266e28743f8a312ca7c7ee300b52d82e17",
						),
						Coefficients: tss.Scalars{
							HexDecode("32ad23c573d211b2266cd5e710351abe00ead1cd55563bf863ed606174d0fde1"),
							HexDecode("a8bb0d73bb89203897fcbf1607045025e5d1f09ead256684b41c38313da23cd8"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("03e90ce3454b707e6e8334428c4da865e912658dba71d5f501fb72b866d0eb8690"),
							HexDecode("0229c799e84406b43408a35849e301aa54ad995e0132312127cd1138cc7d761d2c"),
						},
						KeySyms: tss.Points{
							HexDecode("0218ba7f417710821dc5d4e86ba046cc083181971cde692ff7909be1499fcd03de"),
							HexDecode("0283a20dc326b4cc1a5c6c5d140019ebb88f4970557a8128b6165e1979c5581a9b"),
						},
						SecretShares: tss.Scalars{
							HexDecode("84233eaceae45223566654131e3dbb0b11dfd624005868c60c5372371fdf3650"),
							HexDecode("2cde4c20a66d725bee63132925420b323d02e9dbfe352f0f009d4bdb8d4b31e7"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode("5a623a2e84efa588797d7a4fad1765f65ea41559264755de920c55322838eff028ecfd8c84cc823f90b0aa2b0dbea504"),
							HexDecode("364f4dbd7f9099dfabfacba14a1b71788b52decc479ef6bdb960eabfa01293058817e8b044549dd6ef6e749fe2757316"),
						},
						PrivKey: HexDecode("e4b3fd82472bfd8d1d6d8eebb19ef7ec4b5c7b783305fceb4f8b98b7a58b8157"),
						PubKeySignature: HexDecode(
							"03cda14c12880e5eee98aa652ca80033539bffbc4e2eca8e7d05b5b04c73dbee19fc1977817df73ebc703119a0310ae63958065d10cbce45d92c5cfcbf5d762f59",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode("0372cbbcaec0d2ec31c3d894c23b3c7497f4383de756835fea963a5c2b5be0c7ed0356b1e3345ed7d7c5039cca70c8c8207fabe09ed51b82d8f8057f0580a43c0f7d1f2926ef027fc226b559ddefead6d3a3da851d369836e766c56885ab7677376e"),
							HexDecode("032eec8eb3e5be690a4c92260b1c95e31b2ea57f9d06ef3a5c3cbe95d01571a4ee02ae8bdd5e57282ecb88b391ad8c645de71e96509cc32bbe21c63df58d5d955bde400abb72b277199a33a400f6f5c5bc75265fdf7cfaaf11705ad9816df2f09088"),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("c6b8e5314e8dfd9b0c4880374d27f42ba50fe753c178eb42f735ae78b6b2d68f"),
						OneTimeSignature: HexDecode(
							"0380c5e12b16b2200b72ebdd950f30cd266266e424bc28511ef81aa37c6a15ca0f76a1e06f75508b700fc85fe9244ba126002280acac3f437e81a9ddeed042638e",
						),
						A0PrivKey: HexDecode("1895756c29647ca1b14e87096ca67b14ac983559f23c535402f58e0d69aef88b"),
						A0Signature: HexDecode(
							"02a381f467a0e84caae5e40d32e162c378c9adb625d583f596421060c6b3bb981146f84a558e8809b4980239e4f78b99fcbe68834d621cb084140e8849ba48d630",
						),
						Coefficients: tss.Scalars{
							HexDecode("1895756c29647ca1b14e87096ca67b14ac983559f23c535402f58e0d69aef88b"),
							HexDecode("de5839a83e4e110d7d0746f254e74d60e769b6e4ed6fa4fa13e96bbe417afaf3"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("03f19598161ea1aa499f565bab580b9e65a130bc46e08dcca7535308a5c6663d53"),
							HexDecode("02f115e4f31816720ab285f3eaf08dbd0fa875f4c3ad4395b6f710ed1a21bb5664"),
						},
						KeySyms: tss.Points{
							HexDecode("0218ba7f417710821dc5d4e86ba046cc083181971cde692ff7909be1499fcd03de"),
							HexDecode("03e3943bba202780275763d4a765435884d9b0afd078485c759baf89bac3ba9963"),
						},
						SecretShares: tss.Scalars{
							HexDecode("f6edaf1467b28daf2e55cdfbc18dc8759401ec3edfabf84e16def9cbab29f37e"),
							HexDecode("b39e2264e44eafca28645be06b5c6339ed77a03b5bfa01cabf0d142e8db366e2"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode("efa2c5ee920d6b4c704354fbb94c96812fa8d956383a9e93b3e63923fe26365a1c223188ffa46a00ffd912db5ef3f2e1"),
							HexDecode("487d0f7dfadd7fbcd8cf5b81ba8de0781f72fd4484ebca069d5fbab273216346e29da6d2a55a6df3c4635082d8780f81"),
						},
						PrivKey: HexDecode("ffbc3899b67ac41b3222d444a0a2c75e2d2d55a2aa71ed064e9be772c055af55"),
						PubKeySignature: HexDecode(
							"0329bece4c42bf6c0bf9b4e6f987a6ef05d16fc0cd557a362da4b2e6909c4905a6a31d4670bdd2b37ba2cd36b58e323c8286984a5cb3d4a6d5ebcafef48c6d703f",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode("03a27d33b3021365d1eab97bf6adef79b511cc278498113e927c891642d81f902002ac71d05629f60219d90cb67b1adfb729407092475094530a97f4939c9b0437e3a70a0e9d3cf044f3aa1ad88b91ce2796bc9a3f3d2de94177f67d7486e38e73f1"),
							HexDecode("020d6276ada49e605d00700ea66e9ef76e90de8c44094b9fa2dfa92e744277b441028b452cdaaf8c6c59b6ee5b02b89394bfe72b414bcabbedefe859c840872e5f7c98b3063beb346e658278a9ffff4b93d395b562ab5cfb349bf3d6ab2a0fbd0823"),
						},
					},
					{
						ID:             3,
						OneTimePrivKey: HexDecode("72e9945e6ccdee9a164574a213de49f6f2532b428234fb67c974ad8244f8fca7"),
						OneTimeSignature: HexDecode(
							"02f12cd884d2f0359266caab11ecbe5e7d9726804254e7e65368fdb2e8ae338190f6fd1a81882f87ae43e8bba511116871b7e8b0812d6a571e96f4f8904408af43",
						),
						A0PrivKey: HexDecode("7e6929393aa6a8ab30fceca245bf92a7bc089a2674077d83e9985b8dac415ced"),
						A0Signature: HexDecode(
							"03740b8ebea15764d9c9648049a40457732637accf05cb9dc5b76e769e2082d7b803f715b0687d6a5a7ffcb6c7180f25d2853af270a55aa7e153bda8f91bd52bf3",
						),
						Coefficients: tss.Scalars{
							HexDecode("7e6929393aa6a8ab30fceca245bf92a7bc089a2674077d83e9985b8dac415ced"),
							HexDecode("93f4f3fb75779547ffb13f50931831e889f2ec743b682513b6af67e53c1978b5"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("0396ab224ca0dbd7cc3b303f7100a71ac930c20e49dd7308d421b6f72831fdd856"),
							HexDecode("034a458d016c2841b5d5d39ad9c27bc9569c8ee72e9d2744f221bed5cb7200cf77"),
						},
						KeySyms: tss.Points{
							HexDecode("0283a20dc326b4cc1a5c6c5d140019ebb88f4970557a8128b6165e1979c5581a9b"),
							HexDecode("03e3943bba202780275763d4a765435884d9b0afd078485c759baf89bac3ba9963"),
						},
						SecretShares: tss.Scalars{
							HexDecode("125e1d34b01e3df330ae2bf2d8d7c4918b4ca9b40027025be07564e618249461"),
							HexDecode("a65311302595d33b305f6b436beff67a153f96283b8f276f9724cccb543e0d16"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode("3d1dab73ed4bdef79f6fbef372dcc8a747cb1fd8448a405c0dfa8d6971941241cca9f02d758784aa74e86147acf3df39"),
							HexDecode("4bfd9fd3e644275397009b8486011e2701e26e47aa6616e5012ff5d0c407114b4b4dd72ea5301637829c1a900bc96b57"),
						},
						PrivKey: HexDecode("1ac473b125c98aa946d8199d8fa696d1544f52e672953ce58dd9d7a10ae99c12"),
						PubKeySignature: HexDecode(
							"02b4c20e7025a241153ed11bf609ca56158ec00ee4659aba5a03ae462b84167a440e771e2ac285e26d0e5d225f0b5d741c4dd74b94c3754a8db49e5aeeefa9dd68",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode("02d0de565cc13dd0fb199920f4d4ad29a4b20b1d23f4b0197fb5f4e70ea1e3e19f039eefb2bc14b2020d4ae8d36484123da3ba8e6ea380d35f934e6c637dd40a919257228ff64901de13157fb80f8b87e211f2ebab8e189c46d3a2fee6ededda84dc"),
							HexDecode("035ba017738a2c38d1966791bab75ea3330efbac7e6a5e7cadc5a502003c7f7b9c03e8385e4825101ed8b597ace877ee73aed5c3a9b777a21aa5554157fff3698d62feca03f2d8cf15896b6af43e2e10976ee082a965d7cc627d968e34554066ff5d"),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       1,
					Data:     []byte("data"),
					PubNonce: HexDecode("027f0906568a068e3dba900e407b086fcc078284e44d50b877a0589b3354c18789"),
					Commitment: HexDecode(
						"000000000000000202c3e8303ff3ab6162e9937b974d783a71d7c99866d7389b16849b1b31988c034102da5ee04fc506dfdc4dc587efda8a71b9a8633678417a6239bff3561ed62b3bfc0000000000000003030dd45952da828c4c0870858506a325b5de0b1195fbed6ec50c1121c9c13eb04c029f8afbedab3986071f2848db288d87b733da52cbedc27d7ce3b6f7e1a4e01c05",
					),
					Signature: HexDecode(
						"027f0906568a068e3dba900e407b086fcc078284e44d50b877a0589b3354c1878953564beae1470af077ef8bf1da1a94402ef144d3b19deb3c4b01cf01896bc2a6",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 2,
							PrivD: HexDecode(
								"b6f6e21b4090c9e74be0ea9ce6e0efc2eef8ec6a07a11c21a51ce2e90acc3f75",
							),
							PrivE: HexDecode(
								"d4637d5788ce7fd6549a7f9e71fc67fe900482da9c634630e2820c2e32f6dc93",
							),
							BindingFactor: HexDecode(
								"6370f893c2c6a690167a7bd6a68e5dba820bdda6bf9918a268d9803f93a4475a",
							),
							PrivNonce: HexDecode(
								"02c503f09190581e520a288425306c6500c0618f2b6fe56257d4a3b1025139b0",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000003",
							),
							Signature: HexDecode(
								"036907a0d92d5b26e53cedcaf82159a32f40847f70babe112aeb91a2c9a459dba9274ba98cb462610a3b975e15b941f2c6df42e5aab085ee18bdab24769bbfdb6f",
							),
						},
						{
							ID: 3,
							PrivD: HexDecode(
								"f24cc73fab7b57357fdee9a63bc847cc8b0c46d480e97e2a298fd3e2df4f7540",
							),
							PrivE: HexDecode(
								"1f6cf5d526d40e5d8febf625433bd9f4e5c359c4be0602096cd98f67525a34a3",
							),
							BindingFactor: HexDecode(
								"cd5974caff745ea0e0c0188234a8ee2ac44ae70a4041c453ef8d4de84f38f2fa",
							),
							PrivNonce: HexDecode(
								"84db3e9d273ef4d72b7e4b9617f785d26fa340880226f6f803aacc4d3a9a61f0",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036413f",
							),
							Signature: HexDecode(
								"032526c7d602c2973436b3e7497fd2edbc4353a90628af64c34f71dc6cd010ee712c0aa25e2ce4a9e63c582ddc20d8a1794fae5f290117fd238d56aa8aedabe737",
							),
						},
					},
				},
				{
					ID:       2,
					Data:     []byte("data"),
					PubNonce: HexDecode("0375462a5782139d93592a51e26e36d99c8d2bbc51787f165afcfbdf81ca5f3667"),
					Commitment: HexDecode(
						"000000000000000202f4e2de38c0f3c190d99772eeee9b91c032ac83ef4e9ad8110be1f922394577b203ef2b2b3d4a0ebdb9b11168bf5b38cdd1ef1e550dad2ec93081926cae9e187b16000000000000000302ff74d305d77671b8ab63a761c6b56215a3c7335161a57b507bbab244700e0fa102ca62e66708596612e44b2044523261cb2e0f52b43475b3b439692eb356289742",
					),
					Signature: HexDecode(
						"0375462a5782139d93592a51e26e36d99c8d2bbc51787f165afcfbdf81ca5f366757220b7a95335cce51623cd25c71aa432a54a24869249da404b1cc1a0c192ade",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 2,
							PrivD: HexDecode(
								"5b30c4aa92e343416f4b8c83f446a79496cb0fe65c46972ae0b0b6d947ab97e3",
							),
							PrivE: HexDecode(
								"939b931782c40dfb161cb15a26467cd845f3179984929323f33dd7b225142244",
							),
							BindingFactor: HexDecode(
								"b081adee5d407dab4e5a768223822ef1b60105a526a7d9be13b08025757e6c6d",
							),
							PrivNonce: HexDecode(
								"e614528b09b432974fefc98a7167235f6cf439b34d8b07e4728d70e7e8c291bd",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000003",
							),
							Signature: HexDecode(
								"024e0e0836e38da513d044d9ac6ccdbbbec24749296b652b82b203f6d69b3537b24f816d040ffee0e74757ea0cb15d15ade67f279719966895400318f5159c273e",
							),
						},
						{
							ID: 3,
							PrivD: HexDecode(
								"4bf931569bd7db77cae4c7040a90aac9a8ebfec3f4dc4d84afa01d411d1f632f",
							),
							PrivE: HexDecode(
								"55c339819a78d531530afc8c93f4b53c0dd74b90e2fb930859554c68fd951994",
							),
							BindingFactor: HexDecode(
								"f38a9a43ce8ed9f27db69f329ffa257eeba1474edcce58199e776928cd7d25a6",
							),
							PrivNonce: HexDecode(
								"0b6a860495e88e1b68dc03d471320c3dad91ba2c40369614dc6bb9a67bdf2f7b",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd036413f",
							),
							Signature: HexDecode(
								"033ab01cfd5f0884b3a8d2e2feead62ae11d97f12b108ef5fdfcd2b008405055fb07a09e7685347be70a0a52c5ab14949543d57ab14f8e350ec4aeb324f67d03a0",
							),
						},
					},
				},
			},
		},
	}
)
