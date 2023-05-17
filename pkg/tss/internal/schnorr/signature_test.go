// Copyright (c) 2015-2020 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package schnorr

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// TestSignatureParsing ensures that signatures are properly parsed including
// error paths.
func TestSignatureParsing(t *testing.T) {
	tests := []struct {
		name string // test description
		sig  string // hex encoded signature to parse
		err  error  // expected error
	}{{
		name: "valid signature 1",
		sig: "c6ec70969d8367538c442f8e13eb20ff0c9143690f31cd3a384da54dd29ec0aa" +
			"4b78a1b0d6b4186195d42a85614d3befd9f12ed26542d0dd1045f38c98b4a405",
		err: nil,
	}, {
		name: "valid signature 2",
		sig: "adc21db084fa1765f9372c2021fb298720f3d13e6d844e2dff751a2d46a69277" +
			"0b989e316f7faf308a5f4a7343c0569465287cf6bff457250d6dacbb361f6e63",
		err: nil,
	}, {
		name: "empty",
		sig:  "",
		err:  ErrSigTooShort,
	}, {
		name: "too short by one byte",
		sig: "adc21db084fa1765f9372c2021fb298720f3d13e6d844e2dff751a2d46a69277" +
			"0b989e316f7faf308a5f4a7343c0569465287cf6bff457250d6dacbb361f6e",
		err: ErrSigTooShort,
	}, {
		name: "too long by one byte",
		sig: "adc21db084fa1765f9372c2021fb298720f3d13e6d844e2dff751a2d46a69277" +
			"0b989e316f7faf308a5f4a7343c0569465287cf6bff457250d6dacbb361f6e6300",
		err: ErrSigTooLong,
	}, {
		name: "r == p",
		sig: "fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigRTooBig,
	}, {
		name: "r > p",
		sig: "fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30" +
			"181522ec8eca07de4860a4acdd12909d831cc56cbbac4622082221a8768d1d09",
		err: ErrSigRTooBig,
	}, {
		name: "s == n",
		sig: "4e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141",
		err: ErrSigSTooBig,
	}, {
		name: "s > n",
		sig: "4e45e16932b8af514961a1d3a1a25fdf3f4f7732e9d624c6c61548ab5fb8cd41" +
			"fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364142",
		err: ErrSigSTooBig,
	}}

	for _, test := range tests {
		_, err := ParseSignature(hexToBytes(test.sig))
		if !errors.Is(err, test.err) {
			t.Errorf("%s mismatched err -- got %v, want %v", test.name, err,
				test.err)
			continue
		}
	}
}

