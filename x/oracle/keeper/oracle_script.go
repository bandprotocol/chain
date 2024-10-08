package keeper

import (
	"bytes"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bandprotocol/chain/v3/x/oracle/types"
)

// HasOracleScript checks if the oracle script of this ID exists in the storage.
func (k Keeper) HasOracleScript(ctx sdk.Context, id types.OracleScriptID) bool {
	return ctx.KVStore(k.storeKey).Has(types.OracleScriptStoreKey(id))
}

// GetOracleScript returns the oracle script struct for the given ID or error if not exists.
func (k Keeper) GetOracleScript(ctx sdk.Context, id types.OracleScriptID) (types.OracleScript, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.OracleScriptStoreKey(id))
	if bz == nil {
		return types.OracleScript{}, types.ErrOracleScriptNotFound.Wrapf("id: %d", id)
	}
	var oracleScript types.OracleScript
	k.cdc.MustUnmarshal(bz, &oracleScript)
	return oracleScript, nil
}

// MustGetOracleScript returns the oracle script struct for the given ID. Panic if not exists.
func (k Keeper) MustGetOracleScript(ctx sdk.Context, id types.OracleScriptID) types.OracleScript {
	oracleScript, err := k.GetOracleScript(ctx, id)
	if err != nil {
		panic(err)
	}
	return oracleScript
}

// SetOracleScript saves the given oracle script to the storage without performing validation.
func (k Keeper) SetOracleScript(ctx sdk.Context, id types.OracleScriptID, oracleScript types.OracleScript) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.OracleScriptStoreKey(id), k.cdc.MustMarshal(&oracleScript))
}

// AddOracleScript adds the given oracle script to the storage.
func (k Keeper) AddOracleScript(ctx sdk.Context, oracleScript types.OracleScript) types.OracleScriptID {
	id := k.GetNextOracleScriptID(ctx)
	k.SetOracleScript(ctx, id, oracleScript)
	return id
}

// MustEditOracleScript edits the given oracle script by id and flushes it to the storage. Panic if not exists.
func (k Keeper) MustEditOracleScript(ctx sdk.Context, id types.OracleScriptID, new types.OracleScript) {
	oracleScript := k.MustGetOracleScript(ctx, id)
	oracleScript.Owner = new.Owner
	oracleScript.Name = modify(oracleScript.Name, new.Name)
	oracleScript.Description = modify(oracleScript.Description, new.Description)
	oracleScript.Filename = modify(oracleScript.Filename, new.Filename)
	oracleScript.Schema = modify(oracleScript.Schema, new.Schema)
	oracleScript.SourceCodeURL = modify(oracleScript.SourceCodeURL, new.SourceCodeURL)
	k.SetOracleScript(ctx, id, oracleScript)
}

// GetAllOracleScripts returns the list of all oracle scripts in the store, or nil if there is none.
func (k Keeper) GetAllOracleScripts(ctx sdk.Context) (oracleScripts []types.OracleScript) {
	store := ctx.KVStore(k.storeKey)
	iterator := storetypes.KVStorePrefixIterator(store, types.OracleScriptStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var oracleScript types.OracleScript
		k.cdc.MustUnmarshal(iterator.Value(), &oracleScript)
		oracleScripts = append(oracleScripts, oracleScript)
	}
	return oracleScripts
}

// AddOracleScriptFile compiles Wasm code (see go-owasm), adds the compiled file to filecache,
// and returns its sha256 reference name. Returns do-not-modify symbol if input is do-not-modify.
func (k Keeper) AddOracleScriptFile(file []byte) (string, error) {
	if bytes.Equal(file, types.DoNotModifyBytes) {
		return types.DoNotModify, nil
	}
	compiledFile, err := k.owasmVM.Compile(file, types.MaxCompiledWasmCodeSize)
	if err != nil {
		return "", types.ErrOwasmCompilation.Wrapf("caused by %s", err.Error())
	}
	return k.fileCache.AddFile(compiledFile), nil
}
