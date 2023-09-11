package tss

// Round1Info contains the data for round 1 of the DKG process of TSS
type Round1Info struct {
	OneTimePrivKey     Scalar
	OneTimePubKey      Point
	OneTimeSignature   Signature
	A0PrivKey          Scalar
	A0PubKey           Point
	A0Signature        Signature
	Coefficients       Scalars
	CoefficientCommits Points
}

// GenerateRound1Info generates the data of round 1 for a member in the DKG process of TSS
func GenerateRound1Info(
	mid MemberID,
	threshold uint64,
	dkgContext []byte,
) (*Round1Info, error) {
	// Generate threshold + 1 key pairs (commits, onetime).
	kps, err := GenerateKeyPairs(threshold + 1)
	if err != nil {
		return nil, NewError(err, "generate key pairs")
	}

	// Get one-time information.
	oneTimePrivKey := kps[0].PrivKey
	oneTimePubKey := kps[0].PubKey
	oneTimeSignature, err := SignOneTime(mid, dkgContext, oneTimePubKey, oneTimePrivKey)
	if err != nil {
		return nil, NewError(err, "sign one time")
	}

	// Get a0 information.
	a0PrivKey := kps[1].PrivKey
	a0PubKey := kps[1].PubKey
	a0Signature, err := SignA0(mid, dkgContext, a0PubKey, a0PrivKey)
	if err != nil {
		return nil, NewError(err, "sign A0")
	}

	// Get coefficients.
	var coefficientCommits Points
	var coefficients Scalars
	for i := 1; i < len(kps); i++ {
		coefficientCommits = append(coefficientCommits, Point(kps[i].PubKey))
		coefficients = append(coefficients, Scalar(kps[i].PrivKey))
	}

	return &Round1Info{
		OneTimePrivKey:     oneTimePrivKey,
		OneTimePubKey:      oneTimePubKey,
		OneTimeSignature:   oneTimeSignature,
		A0PrivKey:          a0PrivKey,
		A0PubKey:           a0PubKey,
		A0Signature:        a0Signature,
		Coefficients:       coefficients,
		CoefficientCommits: coefficientCommits,
	}, nil
}

// SignA0 generates a signature for the A0 in round 1.
func SignA0(
	mid MemberID,
	dkgContext []byte,
	a0Pub Point,
	a0Priv Scalar,
) (Signature, error) {
	var nonce, challenge Scalar
	var pubNonce Point
	var err error
	// We omit implementing a timeout here as the probability of the hash exceeding
	// the curve's order is exceptionally small (1 in 2.67e+38).
	for {
		nonce, pubNonce, err = GenerateDKGNonce()
		if err != nil {
			return nil, err
		}

		challenge, err = HashRound1A0(pubNonce, mid, dkgContext, a0Pub)
		if err == nil {
			break
		}
	}

	return Sign(a0Priv, challenge, nonce, nil)
}

// VerifyA0Signature verifies the signature for the A0 in round 1.
func VerifyA0Signature(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	a0Pub Point,
) error {
	challenge, err := HashRound1A0(signature.R(), mid, dkgContext, a0Pub)
	if err != nil {
		return err
	}

	return Verify(signature.R(), signature.S(), challenge, a0Pub, nil, nil)
}

// SignOneTime generates a signature for the one-time in round 1.
func SignOneTime(
	mid MemberID,
	dkgContext []byte,
	oneTimePub Point,
	onetimePriv Scalar,
) (Signature, error) {
	var nonce, challenge Scalar
	var pubNonce Point
	var err error
	// We omit implementing a timeout here as the probability of the hash exceeding
	// the curve's order is exceptionally small (1 in 2.67e+38).
	for {
		nonce, pubNonce, err = GenerateDKGNonce()
		if err != nil {
			return nil, err
		}

		challenge, err = HashRound1OneTime(pubNonce, mid, dkgContext, oneTimePub)
		if err == nil {
			break
		}
	}

	return Sign(onetimePriv, challenge, nonce, nil)
}

// VerifyOneTimeSignature verifies the signature for the one-time in round 1.
func VerifyOneTimeSignature(
	mid MemberID,
	dkgContext []byte,
	signature Signature,
	oneTimePub Point,
) error {
	challenge, err := HashRound1OneTime(signature.R(), mid, dkgContext, oneTimePub)
	if err != nil {
		return err
	}

	return Verify(signature.R(), signature.S(), challenge, oneTimePub, nil, nil)
}
