package lagrange

import (
	"math/big"
)

// N is the order of the secp256k1 elliptic curve group, represented as a big.Int.
var N, _ = new(big.Int).SetString("115792089237316195423570985008687907852837564279074904382605163141518161494337", 10)

// PRIME_FACTORS contains pre-computed prime factors for the numbers up to 20.
var PRIME_FACTORS = [...][][2]int64{
	2:  {{2, 1}},
	3:  {{3, 1}},
	4:  {{2, 2}},
	5:  {{5, 1}},
	6:  {{2, 1}, {3, 1}},
	7:  {{7, 1}},
	8:  {{2, 3}},
	9:  {{3, 2}},
	10: {{2, 1}, {5, 1}},
	11: {{11, 1}},
	12: {{2, 2}, {3, 1}},
	13: {{13, 1}},
	14: {{2, 1}, {7, 1}},
	15: {{3, 1}, {5, 1}},
	16: {{2, 4}},
	17: {{17, 1}},
	18: {{2, 1}, {3, 2}},
	19: {{19, 1}},
	20: {{2, 2}, {5, 1}},
}

// PRECOMPUTED_POWERS contains pre-computed powers of certain prime numbers.
var PRECOMPUTED_POWERS = [...][]int64{
	2:  {1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144},
	3:  {1, 3, 9, 27, 81, 243, 729, 2187, 6561},
	5:  {1, 5, 25, 125, 625},
	7:  {1, 7, 49},
	11: {1, 11},
	13: {1, 13},
	17: {1, 17},
	19: {1, 19},
}

// ComputeCoefficient calculates the Lagrange coefficient for a given index and set of indices.
// The formula used is 𝚷(j/(j-i)) for all j in S-{i}, where:
// - 𝚷 denotes the product of the following statement
// - S ⊂ {1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
// - i ∈ S
// - j ∈ S-{i}
// The Lagrange coefficient is used in polynomial interpolation in threshold secret sharing schemes.
func ComputeCoefficient(i int64, s []int64) *big.Int {
	// Initialize numerator and denominator to 1
	numerator := big.NewInt(1)
	denominator := big.NewInt(1)

	// Iterate through all elements in S
	for _, j := range s {
		// Skip if j == i
		if j == i {
			continue
		}

		// 𝚷(j)
		numerator.Mul(big.NewInt(int64(j)), numerator)
		// 𝚷(j-i)
		denominator.Mul(big.NewInt(int64(j-i)), denominator)
	}

	// Multiply the numerator by the modular inverse of the denominator.
	// The modular inverse of a number x is a number y such that (x*y) % N = 1, where N is the order of the group.
	// This is equivalent to dividing the numerator by the denominator in modular arithmetic as the following formula.
	// 𝚷(j/(j-i)) = (𝚷(j))/(𝚷(j-i)) = numerator/denominator
	result := new(big.Int).Mul(numerator, denominator.ModInverse(denominator, N))
	return result.Mod(result, N)
}

// ComputeCoefficientPreCompute computes the Lagrange coefficient for a given index i and a set S of indices.
// The function optimizes computations by using pre-computed prime factors and powers of numbers.
// The formula used is 𝚷(j/(j-i)) for all j in S-{i}, where:
// - 𝚷 denotes the product of the following statement
// - S ⊂ {1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20}
// - i ∈ S
// - j ∈ S-{i}
// The Lagrange coefficient is used in polynomial interpolation in threshold secret sharing schemes.
func ComputeCoefficientPreCompute(i int64, s []int64) *big.Int {
	// Counts the power of prime factors in the numerator and denominator of the result.
	counts := make([]int64, 20)

	// Sign of the result (can be negative if the number of negative terms in the product is odd).
	sign := int64(1)

	// Loop through each index j in the set S.
	for _, j := range s {
		// Skip if j == i
		if j == i {
			continue
		}
		// Add the prime factors of j to the numerator.
		for _, v := range PRIME_FACTORS[j] {
			counts[v[0]] += v[1]
		}

		// Subtract the prime factors of (j-i) from the numerator or denominator depending on its sign.
		j_i := j - i
		if j_i < 0 {
			j_i = -j_i
			sign *= -1
		}
		for _, v := range PRIME_FACTORS[j_i] {
			counts[v[0]] -= v[1]
		}
	}

	// Compute the product of the powers of prime factors for the numerator and the denominator.
	numerator := int64(1)
	denominator := int64(1)
	for k, v := range counts {
		if v > 0 {
			numerator *= PRECOMPUTED_POWERS[k][v]
		} else if v < 0 {
			denominator *= PRECOMPUTED_POWERS[k][-v]
		}
	}

	// Multiply the numerator by the modular inverse of the denominator.
	// The modular inverse of a number x is a number y such that (x*y) % N = 1, where N is the order of the group.
	// This is equivalent to dividing the numerator by the denominator in modular arithmetic as the following formula.
	// 𝚷(j/(j-i)) = (𝚷(j))/(𝚷(j-i)) = numerator/denominator
	numeratorBig := big.NewInt(numerator * sign)
	denominatorBig := big.NewInt(denominator)
	result := new(big.Int).Mul(numeratorBig, denominatorBig.ModInverse(denominatorBig, N))
	return result.Mod(result, N)
}
