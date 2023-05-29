package store

import (
	"encoding/json"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
	dbm "github.com/tendermint/tm-db"
)

// Store represents a data store for storing data information for Cylinder process
type Store struct {
	DB dbm.DB
}

// NewStore creates a new instance of Store with the provided database.
func NewStore(db dbm.DB) *Store {
	return &Store{
		DB: db,
	}
}

// SetGroup stores the group information by the given groupID.
func (s *Store) SetGroup(groupID tss.GroupID, group Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}

	return s.DB.Set(GroupStoreKey(groupID), bytes)
}

// GetGroup retrieves the group information by the given groupID.
func (s *Store) GetGroup(groupID tss.GroupID) (Group, error) {
	bytes, err := s.DB.Get(GroupStoreKey(groupID))

	var group Group
	err = json.Unmarshal(bytes, &group)
	if err != nil {
		return Group{}, err
	}

	return group, err
}

// SetDE stores the private (d, E) by the given public (D, E).
func (s *Store) SetDE(pubDE types.DE, privDE DE) error {
	bytes, err := json.Marshal(privDE)
	if err != nil {
		return err
	}

	return s.DB.Set(DEStoreKey(pubDE), bytes)
}

// GetDE retrieves the private (d, E) by the given public (D, E)
func (s *Store) GetDE(pubDE types.DE) (DE, error) {
	bytes, err := s.DB.Get(DEStoreKey(pubDE))

	var de DE
	err = json.Unmarshal(bytes, &de)
	if err != nil {
		return DE{}, err
	}

	return de, err
}

// RemoveDE deletes the private (d, E) by the given public (D, E)
func (s *Store) RemoveDE(pubDE types.DE) error {
	return s.DB.DeleteSync(DEStoreKey(pubDE))
}
