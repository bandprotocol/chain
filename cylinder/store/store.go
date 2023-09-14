package store

import (
	"encoding/json"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	"github.com/bandprotocol/chain/v2/x/tss/types"
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

// SetDKG stores the dkg information by the given group id.
func (s *Store) SetDKG(groupID tss.GroupID, dkg DKG) error {
	bytes, err := json.Marshal(dkg)
	if err != nil {
		return err
	}

	return s.DB.Set(DKGStoreKey(groupID), bytes)
}

// GetDKG retrieves the dkg information by the given group id.
func (s *Store) GetDKG(groupID tss.GroupID) (DKG, error) {
	bytes, err := s.DB.Get(DKGStoreKey(groupID))

	var dkg DKG
	err = json.Unmarshal(bytes, &dkg)
	if err != nil {
		return DKG{}, err
	}

	return dkg, err
}

// DeleteDKG deletes the dkg information by the given group id.
func (s *Store) DeleteDKG(groupID tss.GroupID) error {
	return s.DB.DeleteSync(DKGStoreKey(groupID))
}

// SetGroup stores the group information by the given public key.
func (s *Store) SetGroup(pubKey tss.Point, group Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}

	return s.DB.Set(GroupStoreKey(pubKey), bytes)
}

// GetGroup retrieves the group information by the given public key.
func (s *Store) GetGroup(pubKey tss.Point) (Group, error) {
	bytes, err := s.DB.Get(GroupStoreKey(pubKey))

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

	return s.DB.SetSync(DEStoreKey(pubDE), bytes)
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

// DeleteDE deletes the private (d, E) by the given public (D, E)
func (s *Store) DeleteDE(pubDE types.DE) error {
	return s.DB.DeleteSync(DEStoreKey(pubDE))
}
