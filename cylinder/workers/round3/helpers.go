package round3

import (
	"github.com/bandprotocol/chain/v2/cylinder/client"
	"github.com/bandprotocol/chain/v2/cylinder/store"
	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// getOwnPrivKey calculates the own private key for the group member.
// It returns the own private key, a slice of complaints (if any), and an error, if any.
func getOwnPrivKey(group store.Group, groupRes *client.GroupResponse) (tss.Scalar, []types.Complaint, error) {
	var secretShares tss.Scalars
	var complaints []types.Complaint
	for senderID := uint64(1); senderID <= groupRes.Group.Size_; senderID++ {
		// Calculate your own secret value
		if senderID == uint64(group.MemberID) {
			secretShare, err := tss.ComputeSecretShare(group.Coefficients, group.MemberID)
			if err != nil {
				return nil, nil, err
			}
			secretShares = append(secretShares, secretShare)
			continue
		}

		secretShare, complaint, err := getSecretShare(
			group.MemberID,
			tss.MemberID(senderID),
			group.OneTimePrivKey,
			groupRes,
		)
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
// It takes the Member ID of sender and receiver, private key of receiver, and the group response as input.
// It returns the secret share, complaint if secret share is not valid, and any error encountered during the process.
func getSecretShare(
	receiverID tss.MemberID,
	senderID tss.MemberID,
	privKeyReceiver tss.Scalar,
	groupRes *client.GroupResponse,
) (tss.Scalar, *types.Complaint, error) {
	// Get Round1Info of receiver
	r1Receiver, err := groupRes.GetRound1Info(receiverID)
	if err != nil {
		return nil, nil, err
	}

	// Get Round1Info of complinant
	r1Sender, err := groupRes.GetRound1Info(senderID)
	if err != nil {
		return nil, nil, err
	}

	// Get encrypted secret share for complaner from respondent
	encSecretShare, err := groupRes.GetEncryptedSecretShare(senderID, receiverID)
	if err != nil {
		return nil, nil, err
	}

	// Calculate keySym
	keySym, err := tss.ComputeKeySym(privKeyReceiver, r1Sender.OneTimePubKey)
	if err != nil {
		return nil, nil, err
	}

	// Decrypt secret share
	secretShare, err := tss.DecryptSecretShare(encSecretShare, keySym)
	if err != nil {
		return nil, nil, err
	}

	// Verify secret share
	err = tss.VerifySecretShare(receiverID, secretShare, r1Sender.CoefficientCommits)
	if err != nil {
		// Generate complaint if we fail to verify secret share
		sig, keySym, err := tss.SignComplaint(
			r1Receiver.OneTimePubKey,
			r1Sender.OneTimePubKey,
			privKeyReceiver,
		)
		if err != nil {
			return nil, nil, err
		}

		complaint := &types.Complaint{
			Complainant: receiverID,
			Respondent:  senderID,
			KeySym:      keySym,
			Signature:   sig,
		}

		return nil, complaint, nil
	}

	return secretShare, nil, nil
}
