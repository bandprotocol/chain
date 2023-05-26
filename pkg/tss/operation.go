package tss

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ComputeKeySym computes the key symmetry between a private key and a public key.
// It returns the computed key symmetry as a PublicKey and an error, if any.
func ComputeKeySym(rawPrivKeyI PrivateKey, rawPubKeyJ PublicKey) (PublicKey, error) {
	privKeyI, err := rawPrivKeyI.Scalar()
	if err != nil {
		return nil, err
	}

	pubKeyJ, err := rawPubKeyJ.Point()
	if err != nil {
		return nil, err
	}

	keySym := new(secp256k1.JacobianPoint)
	secp256k1.ScalarMultNonConst(privKeyI, pubKeyJ, keySym)

	return ParsePublicKey(keySym), nil
}

// ComputeNonceSym computes the nonce symmetry between a nonce value and a public key.
// It returns the computed nonce symmetry as a PublicKey and an error, if any.
func ComputeNonceSym(rawNonce Scalar, rawPubKeyJ PublicKey) (PublicKey, error) {
	nonce, err := rawNonce.Parse()
	if err != nil {
		return nil, err
	}

	pubKeyJ, err := rawPubKeyJ.Point()
	if err != nil {
		return nil, err
	}

	nonceSym := new(secp256k1.JacobianPoint)
	secp256k1.ScalarMultNonConst(nonce, pubKeyJ, nonceSym)

	return ParsePublicKey(nonceSym), nil
}

// SumPoints computes the sum of multiple points.
// It returns the computed sum as a Point and an error, if any.
func SumPoints(rawPoints ...Point) (Point, error) {
	points, err := Points(rawPoints).Parse()
	if err != nil {
		return nil, err
	}

	return ParsePoint(sumPoints(points...)), nil
}

// SumScalars computes the sum of multiple scalars.
// It returns the computed sum as a Scalar.
func SumScalars(rawScalars ...Scalar) (Scalar, error) {
	scalars, err := Scalars(rawScalars).Parse()
	if err != nil {
		return nil, err
	}

	return ParseScalar(sumScalars(scalars...)), nil
}

// solveScalarPolynomial solves a scalar polynomial equation.
// It takes scalars as coefficients and a value x, and returns the result as a *secp256k1.ModNScalar.
func solveScalarPolynomial(scalars []*secp256k1.ModNScalar, x *secp256k1.ModNScalar) *secp256k1.ModNScalar {
	var result secp256k1.ModNScalar

	for i := len(scalars) - 1; i >= 0; i-- {
		// Compute newResult = oldResult * x + scalar
		result.Mul(x).Add(scalars[i])
	}

	return &result
}

// solvePointPolynomial solves a point polynomial equation.
// It takes points as coefficients and a value x, and returns the result as a *secp256k1.JacobianPoint.
func solvePointPolynomial(points []*secp256k1.JacobianPoint, x *secp256k1.ModNScalar) *secp256k1.JacobianPoint {
	var result secp256k1.JacobianPoint

	for i := len(points) - 1; i >= 0; i-- {
		// Compute newValue = point + x * oldValue.
		var xR, newValue secp256k1.JacobianPoint
		secp256k1.ScalarMultNonConst(x, &result, &xR)
		secp256k1.AddNonConst(points[i], &xR, &newValue)

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
