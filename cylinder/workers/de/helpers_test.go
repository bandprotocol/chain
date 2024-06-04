package de_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v2/cylinder/workers/de"
	"github.com/bandprotocol/chain/v2/x/tss/types"
)

type MockDEGetter struct{}

func (m *MockDEGetter) HasDE(de types.DE) bool { return false }

func TestGenerateDEs(t *testing.T) {
	privDEs, err := de.GenerateDEs(10, []byte("secret"), &MockDEGetter{})
	assert.NoError(t, err)

	for _, privDE := range privDEs {
		err = privDE.PrivD.Validate()
		assert.NoError(t, err)

		err = privDE.PrivE.Validate()
		assert.NoError(t, err)
	}
}
