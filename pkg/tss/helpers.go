package tss

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func ConcatBytes(data ...[]byte) []byte {
	var res []byte
	for _, b := range data {
		res = append(res, b...)
	}

	return res
}

func GenerateKeyPairs(n uint64) (KeyPairs, error) {
	var kps KeyPairs
	for i := uint64(0); i < n; i++ {
		kp, err := GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		kps = append(kps, kp)
	}

	return kps, nil
}

func GenerateKeyPair() (KeyPair, error) {
	key, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		PrivateKey: key.Serialize(),
		PublicKey:  key.PubKey().SerializeCompressed(),
	}, nil
}

func GenerateKeySymIJ(rawPrivKeyI PrivateKey, rawPubKeyJ PublicKey) (PublicKey, error) {
	privKeyI := rawPrivKeyI.Parse()

	pubKeyJ, err := rawPubKeyJ.Point()
	if err != nil {
		return nil, err
	}

	var keySymIJ secp256k1.JacobianPoint
	secp256k1.ScalarMultNonConst(&privKeyI.Key, pubKeyJ, &keySymIJ)

	return ParsePublicKey(&keySymIJ), nil
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

// y = f_ij(x) = a_0 + a_1 * x^1 + ... + a_n-1 * x^(n-1) + a_n * x^n
// rawScalars = coefficients = a_0, a_1, ..., a_n-1, a_n
// rawX = x
func solveScalarEquation(scalars []*secp256k1.ModNScalar, x *secp256k1.ModNScalar) *secp256k1.ModNScalar {
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

func solvePointEquation(points []*secp256k1.JacobianPoint, x *secp256k1.ModNScalar) *secp256k1.JacobianPoint {
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
