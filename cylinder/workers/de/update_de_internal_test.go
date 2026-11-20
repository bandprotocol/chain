package de

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestCacheDEConcurrentAccess exercises the cacheDEs map from two goroutines the same
// way the worker loop and listenMsgResponses do. Run with -race to catch regressions.
func TestCacheDEConcurrentAccess(t *testing.T) {
	u := &UpdateDE{cacheDEs: make(map[uint64]int64)}

	const n = 2000
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := uint64(1); i <= n; i++ {
			u.setCacheDE(i, int64(i))
		}
	}()

	go func() {
		defer wg.Done()
		for i := uint64(1); i <= n; i++ {
			u.popCacheDE(i)
		}
	}()

	wg.Wait()

	for i := uint64(1); i <= n; i++ {
		u.popCacheDE(i)
	}
	require.Empty(t, u.cacheDEs)
}

func TestPopCacheDE(t *testing.T) {
	u := &UpdateDE{cacheDEs: make(map[uint64]int64)}
	u.setCacheDE(7, 42)

	require.Equal(t, int64(42), u.popCacheDE(7))
	require.Equal(t, int64(0), u.popCacheDE(7))
	require.Empty(t, u.cacheDEs)
}
