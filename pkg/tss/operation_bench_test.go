package tss_test

import (
	"math/rand"
	"testing"

	"github.com/bandprotocol/chain/v2/pkg/tss"
)

func BenchmarkSumScalars(b *testing.B) {
	tests := []struct {
		name         string
		numOfScalars int
	}{
		{
			name:         "1 scalar",
			numOfScalars: 1,
		},
		{
			name:         "2 scalars",
			numOfScalars: 2,
		},
		{
			name:         "4 scalars",
			numOfScalars: 4,
		},
		{
			name:         "8 scalars",
			numOfScalars: 8,
		},
		{
			name:         "16 scalars",
			numOfScalars: 16,
		},
		{
			name:         "20 scalars",
			numOfScalars: 20,
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			rand.Seed(0)

			var scalars tss.Scalars
			for i := 0; i < test.numOfScalars; i++ {
				scalar := make([]byte, 32)
				rand.Read(scalar)
				scalars = append(scalars, scalar)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tss.SumScalars(scalars...)
			}
		})
	}
}

func BenchmarkSumPoints(b *testing.B) {
	tests := []struct {
		name        string
		numOfPoints int
	}{
		{
			name:        "1 point",
			numOfPoints: 1,
		},
		{
			name:        "2 points",
			numOfPoints: 2,
		},
		{
			name:        "4 points",
			numOfPoints: 4,
		},
		{
			name:        "8 points",
			numOfPoints: 8,
		},
		{
			name:        "16 points",
			numOfPoints: 16,
		},
		{
			name:        "20 points",
			numOfPoints: 20,
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			rand.Seed(0)

			var points tss.Points
			for i := 0; i < test.numOfPoints; i++ {
				_, point, _ := tss.GenerateDKGNonce()
				points = append(points, point)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tss.SumPoints(points...)
			}
		})
	}
}

func BenchmarkSolveScalarPolynomial(b *testing.B) {
	tests := []struct {
		name              string
		numOfCoefficients int
	}{
		{
			name:              "1 coefficient",
			numOfCoefficients: 1,
		},
		{
			name:              "2 coefficients",
			numOfCoefficients: 2,
		},
		{
			name:              "4 coefficients",
			numOfCoefficients: 4,
		},
		{
			name:              "8 coefficients",
			numOfCoefficients: 8,
		},
		{
			name:              "16 coefficients",
			numOfCoefficients: 16,
		},
		{
			name:              "20 coefficients",
			numOfCoefficients: 20,
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			rand.Seed(0)
			x := make([]byte, 32)
			rand.Read(x)

			var coeffs tss.Scalars
			for i := 0; i < test.numOfCoefficients; i++ {
				coeff := make([]byte, 32)
				rand.Read(coeff)
				coeffs = append(coeffs, coeff)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tss.SolveScalarPolynomial(coeffs, x)
			}
		})
	}
}

func BenchmarkSolvePointPolynomial(b *testing.B) {
	tests := []struct {
		name              string
		numOfCoefficients int
	}{
		{
			name:              "1 coefficient",
			numOfCoefficients: 1,
		},
		{
			name:              "2 coefficients",
			numOfCoefficients: 2,
		},
		{
			name:              "4 coefficients",
			numOfCoefficients: 4,
		},
		{
			name:              "8 coefficients",
			numOfCoefficients: 8,
		},
		{
			name:              "16 coefficients",
			numOfCoefficients: 16,
		},
		{
			name:              "20 coefficients",
			numOfCoefficients: 20,
		},
	}

	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			rand.Seed(0)
			x := make([]byte, 32)
			rand.Read(x)

			var coeffs tss.Points
			for i := 0; i < test.numOfCoefficients; i++ {
				_, point, _ := tss.GenerateDKGNonce()
				coeffs = append(coeffs, point)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				tss.SolvePointPolynomial(coeffs, x)
			}
		})
	}
}
