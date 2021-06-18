package main

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sync"
	"time"
)

// LimitStore defines the interface for the store of faucet limits.
type LimitStore interface {
	// Get returns limit from store by the given key.
	Get(key string) (*WithdrawalLimit, bool)
	// Set sets limit to store by the given key.
	Set(key string, value *WithdrawalLimit)
	// Remove removes limit from store by the given key.
	Remove(key string)
	// Clean removes deprecated limits.
	Clean(timeLimit time.Duration)
}

// limitStore defines the store of faucet limits.
type limitStore struct {
	sync.RWMutex
	limits map[string]*WithdrawalLimit
}

// WithdrawalLimit defines a store of faucet withdrawals.
type WithdrawalLimit struct {
	WithdrawalAmount sdk.Coins
	LastWithdrawals  map[string]time.Time
}

// NewLimitStore creates a new LimitStore.
func NewLimitStore() LimitStore {
	return &limitStore{
		RWMutex: sync.RWMutex{},
		limits:  make(map[string]*WithdrawalLimit),
	}
}

// Get implements LimitStore interface.
func (s *limitStore) Get(key string) (*WithdrawalLimit, bool) {
	s.RLock()
	defer s.RUnlock()
	res, ok := s.limits[key]
	return res, ok
}

// Set implements LimitStore interface.
func (s *limitStore) Set(key string, value *WithdrawalLimit) {
	s.Lock()
	defer s.Unlock()
	s.limits[key] = value
}

// Remove implements LimitStore interface.
func (s *limitStore) Remove(key string) {
	s.Lock()
	defer s.Unlock()
	delete(s.limits, key)
}

// Clean implements LimitStore interface.
func (s *limitStore) Clean(timeLimit time.Duration) {
	toRemove := make([]string, 0, 10)
	for k, v := range s.limits {
		denomsUnpend := 0
		for _, lw := range v.LastWithdrawals {
			if time.Now().Sub(lw) > timeLimit {
				denomsUnpend++
			}
		}
		if denomsUnpend == len(v.LastWithdrawals) {
			toRemove = append(toRemove, k)
		}
	}
	for _, k := range toRemove {
		s.Remove(k)
	}
}
