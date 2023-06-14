package tss

import (
	"errors"
	"math/bits"

	"github.com/ethereum/go-ethereum/crypto"
)

// Encrypt encrypts the given value using the key.
// encrypted value = Hash(key) + value
// It returns the encrypted value as a Scalar.
func Encrypt(value Scalar, key PublicKey) (Scalar, error) {
	k, err := Scalar(Hash(key)).Parse()
	if err != nil {
		return nil, NewError(err, "parse key")
	}

	v, err := value.Parse()
	if err != nil {
		return nil, NewError(err, "parse value")
	}

	res := k.Add(v).Bytes()
	return res[:], nil
}

// Decrypt decrypts the given encrypted value using the key.
// value = encrypted value - Hash(key)
// It returns the decrypted value as a Scalar.
func Decrypt(encValue Scalar, key PublicKey) (Scalar, error) {
	k, err := Scalar(Hash(key)).Parse()
	if err != nil {
		return nil, NewError(err, "parse key")
	}

	ev, err := encValue.Parse()
	if err != nil {
		return nil, NewError(err, "parse encrypted value")
	}

	res := k.Negate().Add(ev).Bytes()

	return res[:], nil
}

// Hash calculates the Keccak-256 hash of the given data.
// It returns the hash value as a byte slice.
func Hash(data ...[]byte) []byte {
	return crypto.Keccak256(data...)
}

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
func OS2IP(X []byte) uint {
	var x uint

	for i := 0; i < len(X); i++ {
		x <<= 8
		x |= uint(X[i])
	}

	return x
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

func ExpandMessageXMD(H func(data ...[]byte) []byte, msg []byte, DST []byte, len_in_bytes int) ([]byte, error) {
	// b_in_bytes, b / 8 for b the output size of H in bits. For example, for b = 256, b_in_bytes = 32
	b_in_bytes := 32
	// s_in_bytes, the input block size of H, measured in bytes. For example, for SHA-256, s_in_bytes = 64
	s_in_bytes := 64

	// 1. ell = ceil(len_in_bytes / b_in_bytes)
	ell := (len_in_bytes + b_in_bytes - 1) / b_in_bytes

	// 2. ABORT if ell > 255 or len_in_bytes > 65535 or len(DST) > 255
	if ell > 255 || len_in_bytes > 65535 || len(DST) > 255 {
		return nil, errors.New("ExpandMessageXMD: input is not within the permissible limits")
	}

	// 3. DST_prime = DST || I2OSP(len(DST), 1)
	dstPrime, err := I2OSP(len(DST), 1)
	if err != nil {
		return nil, err
	}
	dstPrime = append(DST, dstPrime...)

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
	b0 := H(msgPrime)
	// 8. b_1 = H(b_0 || I2OSP(1, 1) || DST_prime)
	// I2OSP(1, 1) -> []byte{1}
	b := H(append(b0, append([]byte{1}, dstPrime...)...))

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
		bi := H(append(b0_xor_bi_1, append([]byte{uint8(i)}, dstPrime...)...))

		// 11. uniform_bytes = b_1 || ... || b_ell
		b = append(b, bi...)
	}

	// 12. return substr(uniform_bytes, 0, len_in_bytes)
	return b[:len_in_bytes], nil
}
