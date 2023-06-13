package tss_test

import "github.com/bandprotocol/chain/v2/pkg/tss"

func (suite *TSSTestSuite) TestError() {
	tests := []struct {
		name      string
		err       error
		expString string
	}{{
		"plain description",
		tss.NewError(tss.ErrInvalidLength, "length"),
		"length: invalid length",
	}, {
		"format description",
		tss.NewError(tss.ErrInvalidLength, "length: %d", 5),
		"length: 5: invalid length",
	}}

	for _, t := range tests {
		suite.Run(t.name, func() {
			str := t.err.Error()
			suite.Require().Equal(t.expString, str)
		})
	}
}