// TestSchnorrSignAndVerify ensures the Schnorr signing function produces the
// expected signatures for a selected set of private keys, messages, and nonces
// that have been independently verified with the Sage computer algebra system.
// It also ensures verifying the signature works as expected.
func TestSchnorrSignAndVerify(t *testing.T) {
	tests := []struct {
		name     string // test description
		key      string // hex encded private key
		msg      string // hex encoded message to sign before hashing
		hash     string // hex encoded hash of the message to sign
		nonce    string // hex encoded nonce to use in the signature calculation
		rfc6979  bool   // whether or not the nonce is an RFC6979 nonce
		expected string // expected signature
	}{{
		name:    "key 0x1, blake256(0x01020304), rfc6979 nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000001",
		hash:    "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		nonce:   "d4e18f08eb87073cb2a6707def02007315f7349c3c132590a0088fefece557ef",
		rfc6979: true,
		expected: "4c68976afe187ff0167919ad181cb30f187e2af1c8233b2cbebbbe0fc97fff61" +
			"8b5953f24beb14c6df56ceb6ec9ea5465c020a5a963579614766d9943cfca3bb",
	}, {
		name:    "key 0x1, blake256(0x01020304), random nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000001",
		hash:    "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		nonce:   "a6df66500afeb7711d4c8e2220960855d940a5ed57260d2c98fbf6066cca283e",
		rfc6979: false,
		expected: "b073759a96a835b09b79e7b93c37fdbe48fb82b000c4a0e1404ba5d1fbc15d0a" +
			"4aaa522666c0591eefc11841a2437bd7200a331e914d59e9b246abf43267f981",
	}, {
		name:    "key 0x2, blake256(0x01020304), rfc6979 nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000002",
		hash:    "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		nonce:   "341682d3064ec802646be9c4a0fd97f8480807fcac3179e97098b8597de909dc",
		rfc6979: true,
		expected: "c6deb3a26c08842612bfd4411a91c90f64cfea2206c758cd1352ff2b93cc3611" +
			"08c859f09cf0872710fca9d39247f930c13e8879f863a202f634d9fa7f98920d",
	}, {
		name:    "key 0x2, blake256(0x01020304), random nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000002",
		hash:    "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		nonce:   "679a6d36e7fe6c02d7668af86d78186e8f9ccc04371ac1c8c37939d1f5cae07a",
		rfc6979: false,
		expected: "4a090d82f48ca12d9e7aa24b5dcc187ee0db2920496f671d63e86036aaa7997e" +
			"900a02199b8d19f759caa01200d60e4fe6d729cbd4b73ae44c22cfe494eadf94",
	}, {
		name:    "key 0x1, blake256(0x0102030405), rfc6979 nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000001",
		hash:    "dc063eba3c8d52a159e725c1a161506f6cb6b53478ad5ef3f08d534efa871d9f",
		nonce:   "cfbabebb15824ff3cfa5f4080a8608aaa9db891541851b27275c61db9d6d7e1c",
		rfc6979: true,
		expected: "461646005002d673c2e903f3c9ff2c2455e60810445ee486b9c36152287bc41a" +
			"bcd2beaf85920a5489c3f78e143dd620d5f0f2ab47f58b5e3b526a6b2cb60719",
	}, {
		name:    "key 0x1, blake256(0x0102030405), random nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000001",
		hash:    "dc063eba3c8d52a159e725c1a161506f6cb6b53478ad5ef3f08d534efa871d9f",
		nonce:   "65f880c892fdb6e7f74f76b18c7c942cfd037ef9cf97c39c36e08bbc36b41616",
		rfc6979: false,
		expected: "72e5666f4e9d1099447b825cf737ee32112f17a67e2ca7017ae098da31dfbb8b" +
			"d899033f78240687fd5706452a61c26b95995b027e77cf873326d461dcb26485",
	}, {
		name:    "key 0x2, blake256(0x0102030405), rfc6979 nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000002",
		hash:    "dc063eba3c8d52a159e725c1a161506f6cb6b53478ad5ef3f08d534efa871d9f",
		nonce:   "f7a8f640df67ba21b619eb742a73cbfc58739153b8772d5b2f8781f33d45e554",
		rfc6979: true,
		expected: "f3632492a72eb8e175b93e1eb31ef382e49f3f3fe385892523beaef9171aa15d" +
			"97fc06c0b1adbcba20141f1fc34be2e03f0150c318d3df14b9f0f39b99186037",
	}, {
		name:    "key 0x2, blake256(0x0102030405), random nonce",
		key:     "0000000000000000000000000000000000000000000000000000000000000002",
		hash:    "dc063eba3c8d52a159e725c1a161506f6cb6b53478ad5ef3f08d534efa871d9f",
		nonce:   "026ece4cfb704733dd5eef7898e44c33bd5a0d749eb043f48705e40fa9e9afa0",
		rfc6979: false,
		expected: "3c4c5a2f217ea758113fd4e89eb756314dfad101a300f48e5bd764d3b6e0f8bf" +
			"f8726db3a75bff8d6b2ae758f0998c518d01a486566fb0e37dfa5e4e74dfb672",
	}, {
		name:    "random key 1, blake256(0x01), rfc6979 nonce",
		key:     "a1becef2069444a9dc6331c3247e113c3ee142edda683db8643f9cb0af7cbe33",
		hash:    "4a6c419a1e25c85327115c4ace586decddfe2990ed8f3d4d801871158338501d",
		nonce:   "c23097718bd90c10ba2e99abff92f21c0eec71796712a772f0ce10f2b1bc6f5f",
		rfc6979: true,
		expected: "0b89d1fb10635e4a5da463c7339fd0f8d2e7d205a8288d4f973635beb8b59f7f" +
			"d8adb0cae6bf802d18ec3be5c94a2c67a078cce2bc2cf59efcda5fcc12c074ce",
	}, {
		name:    "random key 2, blake256(0x02), rfc6979 nonce",
		key:     "59930b76d4b15767ec0e8c8e5812aa2e57db30c6af7963e2a6295ba02af5416b",
		hash:    "49af37ab5270015fe25276ea5a3bb159d852943df23919522a202205fb7d175c",
		nonce:   "342d8326464a0b5866091126e2aa29a960eba8e47dba7bef355b18b3f9011793",
		rfc6979: true,
		expected: "533e99ee9c838af4cc0280b0223ab0560e7e2083694bd5b0cab3c0cb80bc2e1e" +
			"0ceb118a9b82bc8d00ab939eaa678c3c7012a482a12706c8c202f5dfe97b6d69",
	}, {
		name:    "random key 3, blake256(0x03), rfc6979 nonce",
		key:     "c5b205c36bb7497d242e96ec19a2a4f086d8daa919135cf490d2b7c0230f0e91",
		hash:    "b706d561742ad3671703c247eb927ee8a386369c79644131cdeb2c5c26bf6c5d",
		nonce:   "710a4f1a3bee3567b53bd4dd0c9c0e55d76981a5ed488223ca0583bf8a563951",
		rfc6979: true,
		expected: "95c966fd6435d505a492548370b29a3c40efc3fefa3e1d997b3e2788cc33836e" +
			"b2f47b9c90c3525471bdfbbd39b7a9b388e9727249d12497e48f9ca9da216c77",
	}, {
		name:    "random key 4, blake256(0x04), rfc6979 nonce",
		key:     "65b46d4eb001c649a86309286aaf94b18386effe62c2e1586d9b1898ccf0099b",
		hash:    "4c6eb9e38415034f4c93d3304d10bef38bf0ad420eefd0f72f940f11c5857786",
		nonce:   "cb4727000027551b8c2c3b717696dcff46f9ad088050571cb8634038003fc136",
		rfc6979: true,
		expected: "327f4e1dc74948df95dba34f26b63317568325316742fc8276be8cd2544a105c" +
			"f6e2a7533084640b39e5634f7366b1160b178ac10be5bb766bfbb499baacd94e",
	}, {
		name:    "random key 5, blake256(0x05), rfc6979 nonce",
		key:     "915cb9ba4675de06a182088b182abcf79fa8ac989328212c6b866fa3ec2338f9",
		hash:    "bdd15db13448905791a70b68137445e607cca06cc71c7a58b9b2e84a06c54d08",
		nonce:   "665a2ba74200aaee038de3248c1acb8d92ca9c0a89ff63d140755834e04d55e8",
		rfc6979: true,
		expected: "b3ac51091150852794914e12f12b8db00ec517ca8eeca0175a20e62b1a413a5c" +
			"bb152fd33996383e366c42801346adcb1842fdec5c4b2aef4cde035c9103fa35",
	}, {
		name:    "random key 6, blake256(0x06), rfc6979 nonce",
		key:     "93e9d81d818f08ba1f850c6dfb82256b035b42f7d43c1fe090804fb009aca441",
		hash:    "19b7506ad9c189a9f8b063d2aee15953d335f5c88480f8515d7d848e7771c4ae",
		nonce:   "b817c907f71b11359bc2857e39f0f13d3a2cbaaadb722665ea73d7edf38c4342",
		rfc6979: true,
		expected: "01bfb35cf41d809d572d1d891eb474e2c0decf67ebb0f1432edce06b75d73fe0" +
			"133cde88c8894fd89560be9a75550aee0c6c738ee56f0df57e7cfa632c54f121",
	}, {
		name:    "random key 7, blake256(0x07), rfc6979 nonce",
		key:     "c249bbd5f533672b7dcd514eb1256854783531c2b85fe60bf4ce6ea1f26afc2b",
		hash:    "53d661e71e47a0a7e416591200175122d83f8af31be6a70af7417ad6f54d0038",
		nonce:   "7eaa64ba668b3c77b0586695645707236f165a76ed7a53a04c833048995f8bc7",
		rfc6979: true,
		expected: "cb5bd3805bdd0a2e4daf58b30aa26b48c81ca59421ca320ad983c1eef672ad52" +
			"1e5b3757e8e70a52b45b25c062d8f54236a6d8f58f843f59464be7358e2fd4bf",
	}, {
		name:    "random key 8, blake256(0x08), rfc6979 nonce",
		key:     "ec0be92fcec66cf1f97b5c39f83dfd4ddcad0dad468d3685b5eec556c6290bcc",
		hash:    "9bff7982eab6f7883322edf7bdc86a23c87ca1c07906fbb1584f57b197dc6253",
		nonce:   "63e12aa7d19a413577fbf6a0896f13040befb5b675f9238a09b9db400d9f454a",
		rfc6979: true,
		expected: "9fbd427ddaef7c7ab87e5555c1faca398695e423ce44e5fc648b9203e38b69a0" +
			"3ceadfccad30ad26a3b8d071d3119ff8edb35ef1a7e0b8be050b1200cf012214",
	}, {
		name:    "random key 9, blake256(0x09), rfc6979 nonce",
		key:     "6847b071a7cba6a85099b26a9c3e57a964e4990620e1e1c346fecc4472c4d834",
		hash:    "4c2231813064f8500edae05b40195416bd543fd3e76c16d6efb10c816d92e8b6",
		nonce:   "95adf9b15f485dc961061053838dbd0fb1fa8663ac344d78f3833acb5fdbfdc6",
		rfc6979: true,
		expected: "cd9e9100f0fc8b631b40c4d93437eaf608e25ab6ad295d8b6460289ce571fb1e" +
			"49a9795aaa368b3aac1dc99e0d3a6b50af9da3b5bbad7444b0ce2b4071122428",
	}, {
		name:    "random key 10, blake256(0x0a), rfc6979 nonce",
		key:     "b7548540f52fe20c161a0d623097f827608c56023f50442cc00cc50ad674f6b5",
		hash:    "e81db4f0d76e02805155441f50c861a8f86374f3ae34c7a3ff4111d3a634ecb1",
		nonce:   "014c6f95c371ba1dd62e759229b65a7ffced18680f34789a204e1044926722ff",
		rfc6979: true,
		expected: "c379f1c2a35b2f9712a5573fb59c4c29dfdc54cef833dc211716248d5c7e28e1" +
			"3b5c55090008c1d5363892841cc79de33fdb4e28d90a65edf3bf6445269524e7",
	}}

	for _, test := range tests {
		privKey := hexToModNScalar(test.key)
		hash := hexToBytes(test.hash)
		nonce := hexToModNScalar(test.nonce)
		wantSig := hexToBytes(test.expected)

		// Ensure the test data is sane by comparing the provided hashed message
		// and nonce, in the case rfc6979 was used, to their calculated values.
		// These values could just be calculated instead of specified in the
		// test data, but it's nice to have all of the calculated values
		// available in the test data for cross implementation testing and
		// verification.
		if test.rfc6979 {
			privKeyBytes := hexToBytes(test.key)
			nonceBytes := hexToBytes(test.nonce)
			calcNonce := secp256k1.NonceRFC6979(privKeyBytes, hash,
				RFC6979ExtraDataV0[:], nil, 0)
			calcNonceBytes := calcNonce.Bytes()
			if !bytes.Equal(calcNonceBytes[:], nonceBytes) {
				t.Errorf("%s: mismatched test nonce -- expected: %x, given: %x",
					test.name, calcNonceBytes, nonceBytes)
				continue
			}
		}

		// Sign the hash of the message with the given private key and nonce.
		gotSig, err := Sign(privKey, nonce, hash)
		if err != nil {
			t.Errorf("%s: unexpected error when signing: %v", test.name, err)
			continue
		}

		// Ensure the generated signature is the expected value.
		gotSigBytes := gotSig.Serialize()
		if !bytes.Equal(gotSigBytes, wantSig) {
			t.Errorf("%s: unexpected signature -- got %x, want %x", test.name,
				gotSigBytes, wantSig)
			continue
		}

		// Ensure the produced signature verifies as well.
		pubKey := secp256k1.NewPrivateKey(hexToModNScalar(test.key)).PubKey()
		err = Verify(gotSig, hash, pubKey, nil, nil)
		if err != nil {
			t.Errorf("%s: signature failed to verify: %v", test.name, err)
			continue
		}
	}
}

