package lagrange

import (
	"math/big"
)

var N, _ = new(big.Int).SetString("115792089237316195423570985008687907852837564279074904382605163141518161494337", 10)
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

var TTT = [...][]int64{
	2:  {1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536, 131072, 262144},
	3:  {1, 3, 9, 27, 81, 243, 729, 2187, 6561},
	5:  {1, 5, 25, 125, 625},
	7:  {1, 7, 49},
	11: {1, 11},
	13: {1, 13},
	17: {1, 17},
	19: {1, 19},
}

func ComputeCoefficient(i int64, s []int64) *big.Int {
	numerator := big.NewInt(1)
	denominator := big.NewInt(1)
	for _, j := range s {
		if j != i {
			numerator.Mul(big.NewInt(int64(j)), numerator)

			j_i := j - i
			denominator.Mul(big.NewInt(int64(j_i)), denominator)
		}
	}

	result := new(big.Int).Mul(numerator, denominator.ModInverse(denominator, N))
	return result.Mod(result, N)
}

// TODO-TSS: Need to fix on some case (e.g. i = 1, s = [1,2] --> the result should not be 1)
// func ComputeCoefficient2(i int64, s []int64) *big.Int {
// 	counts := make([]int64, 20)
// 	for _, j := range s {
// 		if j != i {
// 			for _, v := range PRIME_FACTORS[j] {
// 				counts[v[0]] += v[1]
// 			}

// 			j_i := j - i
// 			if j_i < 0 {
// 				j_i = -j_i
// 			}
// 			for _, v := range PRIME_FACTORS[j_i] {
// 				counts[v[0]] -= v[1]
// 			}
// 		}
// 	}

// 	numerator := int64(1)
// 	denominator := int64(1)
// 	for k, v := range counts {
// 		if v > 0 {
// 			numerator *= TTT[k][v]
// 		} else if v < 0 {
// 			denominator *= TTT[k][-v]
// 		}
// 	}

// 	numeratorBig := big.NewInt(numerator)
// 	denominatorBig := big.NewInt(denominator)
// 	result := new(big.Int).Mul(numeratorBig, denominatorBig.ModInverse(denominatorBig, N))
// 	return result.Mod(result, N)
// }
