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
				PubKey:     HexDecode("030b03a4e74e06e18de6bfd16a06e6401bc1fe74a983817c4ac3c1e2f4048e0a4c"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("6ecaf984d4e1e9be0e6c187267a22410f7c6afc5c97e55e7e53e24b9ed5dc181"),
						OneTimeSignature: HexDecode(
							"0385d2f44d3b4c7d7a154a95ef2fe3710dc4ce25105c8e7117a005a25f5503412b654f52a7fb185d2ea7d2ef344de505e24a8afa9d3f3dc99704c20ab51567ae8c",
						),
						A0PrivKey: HexDecode("13b799faf18186a813df0e8cdc06bfe7dbbff5b285a0db4af211456c65a0dd83"),
						A0Signature: HexDecode(
							"02a48aba514be865fedb831675a58bf001ba30883a7616649b2729dc89c0b9e3b615675760cb5532ce0a5bf424be0e5df05a692099140e2e66cf1d2a27cb866cf1",
						),
						Coefficients: tss.Scalars{
							HexDecode("13b799faf18186a813df0e8cdc06bfe7dbbff5b285a0db4af211456c65a0dd83"),
							HexDecode("0e1632559c62d67bb8a5b962a32169555d464b489bbd9c5bc39f5cd555134da6"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("031795a75bd66ac8d9b25352889bf0aa41b7f4eb56ffa3397c80d51893d3a28524"),
							HexDecode("02739a6ea75df7d9d1dfb2d14e5b9e38961eb0bbcc079a6c87fb12f3d03650aff9"),
						},
						KeySyms: tss.Points{
							HexDecode("034e977c0e89f09aebe0eb0d896b1fd71153e32f6ded23e8f4d6a8da9e986e6aaa"),
						},
						SecretShares: tss.Scalars{
							HexDecode("2fe3fea62a47339f852a815222499292964c8c43bd1c1402794fff170fc778cf"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode(
								"a63b103ac417f5715b831a34293f73b990659687a2f34dd3e247dcf04aceeedfa4966deca4300dab6895c2242dcbfb04",
							),
						},
						PrivKey: HexDecode("335c141471a232212db60ac01be2e02f62e9a3575da75d343bab5851470427d2"),
						PubKeySignature: HexDecode(
							"039f380a66d8ceff0cdd750a910b4e401d1df8ea7fc09f64c6543c68d328d171f96b72af863626bbe1632424793313dcbbe96e631a422b2a73925a5ec9949cfa31",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode(
								"03b59f97366fb4042697306d8428643f4f3d5ed53f9958bbd81a4270113be8c3cc03a8a8535598913f0893ae94438823783e5ec63468b829542d2ff4eb029d835a1fbd145d736361bc223bfb8bb3f15f7ebf425c7fd561773bdade6158192d1570e6",
							),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("7b6dda85fdedf3adca26a64580a67ff1cdeb0c32ff609cfac94ab147e2ecd124"),
						OneTimeSignature: HexDecode(
							"03cde0ccb0345295ac3211a86483fc8a6de9c08f132c5be591aacd19d2e8f8970ffd95f17aa91af994b26fed6dd8b59891ef8df34d81aa8fecc25d101e7c2b6f64",
						),
						A0PrivKey: HexDecode("d71865279f2332653ffe7c928d41c7934604656688d023d80e3c9a8787bde6ad"),
						A0Signature: HexDecode(
							"03c760e8d727684d3933e1e48a2dee6fe0bb285665ae887342830d57d86edb7616c202e2cf7d6e27759a75004283a1a4177b2feed21c9a750181345a4d5ee4bf67",
						),
						Coefficients: tss.Scalars{
							HexDecode("d71865279f2332653ffe7c928d41c7934604656688d023d80e3c9a8787bde6ad"),
							HexDecode("3a75e29c449aa2982132c63e0f78ef5d9e8dd9dc62c161f137907a14d4c8573d"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("029c7530f841d965055eda3fe23f0ae4f4f64b6fdc576c0ad54fda0d928965258f"),
							HexDecode("02f11f888a4e1c90a063a2486b6c44c96f7b82896e2ff25f7e6ac20e0c72dd97cb"),
						},
						KeySyms: tss.Points{
							HexDecode("034e977c0e89f09aebe0eb0d896b1fd71153e32f6ded23e8f4d6a8da9e986e6aaa"),
						},
						SecretShares: tss.Scalars{
							HexDecode("118e47c3e3bdd4fd613142d09cbab6f229e3625c3c48e58d85fab60f8c4ffca9"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode(
								"dc7c17c9b6df6933f53469e30eca4e93c455db6614127cbcb7f037b4bd5a46f2461b91c795df51f6f0816ae766f08e41",
							),
						},
						PrivKey: HexDecode("7be82906529fab35078e8a60ce7d38e25ebdc87c5c265b8136db2f3b70dfccb5"),
						PubKeySignature: HexDecode(
							"0284e96304c8ee957915060e70694f14c83dbd3f1170c2218395326fdfeb5c7f521bf996f3bd368e9cb42e633d5ced36fd8c80ff2c38b6a8e019aa570de0acd8ec",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode(
								"021ae2557f6d97a61dc4d7ea51967e708ebc79362c94475bb08d790af213466a8002cab49b6367c24cf4d4b255f52defa18cccc38b9fbfcfe51949c8f7a27d30e78bbcdbbafc0a53e2359e4111da87539e9170402d00d13208b740fc39b896719ec5",
							),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       1,
					Data:     []byte("data"),
					PubNonce: HexDecode("02294d453a91d5bc13bf45737de1ed4ebd73fbc7b15e6f00d35f0f772f5fe569a1"),
					Commitment: HexDecode(
						"000000000000000102d812ecb326472267d35904bdec1cd9452fa1a1700817738bcd8c95008a052e4f03351f67cd1c991fcfc8c49f6d33a96e394f69c81a0f8721e67fd74d18bced125c00000000000000020328ce284017d252a011f48a239f798dab414560ba14a341e16167cf5e1e71cc0703cadfa654865e228b3927d3c88918bc76b3c68137dd59091e65079890c1451783",
					),
					Signature: HexDecode(
						"02294d453a91d5bc13bf45737de1ed4ebd73fbc7b15e6f00d35f0f772f5fe569a1df8de9f4f2a046eb25ec45194ad2ed4b3d2339baeda82c97dd1af02cdd63f98f",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"945dde9bb7acaeafa76ae62dbb8e5119ce0e5fce72e17ac61028818d1a428d7a",
							),
							PrivE: HexDecode(
								"245c8d210ebfc96d8559da38e684508fef98f1738f948d40e3650241dfd7a503",
							),
							BindingFactor: HexDecode(
								"ae9a1b68e40a3551c81e45e2498c718d1d27e468ece35202a8698db5c86c6869",
							),
							PrivNonce: HexDecode(
								"60b426005721d44ec75d5303511559d2c572d8434365150e83609cac2ae57dfb",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000002",
							),
							Signature: HexDecode(
								"02a7c6ad47bb5da9add28632e5ee77bbae2c354f5247493428218b9af4d4e13f8e251e3ef7240b42e2f2f04b6a2be6c69a5463d0a9c2ecc46535ec3b4c55b4e330",
							),
						},
						{
							ID: 2,
							PrivD: HexDecode(
								"70a80248f58a088beb79bd3b7dd99c7e428ccd5e7b222963d8bc936479c5b5b0",
							),
							PrivE: HexDecode(
								"0144944e9a84ba6b6d3907394ca4bf3c168ecaaace8bc694a3b6e0956c6495f1",
							),
							BindingFactor: HexDecode(
								"af0395791336a8b3a215f8737b995bbbf6f3524422a303864fbc6727dbee49e0",
							),
							PrivNonce: HexDecode(
								"316738d7d115f20eb69e3dd2635f00fe291eaff5d4322e0884795e3213f56580",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140",
							),
							Signature: HexDecode(
								"0266f6406ba8d0b2ef3ec04a161e0098a65199398ed3f528d361ec0a9b5976ea4fba6faafdce95040832fbf9af1eec26b0e8bf69112abb6832a72eb4e087af165f",
							),
						},
					},
				},
				{
					ID:       2,
					Data:     []byte("data"),
					PubNonce: HexDecode("029e75f18da212a078ce622a04d170e29ddfd7f5ed0fc6f95e0a71ebf3fdf6aa73"),
					Commitment: HexDecode(
						"0000000000000001032a970f94d5caa344cc48eb13a72b4792fd5dc48c55ee9d5b762049491f35b58e029df4d54b642db7f57f1a9d7efd9acbff6ee6de44cd7489e8084ed7f598f53eab000000000000000203ba1217cc684582eddc12b8d9b50760e3653251ab86d919a00491ae4f42e32c180218d7128df78bd841504d7304b7f0365cf4994d3b10be5e0263d2f3bde26d6713",
					),
					Signature: HexDecode(
						"029e75f18da212a078ce622a04d170e29ddfd7f5ed0fc6f95e0a71ebf3fdf6aa73f1c2dbc97be81e4da5e9118d0bd68d9eae6ec4b703d592fdc9abebb6446c76d5",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"4d20211951dae62dfd35ee38e922f2407755a8ab1df0be356cb2759fc13712d2",
							),
							PrivE: HexDecode(
								"30f21ef68b5c900b1d88e2484812e49b6e9aa72e06f0852834d62297c6e170e2",
							),
							BindingFactor: HexDecode(
								"231aa97b75a049cb87fceb9fbd0c3c4cc44bbeead1b83f040139664d7b9a5c33",
							),
							PrivNonce: HexDecode(
								"b062f612065ccec2e9fa54266406e681df40356ca246e8bc0340d057774fee8f",
							),
							Lagrange: HexDecode(
								"0000000000000000000000000000000000000000000000000000000000000002",
							),
							Signature: HexDecode(
								"02278f75f49b19cffafbe6e323bc046612931a051e868b3dc8fc68cbf210ce85847eedd75666989cb8b0465b5be814059f6971156fa4c4b1aa0c014dbb8ea6e3f4",
							),
						},
						{
							ID: 2,
							PrivD: HexDecode(
								"ea3169d415696cccd3abdacaa8ccd4da1cf2ab7832e1d9da3b078cab85cfb28b",
							),
							PrivE: HexDecode(
								"2fa7c6d495f7711ac6512ce2dc07b25242897fa68f2af8634e56deca1e0d7a4e",
							),
							BindingFactor: HexDecode(
								"ea9975067ea87d94fd3312617ba61bffad003d0ed595c4b72c683ff44c5f4e69",
							),
							PrivNonce: HexDecode(
								"32b80afd035a3343220449f3cf27e5b102cfea7c83487cef7b9e4dbf796772bd",
							),
							Lagrange: HexDecode(
								"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364140",
							),
							Signature: HexDecode(
								"03fba5d6bd33a24e4909ae217014dc487826bcb2463c435dbc509d5a3561d2572872d50473154f8194f5a2b63123c287ff44fdaf475f10e153bdaa9dfab5c592e1",
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
				PubKey:     HexDecode("02313153ccf3c45a9c64bc7e4f761d9912f2378420f66808a47e2db84b2cd987af"),
				Members: []Member{
					{
						ID:             1,
						OneTimePrivKey: HexDecode("93c8671dd031fbccdbd163d84ed57f4f5fadf9204e780fe24197bdccfdb9640d"),
						OneTimeSignature: HexDecode(
							"0391e0de325692cbb0c7e8f372daa5b35693eecd4c39f96bae45320e0f7a9808326c123b1fade7174a76c2c632537d968e63df6e7a3a86a29dccf053c42fd1df7f",
						),
						A0PrivKey: HexDecode("e349800cc66f41159816ffe8f99b0b1f7045807fae1a11c4358d1e71642b3367"),
						A0Signature: HexDecode(
							"02452ba9f19551c9456ef9912de82cc13ebab3321296f67a0229d5020f583405880b2ffe50a90828da8bf1c301b3562ee5b78c56b130cdcce6a0263e0a21ef4817",
						),
						Coefficients: tss.Scalars{
							HexDecode("e349800cc66f41159816ffe8f99b0b1f7045807fae1a11c4358d1e71642b3367"),
							HexDecode("73ec56071ca7c56ea583af10fce43a697861a53b9899b71979377f99c7465d31"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("039b725d6901e2fd190d6a8b1b9138a2231a17f1a7bebb96835fe97cd37bd321c9"),
							HexDecode("032ad8a5475329b4016a1d042f3dbacc15f92761035415a80e88e134349a08f48b"),
						},
						KeySyms: tss.Points{
							HexDecode("02d5f5bb5aa662794549fae09612960cfa3f336dd075cb9b9da72ee8250c55c20c"),
							HexDecode("03753b25ac4d4d4dee982716e0545b8636c863ed2316c9dacfc84a9cd6c97104d5"),
						},
						SecretShares: tss.Scalars{
							HexDecode("cb222c1affbecbf2e31e5e0af3637ff3a659ee103004dfbb6829bf182281ac88"),
							HexDecode("3f0e82221c66916188a20d1bf047ba5e640cb6651955f699218ee0251991c878"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode("575041969918a28ae5e5d8f98b0ca2c282f7bdb7256a2b0b576b584383752258ddf31d33e250bd967bb34fd0e1ec2d2a"),
							HexDecode("d5fa2b4d566a3fb5fc9c58a99d286879725effdecfa181ad45a04a4d055639e42f16febd9964dd54bf656379e22435b7"),
						},
						PrivKey: HexDecode("09fc1df95335655b4e9aed4d64c79c09d665584b976d8859b5c6f510399cb136"),
						PubKeySignature: HexDecode(
							"030ae7b396765aa65b51e94af6aae2e6db4e65e1d74ce26f51902f4c51cca92c575af8d5ca34bb00aa893e1cc45149ecccb68b320890738e8828008f6a49e5b8e4",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode("024d1e6cc957ca6ba56e0419d89d8d5dcb936ea7c00ee18dbc5420ccdc9e0428df03bb8c37af7ed4fd078da506c8f764a8637d22062c0cecefc388837a70181186da44917e2d3bfe997481b848ab21316f6259b92b73e27b5750a3e5d27a2bfdc829"),
							HexDecode("02d70e90c64addc17abdb54e0a4368568cd1df973a8c6a5de04244e32c3ba0d55703094e84ec64c24c75c9de536aefc1bdc5e1857a9ee911a34b43962ea8e562673fea847854ea66731a296ea776cf28b250b20de5c1f8cbb8169ea4657817ff5540"),
						},
					},
					{
						ID:             2,
						OneTimePrivKey: HexDecode("071f557faf8d5b7ff7df573668f6fad9f418e95593b7c05119f78889f56a825a"),
						OneTimeSignature: HexDecode(
							"0226df4b397f5e0ee1d7f08a49061094fe2046b7964a3f3244ec19a8f6934330658e256bbc88ba3222530ef8e3e3e57893eccab8cd5c618aff14db087b21612087",
						),
						A0PrivKey: HexDecode("1cbff1dd099afbe0effe8375875390687a74cc431d08b6f069f2401eaa6a238a"),
						A0Signature: HexDecode(
							"03eb74e4d6083f175d1c0ca4cbaa580a0e627e1dc50094b3e3eaf303c32d7202fe5baa50670df69956213de016a1c72bfa5fd8181a2cbb46e469da9912a1cd89cb",
						),
						Coefficients: tss.Scalars{
							HexDecode("1cbff1dd099afbe0effe8375875390687a74cc431d08b6f069f2401eaa6a238a"),
							HexDecode("4377afff407eaee0bd7622c7ea69539ea94972203970ad5c130031b6fd333ec3"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("021f50a27c36f1a9d360983eca9ec05ba43423785c719833e771c27ceaada2c058"),
							HexDecode("039bdaecdf749a20b9af986cc51c3e4eb3313c20372231e0a0b30b423710a6538f"),
						},
						KeySyms: tss.Points{
							HexDecode("02d5f5bb5aa662794549fae09612960cfa3f336dd075cb9b9da72ee8250c55c20c"),
							HexDecode("02222fc844758eeffff9eda3c19a36462a239b4754b8f24be0dd48fcee2a208d44"),
						},
						SecretShares: tss.Scalars{
							HexDecode("6037a1dc4a19aac1ad74a63d71bce40723be3e635679644c7cf271d5a79d624d"),
							HexDecode("e72701dacb1708832860ebcd468f8b44765122a3c95abf04a2f2d543a203dfd3"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode("4d8c12fd41994b481d41ce26516cff381260aa9af55ea16f3bed1694586591e08e7c85930058fc7af786d09ec17d395b"),
							HexDecode("4748bafef90b2e5aa08446a579b331e950f61c74cb363417f9ebf9ce82056c0f971add7561c6c183cc4177d33a05ec7d"),
						},
						PrivKey: HexDecode("bb4dd04815ba94c0ad3430b8b3343d998379e272fdfdd3b71749d43e2df1574f"),
						PubKeySignature: HexDecode(
							"02d3fbdf3584c8e4c4d718cbe7ee8c864a8b7f0deb4efb513ec57700af470e3fe0fb97c1daa2ebf3396428a54f67bf32c8afa9d089d956b5f8845bc21fce5e4f1b",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode("0310ce5bc136008f58076d277ecabb918e3f7a09eccfe1ff4611e955016844570103e2707c67caab7b2fe7b84c472946c1390c833e846ee8f89da2bb1a79089c8fdceb0a74cd60bc1318408e8225175b368288b1bc4ae6a6532511a3055547d865fd"),
							HexDecode("035e7ccb6be54ae50d3278766be346f065ce2b8f8d4c76fb667376fbd3525244620280bd4aa15ffde37cac4f5259cc71b41911f562c308fb286e03fd499e29303285c4519b4a6d5f15c4f055af05f2b1dd71cf50d0944d906f9d7895f8266c60c197"),
						},
					},
					{
						ID:             3,
						OneTimePrivKey: HexDecode("ad3e4bf80d02a4df9e7962c33ece0c0e3693b67c9ebba42cd9033a488601180b"),
						OneTimeSignature: HexDecode(
							"0245661fef4e39e23a5bfcdb3545c83090cc8a97d4ba598f4c2f2a8e7eb9545eb664387dee2525675d53294311b667789122ec31b459a4ce4bba9d102f9c9d820f",
						),
						A0PrivKey: HexDecode("58a0f9c0c0a5f8ff67ec2683956c5eefb3f43b2ec44bb4bf3469746bd71f36ae"),
						A0Signature: HexDecode(
							"039186adaf71b62ebfaeea8708f6c4e828986c06ffe169484af76856c6ea9c01858f7761990f411411d81f8b21bad895aed3f9494ffef65ca1ba3c7e2ab09538b0",
						),
						Coefficients: tss.Scalars{
							HexDecode("58a0f9c0c0a5f8ff67ec2683956c5eefb3f43b2ec44bb4bf3469746bd71f36ae"),
							HexDecode("f9edac48655ebb15fb9f7192671f138646184fb243ce8723951d8c6a00114b66"),
						},
						CoefficientCommits: tss.Points{
							HexDecode("0259fb53386b582582d9673c1f432988ea4a48f160c71d9fde6e2e082f59840352"),
							HexDecode("03891c6b2a768eee9b7c025bd233b83796448ea4db830b61a78eaa41482fc5fac9"),
						},
						KeySyms: tss.Points{
							HexDecode("03753b25ac4d4d4dee982716e0545b8636c863ed2316c9dacfc84a9cd6c97104d5"),
							HexDecode("02222fc844758eeffff9eda3c19a36462a239b4754b8f24be0dd48fcee2a208d44"),
						},
						SecretShares: tss.Scalars{
							HexDecode("528ea6092604b415638b9815fc8b72773f5dadfa58d19ba709b4a24906fa40d3"),
							HexDecode("4c7c52518b636f2b5f2b09a863aa85fecac720c5ed57828edeffd02636d54af8"),
						},
						EncSecretShares: tss.EncSecretShares{
							HexDecode("20ae6ce299bdfc625862affd3038dd36c07ac0061485118e06b90f26bc158de915a5c674f739dbaa6bc4ec8913d5e6cb"),
							HexDecode("917461239c5121b76d28e1fc802610a0ebf457f9a5adac5f90a588cce949a287a438cd98c518a8303a9c29cbd67ade6e"),
						},
						PrivKey: HexDecode("6c9f8296d83fc4260bcd742401a0df2a75df8fb3b5457ed8b8fa54df520fbc27"),
						PubKeySignature: HexDecode(
							"02598db295bbd1a3d2db5cb3ccdb866e454cd6c5ab585983ba911d691e77e4debd6a1b161a3391b1c37ceb6f89ca31587183308abb202135190fa9c7c1b5d1ae81",
						),
						ComplaintSignatures: tss.ComplaintSignatures{
							HexDecode("028b6cb8924663f82e473466a4c487679905b6945b1401515e584844df1537389e033b735b61be5740fbdfe305e7d06ae72e3fa44a66c9c21137acb75f7fd9c04cae64e578e511f724043cec9c6c351387c6cc8c97cb658cd2b4cd358d921fbb8f09"),
							HexDecode("0358f10bf71898c727aefb43c8e6acb38b9a2b62f72d15a74149284c7f81a1786d03e2707eb85d920cfc3cd179a9b888141174fff9d1db864cb91d64f2bfe5de5c4887c1a0c264434aebb553a2610622e8ae0f5cadd5c01b8fa3392aff130174bef3"),
						},
					},
				},
			},
			[]Signing{
				{
					ID:       1,
					Data:     []byte("data"),
					PubNonce: HexDecode("028946dc95a92a9ae9bfa9fb4abe5eb23f86720644ddfcd143e0ca0ab835b4cacc"),
					Commitment: HexDecode(
						"0000000000000001023819d6863539fc282d85325286372d3af73c0470e12df80a2d3fb5d094b52fe60325ec01f1a8f16e6cea56d89f9ceb515bcaebf48ea4c714cdd5871be2247833820000000000000003031a28a58fd4623df84a4d194687b37e8563648ec6623ef51d04f1361466a201db02be22010c5e76935699932e778930a6847f5c2398556fa806dd637ea15df87f16",
					),
					Signature: HexDecode(
						"028946dc95a92a9ae9bfa9fb4abe5eb23f86720644ddfcd143e0ca0ab835b4cacc3938c21de9e557c20ab022b1044dc96d47557adeec1a0cb4284a299001440b31",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 1,
							PrivD: HexDecode(
								"5ed8daa705da9ed1a5f12c9084deabbca692ec0a9459fae157b4f545c094a78a",
							),
							PrivE: HexDecode(
								"11a7ea78f9166e7afb0ae1bb75b04c4aa2f5642bc490ee24e400ca2676f3494d",
							),
							BindingFactor: HexDecode(
								"6155fcebddeec3e6db955315679a2966412dc2d515493d2c1529f7535cfd39a5",
							),
							PrivNonce: HexDecode(
								"09ec611c75bd756ac41afc29be4c5aebebbb66f19139ab8790bc270280d0215c",
							),
							Lagrange: HexDecode(
								"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a2",
							),
							Signature: HexDecode(
								"03cbb6843074eb285bc427bf59bd6e3f8547fe1745b230240e5a9f952f4a8b517c62de0896dac6af63ee6195b3ff068590a9e7fe1e6f1cd81f90ae4f076e28cba5",
							),
						},
						{
							ID: 3,
							PrivD: HexDecode(
								"788b24869cca1b2306f3b69828b534af482f7d94efb1813106a52b1b80c3f0b1",
							),
							PrivE: HexDecode(
								"dcbdd60e0395bd2b37e4ae5a12e5603c5aaba371272d9c279bb5173285ba9eb4",
							),
							BindingFactor: HexDecode(
								"4e5d398a66f805db95a05f7cd04ce7593686ef00345b8e2e30c4c6946b003404",
							),
							PrivNonce: HexDecode(
								"e9d6f400a9336b6506ea2e043322da6c9c4690d2d62d36b391938ddbaa2fd170",
							),
							Lagrange: HexDecode(
								"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0",
							),
							Signature: HexDecode(
								"03f6178ceeff18890a65d4902d98e9f90d7ed71b337dc40aa4934e16a64568d1e0d65ab9870f1ea85e1c4e8cfd054743db581c59a72c45d4d0576e3915635180cd",
							),
						},
					},
				},
				{
					ID:       2,
					Data:     []byte("data"),
					PubNonce: HexDecode("02b00fe5cb30cd33e5cb528daabca5dc1db94a38e1624662fb34093d28a342489b"),
					Commitment: HexDecode(
						"0000000000000003036cab77f534e9e9be21f7364a8ad66fee0cf46f8869ae90dd41ad75c66fccc9f702e5ff600219007412fcd423bd3db7f8717962008f2e4b042c197b534aea3bd6e90000000000000001022c641f8b59cbd4a5eac0f665308fde555bd78b4dac33b4d460415e990e057f0b03cbe778eada3a5d4a9f3c1566fe9acfdf13db4647e087c43c6e22b1bfbec8215c",
					),
					Signature: HexDecode(
						"02b00fe5cb30cd33e5cb528daabca5dc1db94a38e1624662fb34093d28a342489b2494d5654ff0cbe519152fbed05ec4d7a04fbd922e326c7b813011d246824b71",
					),
					AssignedMembers: []AssignedMember{
						{
							ID: 3,
							PrivD: HexDecode(
								"5d73452bd77e0c62d1adab2370f923375940b9a010927be0f24dfc319614ea13",
							),
							PrivE: HexDecode(
								"98d4c5bd7b2a3516625de6dd7f3f30526da618f9a6b52fc1d33d78cc6198b1df",
							),
							BindingFactor: HexDecode(
								"b26874d7078b899f8752847740d9d8953dc7648736d08767afb54716d1b85517",
							),
							PrivNonce: HexDecode(
								"89a14be653d83f1f6db7010e5d77dea249ae0ac9554fcdeadf9ac0af9a79c966",
							),
							Lagrange: HexDecode(
								"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a0",
							),
							Signature: HexDecode(
								"03f127ab0c221a1817b4284395bf874efac62132ce1991a18682e0969b8e2082ce160906aa02a8d0c519a8f6aba96096c815fd22b4f27a234d61ac5385f8e08449",
							),
						},
						{
							ID: 1,
							PrivD: HexDecode(
								"ef21bf202f3a812dc2a690eb5688bc474978d0cba420d4285c1ed32bb122f808",
							),
							PrivE: HexDecode(
								"6f2ef21a505970dce5fbc7a521d1eeda3daf03dd1a8d6534f4c995607de28763",
							),
							BindingFactor: HexDecode(
								"9a15063ec39f11b021db4f17fbe011a276b80a39859e47f282838a15b8c1d500",
							),
							PrivNonce: HexDecode(
								"cdc2207eb8e77cced34ce31c4f56534b37597a4a0da2791156bf7711a34c3fe6",
							),
							Lagrange: HexDecode(
								"7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a2",
							),
							Signature: HexDecode(
								"028d31759f85d662a43d6b757db8c1320a75352139d7b8ace0cfdf57935c141b870e8bcebb4d47fb1fff6c391326fe2e0f8a529add3bb8492e1f83be4c4da1c728",
							),
						},
					},
				},
			},
		},
	}
)
