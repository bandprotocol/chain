package round3

import (
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// getOwnPrivKey calculates the own private key for the group member.
// It returns the own private key, a slice of complaints (if any), and an error, if any.
func getOwnPrivKey(group store.Group, groupRes *client.GroupResponse) (tss.PrivateKey, []types.Complaint, error) {
	var secretShares tss.Scalars
	var complaints []types.Complaint
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

		secretShare, complaint, err := getSecretShare(group.MemberID, tss.MemberID(j), group.OneTimePrivKey, groupRes)
		if err != nil {
			return nil, nil, err
		}

		if complaint != nil {
			complaints = append(complaints, *complaint)
			continue
		}

		// Add secret share if verification is successful
		secretShares = append(secretShares, secretShare)
	}

	if len(complaints) > 0 {
		return nil, complaints, nil
	}

	ownPrivKey, err := tss.ComputeOwnPrivateKey(secretShares...)
	if err != nil {
		return nil, nil, err
	}

	return ownPrivKey, nil, nil
}

// getSecretShare calculates and retrieves the decrypted secret share between two members.
// It takes the Member ID of I and J, private key of Member I, and the group response as input.
// It returns the secret share, complaint if secret share is not valid, and any error encountered during the process.
func getSecretShare(
	i tss.MemberID,
	j tss.MemberID,
	privKeyI tss.PrivateKey,
	groupRes *client.GroupResponse,
) (tss.Scalar, *types.Complaint, error) {
	// Get Round1Info of I
	round1InfoI, err := groupRes.GetRound1Info(i)
	if err != nil {
		return nil, nil, err
	}

	// Get Round1Info of J
	round1InfoJ, err := groupRes.GetRound1Info(j)
	if err != nil {
		return nil, nil, err
	}

	// Get encrypted secret share for I from J
	encSecretShare, err := groupRes.GetEncryptedSecretShare(j, i)
	if err != nil {
		return nil, nil, err
	}

	// Calculate keySym
	keySym, err := tss.ComputeKeySym(privKeyI, round1InfoJ.OneTimePubKey)
	if err != nil {
		return nil, nil, err
	}

	// Decrypt secret share
	secretShare, err := tss.DecryptSecretShare(encSecretShare, keySym)
	if err != nil {
		return nil, nil, err
	}

	// Verify secret share
	err = tss.VerifySecretShare(i, secretShare, round1InfoJ.CoefficientsCommit)
	if err != nil {
		// Generate complaint if we fail to verify secret share
		sig, keySym, err := tss.SignComplaint(
			round1InfoI.OneTimePubKey,
			round1InfoJ.OneTimePubKey,
			privKeyI,
		)
		if err != nil {
			return nil, nil, err
		}

		complaint := &types.Complaint{
			I:      i,
			J:      j,
			KeySym: keySym,
			Sig:    sig,
		}

		return nil, complaint, nil
	}

	return secretShare, nil, nil
}
