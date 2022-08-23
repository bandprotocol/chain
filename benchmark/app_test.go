package benchmark

import (
	bandapp "github.com/bandprotocol/chain/v2/app"
	dbm "github.com/tendermint/tm-db"
)

func setup(db dbm.DB, withGenesis bool, invCheckPeriod uint) *bandapp.BandApp {
	return &bandapp.BandApp{}
}
