package testapp

import (
	"fmt"
	"github.com/GeoDB-Limited/odin-core/pkg/filecache"
	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
	"path/filepath"
)

func getGenesisDataSources(homePath string) []oracletypes.DataSource {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	DataSources = []oracletypes.DataSource{{}} // 0th index should be ignored
	for idx := 0; idx < 5; idx++ {
		id := idx + 1
		idxStr := fmt.Sprintf("%d", id)
		hash := fc.AddFile([]byte("code" + idxStr))
		ds := oracletypes.NewDataSource(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, Coins1000000loki,
		)
		ds.ID = oracletypes.DataSourceID(id)
		DataSources = append(DataSources, ds)
	}
	return DataSources[1:]
}

func getGenesisOracleScripts(homePath string) []oracletypes.OracleScript {
	dir := filepath.Join(homePath, "files")
	fc := filecache.New(dir)
	OracleScripts = []oracletypes.OracleScript{{}} // 0th index should be ignored
	wasms := [][]byte{
		Wasm1, Wasm2, Wasm3, Wasm4, Wasm56(10), Wasm56(10000000), Wasm78(10), Wasm78(2000), Wasm9,
	}
	for idx := 0; idx < len(wasms); idx++ {
		id := idx + 1
		idxStr := fmt.Sprintf("%d", id)
		hash := fc.AddFile(compile(wasms[idx]))
		os := oracletypes.NewOracleScript(
			Owner.Address, "name"+idxStr, "desc"+idxStr, hash, "schema"+idxStr, "url"+idxStr,
		)
		os.ID = oracletypes.OracleScriptID(id)
		OracleScripts = append(OracleScripts, os)
	}
	return OracleScripts[1:]
}
