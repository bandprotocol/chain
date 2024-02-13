package tss

import (
	"errors"
	"math/big"
	"math/bits"

	"github.com/ethereum/go-ethereum/crypto"
)

var ContextStringConst = "TSSLib-secp256k1-SHA256-v0"

// I2OSP - Integer-to-Octet-String primitive converts a nonnegative integer to
// an octet string of a specified length `len(buf)`, and stores it in `buf`.
// Reference: https://datatracker.ietf.org/doc/html/rfc8017#section-4.1
func I2OSP(x, xLen int) ([]byte, error) {
	if x < 0 || xLen < 0 {
		return nil, errors.New("I2OSP: x and xLen must be non-negative integer to be converted")
	}

	if bits.Len(uint(x)) > xLen*8 {
		return nil, errors.New("I2OSP: integer too large")
	}

	buf := make([]byte, xLen)
	for i := xLen - 1; i >= 0; i-- {
		buf[i] = byte(x & 0xff)
		x >>= 8
	}

	return buf, nil
}

// OS2IP - Octet-String-to-Integer primitive converts an octet string to a
// nonnegative integer.
// Reference: https://datatracker.ietf.org/doc/html/rfc8017#section-4.2
func OS2IP(buf []byte) *big.Int {
	// SetBytes interprets buf as the bytes of a big-endian unsigned integer.
	return new(big.Int).SetBytes(buf)
}

// strxor performs a bitwise XOR operation on two byte slices of equal length.
// It returns a new byte slice containing the result of the XOR operation.
//
// The function takes two arguments:
// - str1: The first byte slice.
// - str2: The second byte slice.
//
// Both byte slices must be of the same length. If they are not, the function
// returns an error.
func strxor(str1, str2 []byte) ([]byte, error) {
	if len(str1) != len(str2) {
		return nil, errors.New("strxor: lengths of input strings do not match")
	}
	xor := make([]byte, len(str1))
	for i := 0; i < len(str1); i++ {
		xor[i] = str1[i] ^ str2[i]
	}
	return xor, nil
}

