package tss_test

import (
	"crypto/rand"
	"testing"

	"github.com/bandprotocol/chain/v3/pkg/tss"
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
			var scalars tss.Scalars
			for i := 0; i < test.numOfScalars; i++ {
				scalar := make([]byte, 32)
				if _, err := rand.Read(scalar); err != nil {
					b.Fatal(err)
				}
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
			var points tss.Points
			for i := 0; i < test.numOfPoints; i++ {
				_, point, _ := tss.GenerateDKGNonce()
				points = append(points, point)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = tss.SumPoints(points...)
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
			x := make([]byte, 32)
			if _, err := rand.Read(x); err != nil {
				b.Fatal(err)
			}

			var coeffs tss.Scalars
			for i := 0; i < test.numOfCoefficients; i++ {
				coeff := make([]byte, 32)
				if _, err := rand.Read(coeff); err != nil {
					b.Fatal(err)
				}
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
			x := make([]byte, 32)
			if _, err := rand.Read(x); err != nil {
				b.Fatal(err)
			}

			var coeffs tss.Points
			for i := 0; i < test.numOfCoefficients; i++ {
				_, point, _ := tss.GenerateDKGNonce()
				coeffs = append(coeffs, point)
			}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				if _, err := tss.SolvePointPolynomial(coeffs, x); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