// TestSchnorrSignAndVerifyRandom ensures the Schnorr signing and verification
// work as expected for randomly-generated private keys and messages.  It also
// ensures invalid signatures are not improperly verified by mutating the valid
// signature and changing the message the signature covers.
func TestSchnorrSignAndVerifyRandom(t *testing.T) {
	// Use a unique random seed each test instance and log it if the tests fail.
	seed := time.Now().Unix()
	rng := rand.New(rand.NewSource(seed))
	defer func(t *testing.T, seed int64) {
		if t.Failed() {
			t.Logf("random seed: %d", seed)
		}
	}(t, seed)

	for i := 0; i < 50000; i++ {
		// Generate a random private key.
		var buf [32]byte
		if _, err := rng.Read(buf[:]); err != nil {
			t.Fatalf("failed to read random private key: %v", err)
		}
		var privKeyScalar secp256k1.ModNScalar
		privKeyScalar.SetBytes(&buf)
		privKey := secp256k1.NewPrivateKey(&privKeyScalar)

		// Generate a random hash to sign.
		var hash [32]byte
		if _, err := rng.Read(hash[:]); err != nil {
			t.Fatalf("failed to read random hash: %v", err)
		}

		// Sign the hash with the private key and then ensure the produced
		// signature is valid for the hash and public key associated with the
		// private key.
		nonce := secp256k1.NonceRFC6979(privKey.Serialize(), hash[:], RFC6979ExtraDataV0[:], nil, 0)
		sig, err := Sign(&privKey.Key, nonce, hash[:])
		if err != nil {
			t.Fatalf("failed to sign\nprivate key: %x\nhash: %x",
				privKey.Serialize(), hash)
		}
		pubKey := privKey.PubKey()
		if err := Verify(sig, hash[:], pubKey, nil, nil); err != nil {
			t.Fatalf("failed to verify signature\nsig: %x\nhash: %x\n"+
				"private key: %x\npublic key: %x", sig.Serialize(), hash,
				privKey.Serialize(), pubKey.SerializeCompressed())
		}

		// Change a random bit in the signature and ensure the bad signature
		// fails to verify the original message.
		goodSigBytes := sig.Serialize()
		badSigBytes := make([]byte, len(goodSigBytes))
		copy(badSigBytes, goodSigBytes)
		randByte := rng.Intn(len(badSigBytes))
		randBit := rng.Intn(7)
		badSigBytes[randByte] ^= 1 << randBit
		badSig, err := ParseSignature(badSigBytes)
		if err != nil {
			t.Fatalf("failed to create bad signature: %v", err)
		}
		if err := Verify(badSig, hash[:], pubKey, nil, nil); err == nil {
			t.Fatalf("verified bad signature\nsig: %x\nhash: %x\n"+
				"private key: %x\npublic key: %x", badSig.Serialize(), hash,
				privKey.Serialize(), pubKey.SerializeCompressed())
		}

		// Change a random bit in the hash that was originally signed and ensure
		// the original good signature fails to verify the new bad message.
		badHash := make([]byte, len(hash))
		copy(badHash, hash[:])
		randByte = rng.Intn(len(badHash))
		randBit = rng.Intn(7)
		badHash[randByte] ^= 1 << randBit
		if err := Verify(sig, badHash[:], pubKey, nil, nil); err == nil {
			t.Fatalf("verified signature for bad hash\nsig: %x\nhash: %x\n"+
				"pubkey: %x", sig.Serialize(), badHash,
				pubKey.SerializeCompressed())
		}
	}
}

