package de_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/de"
)

func TestGenerateDEs(t *testing.T) {
	privDEs, err := de.GenerateDEs(10, []byte("aaaaaa"))
	assert.Nil(t, err)

	for _, privDE := range privDEs {
		err = privDE.PrivD.Validate()
		assert.Nil(t, err)

		err = privDE.PrivE.Validate()
		assert.Nil(t, err)
	}
}
