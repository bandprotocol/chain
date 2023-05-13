package tss

import "github.com/decred/dcrd/dcrec/secp256k1/v4"

func ComputeKeySym(rawPrivKeyI PrivateKey, rawPubKeyJ PublicKey) (PublicKey, error) {
	privKeyI := rawPrivKeyI.Scalar()

	pubKeyJ, err := rawPubKeyJ.Point()
	if err != nil {
		return nil, err
	}

	keySym := new(secp256k1.JacobianPoint)
	secp256k1.ScalarMultNonConst(privKeyI, pubKeyJ, keySym)

	return ParsePublicKey(keySym), nil
}

func ComputeNonceSym(rawNonce Scalar, rawPubKeyJ PublicKey) (PublicKey, error) {
	pubKeyJ, err := rawPubKeyJ.Point()
	if err != nil {
		return nil, err
	}

	nonce := rawNonce.Parse()

	nG := new(secp256k1.JacobianPoint)
	secp256k1.ScalarBaseMultNonConst(nonce, nG)
	nG.ToAffine()

	if nG.Y.IsOdd() {
		nonce.Negate()
	}

	nonceSym := new(secp256k1.JacobianPoint)
	secp256k1.ScalarMultNonConst(nonce, pubKeyJ, nonceSym)

	return ParsePublicKey(nonceSym), nil
}

func SumPoints(rawPoints ...Point) (Point, error) {
	points, err := Points(rawPoints).Parse()
	if err != nil {
		return nil, err
	}

	return ParsePoint(sumPoints(points...)), nil
}

func SumScalars(rawScalars ...Scalar) Scalar {
	scalars := Scalars(rawScalars).Parse()
	return ParseScalar(sumScalars(scalars...))
}

// y = f_ij(x) = a_0 + a_1 * x^1 + ... + a_n-1 * x^(n-1) + a_n * x^n
// rawScalars = coefficients = a_0, a_1, ..., a_n-1, a_n
// rawX = x
func solveScalarPolynomial(scalars []*secp256k1.ModNScalar, x *secp256k1.ModNScalar) *secp256k1.ModNScalar {
	currentX := new(secp256k1.ModNScalar).SetInt(1)

	// calculate each term ( a_i * x^i )
	var terms []*secp256k1.ModNScalar
	for _, scalar := range scalars {
		// term = ax^i
		term := *scalar
		term.Mul(currentX)
		terms = append(terms, &term)

		// currentX *= x
		currentX.Mul(x)
	}

	// sum up all terms
	return sumScalars(terms...)
}

func solvePointPolynomial(points []*secp256k1.JacobianPoint, x *secp256k1.ModNScalar) *secp256k1.JacobianPoint {
	currentX := new(secp256k1.ModNScalar).SetInt(1)

	var terms []*secp256k1.JacobianPoint
	for _, point := range points {
		// compute each term (x^i * c_i)
		var term secp256k1.JacobianPoint
		secp256k1.ScalarMultNonConst(currentX, point, &term)
		terms = append(terms, &term)

		// new_x *= x
		currentX.Mul(x)
	}

	return sumPoints(terms...)
}

func sumPoints(points ...*secp256k1.JacobianPoint) *secp256k1.JacobianPoint {
	total := new(secp256k1.JacobianPoint)

	for _, point := range points {
		// add new point to the total
		newTotal := new(secp256k1.JacobianPoint)
		secp256k1.AddNonConst(total, point, newTotal)

		// update the total
		total = newTotal
	}

	return total
}

func sumScalars(scalars ...*secp256k1.ModNScalar) *secp256k1.ModNScalar {
	total := new(secp256k1.ModNScalar).SetInt(0)

	for _, scalar := range scalars {
		total.Add(scalar)
	}

	return total
}