// ExpandMessageXMD generates a uniformly random byte string using a cryptographic hash function H that outputs b bits.
//
// Requirements:
//   - H should output b bits where b >= 2 * k (k is the target security level in bits) and b is divisible by 8.
//     This ensures k-bit collision resistance and uniformity of output.
//   - H could be a Merkle-Damgaard hash function like SHA-2, a sponge-based hash function like SHA-3 or BLAKE2,
//     or any hash function that is indifferentiable from a random oracle.
//   - Recommended choices for H are SHA-2 and SHA-3. For a 128-bit security level, b >= 256 bits and either SHA-256
//     or SHA3-256 would be appropriate.
//   - H should ingest fixed-length blocks of data. The length in bits of these blocks is the input block size (s).
//     H requires b <= s for correctness.
//
// Parameters:
// - H: A hash function.
// - msg: Input byte string.
// - DST: Domain Separation Tag, a byte string of at most 255 bytes.
// - lenInBytes: Length of the requested output in bytes, not greater than the lesser of (255 * b_in_bytes) or 2^16-1.
//
// This function returns a byte string of length lenInBytes that is uniformly random and independent of msg,
// provided H meets the requirements above.
//
// https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-hash-to-curve-16#name-expand_message_xmd
func ExpandMessageXMD(h func(data ...[]byte) []byte, msg []byte, dst []byte, len_in_bytes int) ([]byte, error) {
	// b_in_bytes, b / 8 for b the output size of H in bits. For example, for b = 256, b_in_bytes = 32
	b_in_bytes := 32
	// s_in_bytes, the input block size of H, measured in bytes. For example, for SHA-256, s_in_bytes = 64
	s_in_bytes := 64

	// 1. ell = ceil(len_in_bytes / b_in_bytes)
	ell := (len_in_bytes + b_in_bytes - 1) / b_in_bytes

	// 2. ABORT if ell > 255 or len_in_bytes > 65535 or len(DST) > 255
	if ell > 255 || len_in_bytes > 65535 || len(dst) > 255 {
		return nil, errors.New("ExpandMessageXMD: input is not within the permissible limits")
	}

	// 3. DST_prime = DST || I2OSP(len(DST), 1)
	dstPrime, err := I2OSP(len(dst), 1)
	if err != nil {
		return nil, err
	}
	dstPrime = append(dst, dstPrime...)

	// 4. Z_pad = I2OSP(0, s_in_bytes)
	zPad, err := I2OSP(0, s_in_bytes)
	if err != nil {
		return nil, err
	}

	// 5. l_i_b_str = I2OSP(len_in_bytes, 2)
	liBStr, err := I2OSP(len_in_bytes, 2)
	if err != nil {
		return nil, err
	}

	// 6. msg_prime = Z_pad || msg || l_i_b_str || I2OSP(0, 1) || DST_prime
	// I2OSP(0, 1) -> []byte{0}
	msgPrime := append(append(append(zPad, msg...), liBStr...), append([]byte{0}, dstPrime...)...)

	// 7. b_0 = H(msg_prime)
	b0 := h(msgPrime)
	// 8. b_1 = H(b_0 || I2OSP(1, 1) || DST_prime)
	// I2OSP(1, 1) -> []byte{1}
	b := h(append(b0, append([]byte{1}, dstPrime...)...))

	if len(b) < b_in_bytes {
		return nil, errors.New("ExpandMessageXMD: the initial len of b must be >= b_in_bytes")
	}

	// 9. for i in (2, ...,ell):
	for i := 2; i <= ell; i++ {
		// 10. b_i = H(strxor(b_0, b_(i - 1)) || I2OSP(i, 1) || DST_prime)
		b0_xor_bi_1, err := strxor(b0, b[len(b)-b_in_bytes:])
		if err != nil {
			return nil, err
		}
		// I2OSP(i, 1) -> []byte{i} ; i ∈ {1,2,3,...,255}
		bi := h(append(b0_xor_bi_1, append([]byte{uint8(i)}, dstPrime...)...))

		// 11. uniform_bytes = b_1 || ... || b_ell
		b = append(b, bi...)
	}

	// 12. return substr(uniform_bytes, 0, len_in_bytes)
	return b[:len_in_bytes], nil
}

// HashToField is a function that hashes an input message into a set of field elements.
// It is designed to be efficient for certain extension fields, specifically fields of the form GF(p^m).
//
// Parameters:
// - msg: The input message to be hashed.
// - count: The number of field elements to output.
// - p: The characteristic of the finite field F.
// - m: The extension degree of the finite field F.
// - L: A parameter defined as ceil((ceil(log2(p)) + k) / 8), where k is the security parameter of the suite.
// - expand: A function that expands a byte string into a uniformly random byte string.
//
// The function generates field elements that are uniformly random except with bias at most 2^-k,
// where k is the security parameter. It does not use rejection sampling and is designed to be
// amenable to straight line implementations.
//
// The function may fail (abort) if the expand function fails.
//
// Returns:
// - A slice of field elements, each represented as a *big.Int.
// - An error if the expand function returns an error.
//
// https://datatracker.ietf.org/doc/html/draft-irtf-cfrg-hash-to-curve-16#section-5.1
func HashToField(
	msg []byte,
	count int,
	p *big.Int,
	m int,
	l int,
	expand func([]byte, int) ([]byte, error),
) ([][]*big.Int, error) {
	// 1. len_in_bytes = count * m * l
	lenInBytes := count * m * l
	// 2. uniform_bytes = expand_message(msg, DST, len_in_bytes)
	uniformBytes, err := expand(msg, lenInBytes)
	if err != nil {
		return nil, err
	}
	uVals := make([][]*big.Int, count)
	// 3. for i in (0, ..., count - 1):
	for i := 0; i < count; i++ {
		eVals := make([]*big.Int, m)
		// 4. for j in (0, ..., m - 1):
		for j := 0; j < m; j++ {
			// 5. elm_offset = l * (j + i * m)
			elmOffset := l * (j + i*m)
			// 6. tv = substr(uniform_bytes, elm_offset, l)
			tv := uniformBytes[elmOffset : elmOffset+l]
			tvInt := OS2IP(tv)
			// 7. e_j = OS2IP(tv) mod p
			eVals[j] = tvInt.Mod(tvInt, p)
		}
		// 8. u_i = (e_0, ..., e_(m - 1))
		uVals[i] = eVals
	}
	// 9. return (u_0, ..., u_(count - 1))
	return uVals, nil
}

