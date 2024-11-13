package store

import (
	"encoding/json"
	"fmt"

	dbm "github.com/cometbft/cometbft-db"

	storetypes "cosmossdk.io/store/types"

	"github.com/bandprotocol/chain/v3/pkg/tss"
	"github.com/bandprotocol/chain/v3/x/tss/types"
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
func (s *Store) SetDKG(dkg DKG) error {
	bytes, err := json.Marshal(dkg)
	if err != nil {
		return err
	}

	return s.DB.Set(DKGStoreKey(dkg.GroupID), bytes)
}

// GetAllDKGs retrieves all DKGs information
func (s *Store) GetAllDKGs() ([]DKG, error) {
	iterator, err := s.DB.Iterator(DKGStoreKeyPrefix, storetypes.PrefixEndBytes(DKGStoreKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var dkgs []DKG
	for ; iterator.Valid(); iterator.Next() {
		var dkg DKG
		err = json.Unmarshal(iterator.Value(), &dkg)
		if err != nil {
			return nil, err
		}

		dkgs = append(dkgs, dkg)
	}

	return dkgs, err
}

// GetDKG retrieves the dkg information by the given group id.
func (s *Store) GetDKG(groupID tss.GroupID) (DKG, error) {
	bytes, err := s.DB.Get(DKGStoreKey(groupID))
	if err != nil {
		return DKG{}, err
	}

	if bytes == nil {
		return DKG{}, fmt.Errorf("DKG with group ID (%d) doesn't exist", groupID)
	}

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

// SetGroup stores the group information
func (s *Store) SetGroup(group Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}

	return s.DB.Set(GroupStoreKey(group.GroupPubKey), bytes)
}

// GetAllGroups retrieves all groups information
func (s *Store) GetAllGroups() ([]Group, error) {
	iterator, err := s.DB.Iterator(GroupStoreKeyPrefix, storetypes.PrefixEndBytes(GroupStoreKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var groups []Group
	for ; iterator.Valid(); iterator.Next() {
		var group Group
		err = json.Unmarshal(iterator.Value(), &group)
		if err != nil {
			return nil, err
		}

		groups = append(groups, group)
	}

	return groups, err
}

// GetGroup retrieves the group information by the given public key.
func (s *Store) GetGroup(pubKey tss.Point) (Group, error) {
	bytes, err := s.DB.Get(GroupStoreKey(pubKey))
	if err != nil {
		return Group{}, err
	}

	if bytes == nil {
		return Group{}, fmt.Errorf("group with public key (%s) doesn't exist", pubKey)
	}

	var group Group
	err = json.Unmarshal(bytes, &group)
	if err != nil {
		return Group{}, err
	}

	return group, err
}

// SetDE stores the private (d, E)
func (s *Store) SetDE(privDE DE) error {
	bytes, err := json.Marshal(privDE)
	if err != nil {
		return err
	}

	return s.DB.SetSync(DEStoreKey(privDE.PubDE), bytes)
}

// GetAllDEs retrieves all DEs information
func (s *Store) GetAllDEs() ([]DE, error) {
	iterator, err := s.DB.Iterator(DEStoreKeyPrefix, storetypes.PrefixEndBytes(DEStoreKeyPrefix))
	if err != nil {
		return nil, err
	}
	defer iterator.Close()

	var des []DE
	for ; iterator.Valid(); iterator.Next() {
		var de DE
		err = json.Unmarshal(iterator.Value(), &de)
		if err != nil {
			return nil, err
		}

		des = append(des, de)
	}

	return des, err
}

// GetDE retrieves the private (d, E) by the given public (D, E)
func (s *Store) GetDE(pubDE types.DE) (DE, error) {
	bytes, err := s.DB.Get(DEStoreKey(pubDE))
	if err != nil {
		return DE{}, err
	}

	if bytes == nil {
		return DE{}, fmt.Errorf("DE with public DE (%s) doesn't exist", pubDE)
	}

	var de DE
	err = json.Unmarshal(bytes, &de)
	if err != nil {
		return DE{}, err
	}

	return de, err
}

func (s *Store) HasDE(pubDE types.DE) bool {
	bytes, err := s.DB.Get(DEStoreKey(pubDE))
	return err == nil && bytes != nil
}

// DeleteDE deletes the private (d, E) by the given public (D, E)
func (s *Store) DeleteDE(pubDE types.DE) error {
	return s.DB.DeleteSync(DEStoreKey(pubDE))
}
