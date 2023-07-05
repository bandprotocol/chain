package tss

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeKeySym computes the key symmetry between a private key and a public key.
// It returns the computed key symmetry as a PublicKey and an error, if any.
func ComputeKeySym(rawPrivKeyI Scalar, rawPubKeyJ Point) (Point, error) {
	privKeyI := rawPrivKeyI.modNScalar()
	pubKeyJ, err := rawPubKeyJ.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse publicKeyJ")
	}

	keySym := new(secp256k1.JacobianPoint)
	secp256k1.ScalarMultNonConst(privKeyI, pubKeyJ, keySym)

	return NewPointFromJacobianPoint(keySym), nil
}

// ComputeNonceSym computes the nonce symmetry between a nonce value and a public key.
// It returns the computed nonce symmetry as a PublicKey and an error, if any.
func ComputeNonceSym(rawNonce Scalar, rawPubKeyJ Point) (Point, error) {
	nonce := rawNonce.modNScalar()
	pubKeyJ, err := rawPubKeyJ.jacobianPoint()
	if err != nil {
		return nil, NewError(err, "parse publicKeyJ")
	}

	nonceSym := new(secp256k1.JacobianPoint)
	secp256k1.ScalarMultNonConst(nonce, pubKeyJ, nonceSym)

	return NewPointFromJacobianPoint(nonceSym), nil
}

// SumScalars computes the sum of multiple scalars.
// It returns the computed sum as a Scalar.
func SumScalars(rawScalars ...Scalar) Scalar {
	scalars := Scalars(rawScalars).modNScalars()
	return NewScalarFromModNScalar(sumScalars(scalars...))
}

// SolveScalarPolynomial solves a scalar polynomial equation.
// It takes scalars as coefficients and a value x, and returns the result as a scalar and an error, if any.
func SolveScalarPolynomial(rawCoefficients Scalars, rawX Scalar) Scalar {
	coefficients := rawCoefficients.modNScalars()
	x := rawX.modNScalar()
	result := solveScalarPolynomial(coefficients, x)

	return NewScalarFromModNScalar(result)
}

// SumPoints computes the sum of multiple points.
// It returns the computed sum as a Point and an error, if any.
func SumPoints(rawPoints ...Point) (Point, error) {
	points, err := Points(rawPoints).jacobianPoints()
	if err != nil {
		return nil, NewError(err, "parse points")
	}

	return NewPointFromJacobianPoint(sumPoints(points...)), nil
}

// SolvePointPolynomial solves a point polynomial equation.
// It takes points as coefficients and a value x, and returns the result as a point and an error, if any.
func SolvePointPolynomial(rawCoefficientsCommit Points, rawX Scalar) (Point, error) {
	coefficientsCommit, err := rawCoefficientsCommit.jacobianPoints()
	if err != nil {
		return nil, NewError(err, "parse coefficientsCommit")
	}

	x := rawX.modNScalar()
	result := solvePointPolynomial(coefficientsCommit, x)

	return NewPointFromJacobianPoint(result), nil
}

// solveScalarPolynomial solves a scalar polynomial equation of degree 't'.
// The function takes an array of ModNScalar pointers as coefficients (ϕ0, ϕ1, ϕ2, ..., ϕ{t-1}) and a ModNScalar 'x'.
// The order of the coefficients is expected to be from lower degree to higher.
// The function calculates the polynomial using Horner's method for efficient computation:
// ϕ0 + ϕ1*x + ϕ2*x^2 + ... + ϕt*x^{t-1} is computed as: ((...(ϕ{t-1}*x + ϕ{t-2})*x + ϕ{t-3})...)*x + ϕ0
// The result is returned as a pointer to a secp256k1.ModNScalar.
func solveScalarPolynomial(coefficients []*secp256k1.ModNScalar, x *secp256k1.ModNScalar) *secp256k1.ModNScalar {
	var result secp256k1.ModNScalar

	for i := len(coefficients) - 1; i >= 0; i-- {
		// Compute newResult = scalar + oldResult * x
		result.Mul(x).Add(coefficients[i])
	}

	return &result
}

// solvePointPolynomial solves a point polynomial equation of degree 't' on the elliptic curve.
// The function takes an array of JacobianPoint pointers as coefficients (Φ0, Φ1, Φ2, ..., Φ{t-1}) and a ModNScalar 'x'.
// The coefficients are the elliptic curve points obtained from scalar multiplication of the respective ϕ{i}'s with
// the generator point G, i.e., Φ{i} = ϕ{i}G.
// This operation is done to mask the true scalar values ϕ{i} using the properties of the elliptic curve.
// The polynomial is calculated using a variant of Horner's method adapted for elliptic curves:
// Φ0 + Φ1x + Φ2x^2 + ... + Φtx^{t-1} is computed as: ((...(Φ{t-1}*x + Φ{t-2})*x + Φ{t-3})...)*x + Φ0
// The result is returned as a pointer to a secp256k1.JacobianPoint.
func solvePointPolynomial(coefficients []*secp256k1.JacobianPoint, x *secp256k1.ModNScalar) *secp256k1.JacobianPoint {
	var result secp256k1.JacobianPoint

	for i := len(coefficients) - 1; i >= 0; i-- {
		// Compute newValue = point + x * oldValue.
		var xR, newValue secp256k1.JacobianPoint
		secp256k1.ScalarMultNonConst(x, &result, &xR)
		secp256k1.AddNonConst(coefficients[i], &xR, &newValue)

		result = newValue
	}

	return &result
}

// sumPoints sums up multiple *secp256k1.JacobianPoint values.
// total = P1 + P2 + P3 + ... + Pn (Elliptic Curve point addition)
func sumPoints(points ...*secp256k1.JacobianPoint) *secp256k1.JacobianPoint {
	total := new(secp256k1.JacobianPoint)

	for _, point := range points {
		// Add new point to the total.
		newTotal := new(secp256k1.JacobianPoint)
		secp256k1.AddNonConst(total, point, newTotal)

		// Update the total.
		total = newTotal
	}

	return total
}

// sumScalars sums up multiple *secp256k1.ModNScalar values.
// total = (s1 + s2 + s3 + ... + sn) mod (order)
func sumScalars(scalars ...*secp256k1.ModNScalar) *secp256k1.ModNScalar {
	total := new(secp256k1.ModNScalar).SetInt(0)

	for _, scalar := range scalars {
		total.Add(scalar)
	}

	return total
}