// TestVerifyErrors ensures several error paths in Schnorr verification are
// detected as expected.  When possible, the signatures are otherwise valid with
// the exception of the specific failure to ensure it's robust against things
// like fault attacks.
func TestVerifyErrors(t *testing.T) {
	tests := []struct {
		name string // test description
		sigR string // hex encoded r component of signature to verify against
		sigS string // hex encoded s component of signature to verify against
		hash string // hex encoded hash of message to verify
		pubX string // hex encoded x component of pubkey to verify against
		pubY string //  hex encoded y component of pubkey to verify against
		err  error  // expected error
	}{{
		// Signature created from private key 0x01, blake256(0x01020304) over
		// the secp256r1 curve (note the r1 instead of k1).
		name: "pubkey not on the curve, signature valid for secp256r1 instead",
		sigR: "c6c62660176b3daa90dbf4d7e21d9406ce93895771a16c7c5c91258a9b522174",
		sigS: "f5b5583956a6b30e18ff5e865c77a8c4adf47b147d11ea3822b4de63c9f7b909",
		hash: "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		pubX: "6b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c296",
		pubY: "4fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5",
		err:  ErrPubKeyNotOnCurve,
	}, {
		// Signature invented since finding a signature with an r value that is
		// exactly the field prime prior to the modular reduction is not
		// calculable without breaking the underlying crypto.
		name: "r == field prime",
		sigR: "fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f",
		sigS: "e9ae2d0e306497236d4e328dc1a34244045745e87da69d806859348bc2a74525",
		hash: "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		pubX: "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
		pubY: "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
		err:  ErrSigRTooBig,
	}, {
		// Likewise, signature invented since finding a signature with an r
		// value that would be valid modulo the field prime and is still 32
		// bytes is not calculable without breaking the underlying crypto.
		name: "r > field prime (prime + 1)",
		sigR: "fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc30",
		sigS: "e9ae2d0e306497236d4e328dc1a34244045745e87da69d806859348bc2a74525",
		hash: "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		pubX: "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
		pubY: "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
		err:  ErrSigRTooBig,
	}, {
		// Signature invented since finding a signature with an s value that is
		// exactly the group order prior to the modular reduction is not
		// calculable without breaking the underlying crypto.
		name: "s == group order",
		sigR: "4c68976afe187ff0167919ad181cb30f187e2af1c8233b2cbebbbe0fc97fff61",
		sigS: "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141",
		hash: "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		pubX: "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
		pubY: "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
		err:  ErrSigSTooBig,
	}, {
		// Likewise, signature invented since finding a signature with an s
		// value that would be valid modulo the group order and is still 32
		// bytes is not calculable without breaking the underlying crypto.
		name: "s > group order and still 32 bytes (order + 1)",
		sigR: "4c68976afe187ff0167919ad181cb30f187e2af1c8233b2cbebbbe0fc97fff61",
		sigS: "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364142",
		hash: "c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7",
		pubX: "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
		pubY: "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
		err:  ErrSigSTooBig,
	}, {
		// Signature created from private key 0x01, blake256(0x01020304050607).
		// It is otherwise valid.
		name: "odd R",
		sigR: "2c2c71f7bf3e183238b1f20d856e068dc6d37805c8b2d872d0f23d906bc95789",
		sigS: "eb7670ca6ff95c1d5c6785bc72e0781f27c9778758317d82d3053fdbcc9c17b0",
		hash: "ccf8c53a7631aad469d412963d495c729ff219dd2ae9a0c4de4bd1b4c777d49c",
		pubX: "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798",
		pubY: "483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8",
		err:  ErrSigRYIsOdd,
	}}
	// NOTE: There is no test for e >= group order because it would require
	// finding a preimage that hashes to the value range [n, 2^256) and n is
	// close enough to 2^256 that there is only roughly a 1 in 2^128 chance of
	// a given hash falling in that range.  In other words, it's not feasible
	// to calculate.

	for _, test := range tests {
		// Parse test data into types.
		hash := hexToBytes(test.hash)
		pubX, pubY := hexToFieldVal(test.pubX), hexToFieldVal(test.pubY)
		pubKey := secp256k1.NewPublicKey(pubX, pubY)

		// Create the serialized signature from the bytes and attempt to parse
		// it to ensure the cases where the r and s components exceed the
		// allowed range is caught.
		sig, err := ParseSignature(hexToBytes(test.sigR + test.sigS))
		if err != nil {
			if !errors.Is(err, test.err) {
				t.Errorf("%s: mismatched err -- got %v, want %v", test.name, err,
					test.err)
			}

			continue
		}

		// Ensure the expected error is hit.
		err = Verify(sig, hash, pubKey, nil, nil)
		if !errors.Is(err, test.err) {
			t.Errorf("%s: mismatched err -- got %v, want %v", test.name, err,
				test.err)
			continue
		}
	}
}

// hexToModNScalar converts the passed hex string into a ModNScalar and will
// panic if there is an error.  This is only provided for the hard-coded
// constants so errors in the source code can be detected. It will only (and
// must only) be called with hard-coded values.
func hexToModNScalar(s string) *secp256k1.ModNScalar {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	var scalar secp256k1.ModNScalar
	if overflow := scalar.SetByteSlice(b); overflow {
		panic("hex in source file overflows mod N scalar: " + s)
	}
	return &scalar
}

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// hexToFieldVal converts the passed hex string into a FieldVal and will panic
// if there is an error.  This is only provided for the hard-coded constants so
// errors in the source code can be detected. It will only (and must only) be
// called with hard-coded values.
func hexToFieldVal(s string) *secp256k1.FieldVal {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	var f secp256k1.FieldVal
	if overflow := f.SetByteSlice(b); overflow {
		panic("hex in source file overflows mod P: " + s)
	}
	return &f
}
