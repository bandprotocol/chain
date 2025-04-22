package testutil

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	bandtesting "github.com/bandprotocol/chain/v3/testing"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

// DEWithPrivateNonce represents private value (D, E) used in tss signing process.
type DEWithPrivateNonce struct {
	PubDE types.DE
	PrivD tss.Scalar
	PrivE tss.Scalar
}

func GenerateAccounts(n uint64) []bandtesting.Account {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	accounts := make([]bandtesting.Account, n)
	for i := 0; i < len(accounts); i++ {
		accounts[i] = bandtesting.CreateArbitraryAccount(r)
	}

	return accounts
}

func GenerateDE(secret tss.Scalar) DEWithPrivateNonce {
	privD, err := tss.GenerateSigningNonce(secret)
	if err != nil {
		panic(err)
	}

	privE, err := tss.GenerateSigningNonce(secret)
	if err != nil {
		panic(err)
	}
	pubDE := types.NewDE(privD.Point(), privE.Point())

	return DEWithPrivateNonce{PrivD: privD, PrivE: privE, PubDE: pubDE}
}

func GenerateSignature(
	signing types.Signing,
	assignedMembers types.AssignedMembers,
	memberID tss.MemberID,
	de DEWithPrivateNonce,
	privKey tss.Scalar,
) (tss.Signature, error) {
	var member types.AssignedMember
	for _, am := range assignedMembers {
		if am.MemberID == memberID {
			member = am
		}
	}

	if member.MemberID == 0 {
		return nil, fmt.Errorf("member not found")
	}

	// Compute own private nonce
	privNonce, err := tss.ComputeOwnPrivNonce(de.PrivD, de.PrivE, member.BindingFactor)
	if err != nil {
		return nil, err
	}

	// Compute lagrange
	lagrange, err := tss.ComputeLagrangeCoefficient(memberID, assignedMembers.MemberIDs())
	if err != nil {
		return nil, err
	}

	return tss.SignSigning(
		signing.GroupPubNonce,
		signing.GroupPubKey,
		signing.Message,
		lagrange,
		privNonce,
		privKey,
	)
}
