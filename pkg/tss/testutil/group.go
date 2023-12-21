package testutil

import "github.com/bandprotocol/chain/v2/pkg/tss"

type Member struct {
	ID tss.MemberID

	OneTimePrivKey     tss.Scalar
	OneTimeSignature   tss.Signature
	A0PrivKey          tss.Scalar
	A0Signature        tss.Signature
	Coefficients       tss.Scalars
	CoefficientCommits tss.Points

	KeySyms         tss.Points
	SecretShares    tss.Scalars
	EncSecretShares tss.EncSecretShares

	PrivKey             tss.Scalar
	PubKeySignature     tss.Signature
	ComplaintSignatures tss.ComplaintSignatures
}

func (m Member) OneTimePubKey() tss.Point {
	return Point(m.OneTimePrivKey)
}

func (m Member) A0PubKey() tss.Point {
	return Point(m.A0PrivKey)
}

func (m Member) PubKey() tss.Point {
	return Point(m.PrivKey)
}

func CopyMember(src Member) Member {
	return Member{
		ID:                  src.ID,
		OneTimePrivKey:      Copy(src.OneTimePrivKey),
		OneTimeSignature:    Copy(src.OneTimeSignature),
		A0PrivKey:           Copy(src.A0PrivKey),
		A0Signature:         Copy(src.A0Signature),
		Coefficients:        CopySlice(src.Coefficients),
		CoefficientCommits:  CopySlice(src.CoefficientCommits),
		KeySyms:             CopySlice(src.KeySyms),
		SecretShares:        CopySlice(src.SecretShares),
		EncSecretShares:     src.EncSecretShares.Clone(),
		PrivKey:             Copy(src.PrivKey),
		PubKeySignature:     Copy(src.PubKeySignature),
		ComplaintSignatures: CopySlice(src.ComplaintSignatures),
	}
}

func CopyMembers(src []Member) []Member {
	var dst []Member
	for _, m := range src {
		dst = append(dst, CopyMember(m))
	}

	return dst
}

type Group struct {
	ID         tss.GroupID
	DKGContext []byte
	Threshold  uint64
	PubKey     tss.Point
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
		commits = append(commits, member.CoefficientCommits[idx])
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

func CopyGroup(src Group) Group {
	return Group{
		ID:         src.ID,
		DKGContext: Copy(src.DKGContext),
		Threshold:  src.Threshold,
		PubKey:     Copy(src.PubKey),
		Members:    CopyMembers(src.Members),
	}
}
