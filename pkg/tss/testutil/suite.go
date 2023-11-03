package testutil

import (
	"fmt"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
}

func (suite *Suite) RunOnMember(testCases []TestCase, f func(TestCase, Member)) {
	for _, tc := range testCases {
		for i := 1; i <= tc.Group.GetSize(); i++ {
			suite.Run(fmt.Sprintf("%s, Member: %d", tc.Name, i), func() {
				member := tc.Group.GetMember(tss.NewMemberID(i))
				f(tc, member)
			})
		}
	}
}

func (suite *Suite) RunOnPairMembers(testCases []TestCase, f func(TestCase, Member, Member)) {
	for _, tc := range testCases {
		for i := 1; i <= tc.Group.GetSize(); i++ {
			for j := 1; j <= tc.Group.GetSize(); j++ {
				if i == j {
					continue
				}

				suite.Run(fmt.Sprintf("%s, MemberI: %d, MemberJ: %d", tc.Name, i, j), func() {
					memberI := tc.Group.GetMember(tss.NewMemberID(i))
					memberJ := tc.Group.GetMember(tss.NewMemberID(j))
					f(tc, memberI, memberJ)
				})
			}
		}
	}
}

func (suite *Suite) RunOnSigning(testCases []TestCase, f func(TestCase, Signing)) {
	for _, tc := range testCases {
		for _, signing := range tc.Signings {
			suite.Run(fmt.Sprintf("%s, Signing: %d", tc.Name, signing.ID), func() {
				f(tc, signing)
			})
		}
	}
}

func (suite *Suite) RunOnAssignedMember(testCases []TestCase, f func(TestCase, Signing, AssignedMember)) {
	for _, tc := range testCases {
		for _, signing := range tc.Signings {
			for _, assignedMember := range signing.AssignedMembers {
				suite.Run(fmt.Sprintf("%s, Signing: %d, Member: %d", tc.Name, signing.ID, assignedMember.ID), func() {
					f(tc, signing, assignedMember)
				})
			}
		}
	}
}
