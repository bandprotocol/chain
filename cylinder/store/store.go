package store

import (
	"encoding/json"

	"github.com/bandprotocol/chain/v2/pkg/tss"
	dbm "github.com/tendermint/tm-db"
)

type Store struct {
	DB dbm.DB
}

func NewStore(db dbm.DB) *Store {
	return &Store{
		DB: db,
	}
}

func (s *Store) SetGroup(groupID tss.GroupID, group Group) error {
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}

	return s.DB.Set(GroupStoreKey(groupID), bytes)
}

func (s *Store) GetGroup(groupID tss.GroupID) (Group, error) {
	bytes, err := s.DB.Get(GroupStoreKey(groupID))

	var group Group
	err = json.Unmarshal(bytes, &group)
	if err != nil {
		return Group{}, err
	}

	return group, err
}

func (s *Store) SetDE(D, E uint64, priv DE) error {
	bytes, err := json.Marshal(priv)
	if err != nil {
		return err
	}

	return s.DB.Set(DEStoreKey(D, E), bytes)
}

func (s *Store) GetDE(D, E uint64) (DE, error) {
	bytes, err := s.DB.Get(DEStoreKey(D, E))

	var de DE
	err = json.Unmarshal(bytes, &de)
	if err != nil {
		return DE{}, err
	}

	return de, err
}

func (s *Store) RemoveDE(D, E uint64) error {
	return s.DB.DeleteSync(DEStoreKey(D, E))
}
