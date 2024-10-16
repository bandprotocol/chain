package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
)

func TestAssignedMembersPubDs(t *testing.T) {
	// Create a sample AssignedMembers slice
	assignedMembers := types.AssignedMembers{
		{PubD: []byte("1")},
		{PubD: []byte("2")},
		{PubD: []byte("3")},
	}

	// Expected result
	expected := tss.Points{
		tss.Point([]byte("1")),
		tss.Point([]byte("2")),
		tss.Point([]byte("3")),
	}

	// Call the PubDs method
	result := assignedMembers.PubDs()

	// Check if the result matches the expected output
	assert.Equal(t, expected, result, "PubDs result does not match expected output")
}

func TestAssignedMembersPubEs(t *testing.T) {
	// Create a sample AssignedMembers slice
	assignedMembers := types.AssignedMembers{
		{PubE: []byte("1")},
		{PubE: []byte("2")},
		{PubE: []byte("3")},
	}

	// Expected result
	expected := tss.Points{
		tss.Point([]byte("1")),
		tss.Point([]byte("2")),
		tss.Point([]byte("3")),
	}

	// Call the PubEs method
	result := assignedMembers.PubEs()

	// Check if the result matches the expected output
	assert.Equal(t, expected, result, "PubEs result does not match expected output")
}
