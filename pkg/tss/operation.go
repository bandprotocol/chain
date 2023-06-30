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

// solveScalarPolynomial solves a scalar polynomial equation.
// It takes scalars as coefficients and a value x, and returns the result as a *secp256k1.ModNScalar.
func solveScalarPolynomial(coefficients []*secp256k1.ModNScalar, x *secp256k1.ModNScalar) *secp256k1.ModNScalar {
	var result secp256k1.ModNScalar

	for i := len(coefficients) - 1; i >= 0; i-- {
		// Compute newResult = scalar + oldResult * x
		result.Mul(x).Add(coefficients[i])
	}

	return &result
}

// solvePointPolynomial solves a point polynomial equation.
// It takes points as coefficients and a value x, and returns the result as a *secp256k1.JacobianPoint.
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
func sumScalars(scalars ...*secp256k1.ModNScalar) *secp256k1.ModNScalar {
	total := new(secp256k1.ModNScalar).SetInt(0)

	for _, scalar := range scalars {
		total.Add(scalar)
	}

	return total
}