// H_M1_L48 is a helper function that hashes an input message into a set of field elements.
// It uses a hash function H and an expand function defined by the ExpandMessageXMD method.
//
// Parameters:
//   - H: A function that takes a variable number of byte slices and returns a byte slice.
//     This function is used as the hash function in the ExpandMessageXMD method.
//   - count: The number of field elements to produce.
//   - p: The prime number defining the finite field.
//   - msg: The input message to be hashed.
//   - contextString: A domain separation string for the ExpandMessageXMD method.
//
// The function first defines an expand function that uses the provided hash function H and
// the domain separation string. It then calls the HashToField function with this expand function
// and the other parameters to produce the field elements.
//
// Returns:
// - A slice of field elements, each represented as a *big.Int.
// - An error if the HashToField function returns an error.
func H_M1_L48(
	h func(data ...[]byte) []byte,
	count int,
	p *big.Int,
	msg []byte,
	contextString string,
) ([][]*big.Int, error) {
	expand := func(message []byte, lenInBytes int) ([]byte, error) {
		DST := []byte(contextString)
		return ExpandMessageXMD(h, message, DST, lenInBytes)
	}
	// m = 1, L = 48
	fieldElements, err := HashToField(msg, count, p, 1, 48, expand)
	if err != nil {
		return nil, err
	}
	return fieldElements, nil
}

// Implemented as hash_to_field(m, 1) using expand_message_xmd with SHA-256 with parameters
// DST = contextString || "rho", F set to the scalar field, p set to G.Order(), m = 1, and L = 48.
func H1(msg []byte) ([]byte, error) {
	result, err := H_M1_L48(Hash, 1, crypto.S256().Params().P, msg, ContextStringConst+"rho")
	if err != nil {
		return nil, err
	}
	if len(result) == 0 || len(result[0]) == 0 {
		return nil, errors.New("H1: got an empty result from HashToField")
	}
	return result[0][0].Bytes(), err
}

// H2(m): Implemented as hash_to_field(m, 1) using expand_message_xmd with SHA-256 with parameters
// DST = contextString || "chal", F set to the scalar field, p set to G.Order(), m = 1, and L = 48.
func H2(msg []byte) ([]byte, error) {
	result, err := H_M1_L48(Hash, 1, crypto.S256().Params().P, msg, ContextStringConst+"chal")
	if err != nil {
		return nil, err
	}
	if len(result) == 0 || len(result[0]) == 0 {
		return nil, errors.New("H2: got an empty result from HashToField")
	}
	return result[0][0].Bytes(), err
}

// H3(m): Implemented as hash_to_field(m, 1) using expand_message_xmd with SHA-256 with parameters
// DST = contextString || "nonce", F set to the scalar field, p set to G.Order(), m = 1, and L = 48.
func H3(msg []byte) ([]byte, error) {
	result, err := H_M1_L48(Hash, 1, crypto.S256().Params().P, msg, ContextStringConst+"nonce")
	if err != nil {
		return nil, err
	}
	if len(result) == 0 || len(result[0]) == 0 {
		return nil, errors.New("H3: got an empty result from HashToField")
	}
	return result[0][0].Bytes(), err
}

// H4(m): Implemented by computing H(contextString || "msg" || m).
func H4(msg []byte) []byte {
	return Hash(append([]byte(ContextStringConst+"msg"), msg...))
}

// H5(m): Implemented by computing H(contextString || "com" || m).
func H5(msg []byte) []byte {
	return Hash(append([]byte(ContextStringConst+"com"), msg...))
}
