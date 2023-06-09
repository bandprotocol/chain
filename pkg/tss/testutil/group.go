package testutil

import "github.com/bandprotocol/chain/v2/pkg/tss"

type Member struct {
	ID tss.MemberID

	OneTimePrivKey     tss.PrivateKey
	OneTimeSig         tss.Signature
	A0PrivKey          tss.PrivateKey
	A0Sig              tss.Signature
	Coefficients       tss.Scalars
	CoefficientsCommit tss.Points

	KeySyms         tss.PublicKeys
	NonceSyms       tss.PublicKeys
	SecretShares    tss.Scalars
	EncSecretShares tss.Scalars

	PrivKey      tss.PrivateKey
	PubKeySig    tss.Signature
	ComplainSigs tss.Signatures
}

func (m Member) OneTimePubKey() tss.PublicKey {
	return PublicKey(m.OneTimePrivKey)
}

func (m Member) A0PubKey() tss.PublicKey {
	return PublicKey(m.A0PrivKey)
}

func (m Member) PubKey() tss.PublicKey {
	return PublicKey(m.PrivKey)
}

type Group struct {
	ID         tss.GroupID
	DKGContext []byte
	Threshold  uint64
	PubKey     tss.PublicKey
	PubNonce   tss.PublicKey
	Members    []Member
}

func (g Group) GetMember(id tss.MemberID) Member {
	for _, member := range g.Members {
		if member.ID == id {
			return member
		}
	}

	return Member{}
}

func (g Group) GetSize() int {
	return len(g.Members)
}

func (g Group) GetCommits(idx uint64) tss.Points {
	var commits tss.Points
	for _, member := range g.Members {
		commits = append(commits, member.CoefficientsCommit[idx])
	}

	return commits
}

func (g Group) GetAccumulatedCommits() (tss.Points, error) {
	var accCommits tss.Points
	for i := uint64(0); i < g.Threshold; i++ {
		commits := g.GetCommits(i)
		accCommit, err := tss.SumPoints(commits...)
		if err != nil {
			return nil, err
		}

		accCommits = append(accCommits, accCommit)
	}

	return accCommits, nil
}
