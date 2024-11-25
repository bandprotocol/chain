package bandtss_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/bandtss"
)

func TestEncoderPrefix(t *testing.T) {
	require.Equal(t, []byte(bandtss.GroupTransitionMsgPrefix), tss.Hash([]byte("Transition"))[:4])
}
