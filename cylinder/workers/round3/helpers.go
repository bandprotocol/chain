package round3

import (
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// getOwnPrivKey calculates the own private key for the group member.
// It returns the own private key, a slice of complaints (if any), and an error, if any.
func getOwnPrivKey(group store.Group, groupRes *client.GroupResponse) (tss.PrivateKey, []types.Complain, error) {
	var secretShares tss.Scalars
	var complains []types.Complain
	for j := uint64(1); j <= groupRes.Group.Size_; j++ {
		// Calculate your own secret value
		if j == uint64(group.MemberID) {
			secretShare, err := tss.ComputeSecretShare(group.Coefficients, group.MemberID)
			if err != nil {
				return nil, nil, err
			}
			secretShares = append(secretShares, secretShare)
			continue
		}

		secretShare, complain, err := getSecretShare(group.MemberID, tss.MemberID(j), group.OneTimePrivKey, groupRes)
		if err != nil {
			return nil, nil, err
		}

		if complain != nil {
			complains = append(complains, *complain)
			continue
		}

		// Add secret share if verification is successful
		secretShares = append(secretShares, secretShare)
	}

	if len(complains) > 0 {
		return nil, complains, nil
	}

	ownPrivKey, err := tss.ComputeOwnPrivateKey(secretShares...)
	if err != nil {
		return nil, nil, err
	}

	return ownPrivKey, nil, nil
}

// getSecretShare calculates and retrieves the decrypted secret share between two members.
// It takes the Member ID of I and J, private key of Member I, and the group response as input.
// It returns the secret share, complain if secret share is not valid, and any error encountered during the process.
func getSecretShare(
	i tss.MemberID,
	j tss.MemberID,
	privKeyI tss.PrivateKey,
	groupRes *client.GroupResponse,
) (tss.Scalar, *types.Complain, error) {
	// Get Round1Data of I
	round1DataI, err := groupRes.GetRound1Data(i)
	if err != nil {
		return nil, nil, err
	}

	// Get Round1Data of J
	round1DataJ, err := groupRes.GetRound1Data(j)
	if err != nil {
		return nil, nil, err
	}

	// Get encrypted secret share for I from J
	encSecretShare, err := groupRes.GetEncryptedSecretShare(j, i)
	if err != nil {
		return nil, nil, err
	}

	// Calculate keySym
	keySym, err := tss.ComputeKeySym(privKeyI, round1DataJ.OneTimePubKey)
	if err != nil {
		return nil, nil, err
	}

	// Decrypt secret share
	secretShare, err := tss.Decrypt(encSecretShare, keySym)
	if err != nil {
		return nil, nil, err
	}

	// Verify secret share
	err = tss.VerifySecretShare(i, secretShare, round1DataJ.CoefficientsCommit)
	if err != nil {
		// Generate complaint if we fail to verify secret share
		sig, keySym, err := tss.SignComplain(
			round1DataI.OneTimePubKey,
			round1DataJ.OneTimePubKey,
			privKeyI,
		)
		if err != nil {
			return nil, nil, err
		}

		complain := &types.Complain{
			I:      i,
			J:      j,
			KeySym: keySym,
			Sig:    sig,
		}

		return nil, complain, nil
	}

	return secretShare, nil, nil
}
