package oraclekeeper

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"

	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

// HasOracleScript checks if the oracle script of this ID exists in the storage.
func (k Keeper) HasOracleScript(ctx sdk.Context, id oracletypes.OracleScriptID) bool {
	return ctx.KVStore(k.storeKey).Has(oracletypes.OracleScriptStoreKey(id))
}

// GetOracleScript returns the oracle script struct for the given ID or error if not exists.
func (k Keeper) GetOracleScript(ctx sdk.Context, id oracletypes.OracleScriptID) (oracletypes.OracleScript, error) {
	bz := ctx.KVStore(k.storeKey).Get(oracletypes.OracleScriptStoreKey(id))
	if bz == nil {
		return oracletypes.OracleScript{}, sdkerrors.Wrapf(oracletypes.ErrOracleScriptNotFound, "id: %d", id)
	}
	var oracleScript oracletypes.OracleScript
	k.cdc.MustUnmarshalBinaryBare(bz, &oracleScript)
	return oracleScript, nil
}

// GetPaginatedOracleScripts returns oracle scripts with pagination.
func (k Keeper) GetPaginatedOracleScripts(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]oracletypes.OracleScript, *query.PageResponse, error) {
	oracleScripts := make([]oracletypes.OracleScript, 0)
	oracleScriptsStore := prefix.NewStore(ctx.KVStore(k.storeKey), oracletypes.OracleScriptStoreKeyPrefix)

	pageRes, err := query.FilteredPaginate(
		oracleScriptsStore,
		pagination,
		func(key []byte, value []byte, accumulate bool) (bool, error) {
			var oracleScript oracletypes.OracleScript
			if err := k.cdc.UnmarshalBinaryBare(value, &oracleScript); err != nil {
				return false, err
			}
			if accumulate {
				oracleScripts = append(oracleScripts, oracleScript)
			}
			return true, nil
		},
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to paginate oracle scripts")
	}

	return oracleScripts, pageRes, nil
}

// MustGetOracleScript returns the oracle script struct for the given ID. Panic if not exists.
func (k Keeper) MustGetOracleScript(ctx sdk.Context, id oracletypes.OracleScriptID) oracletypes.OracleScript {
	oracleScript, err := k.GetOracleScript(ctx, id)
	if err != nil {
		panic(err)
	}
	return oracleScript
}

// SetOracleScript saves the given oracle script to the storage without performing validation.
func (k Keeper) SetOracleScript(ctx sdk.Context, id oracletypes.OracleScriptID, oracleScript oracletypes.OracleScript) {
	store := ctx.KVStore(k.storeKey)
	store.Set(oracletypes.OracleScriptStoreKey(id), k.cdc.MustMarshalBinaryBare(&oracleScript))
}

// AddOracleScript adds the given oracle script to the storage.
func (k Keeper) AddOracleScript(ctx sdk.Context, oracleScript oracletypes.OracleScript) oracletypes.OracleScriptID {
	id := k.GetNextOracleScriptID(ctx)
	k.SetOracleScript(ctx, id, oracleScript)
	return id
}

// MustEditOracleScript edits the given oracle script by id and flushes it to the storage. Panic if not exists.
func (k Keeper) MustEditOracleScript(ctx sdk.Context, id oracletypes.OracleScriptID, new oracletypes.OracleScript) {
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
func (k Keeper) GetAllOracleScripts(ctx sdk.Context) (oracleScripts []oracletypes.OracleScript) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, oracletypes.OracleScriptStoreKeyPrefix)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var oracleScript oracletypes.OracleScript
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &oracleScript)
		oracleScripts = append(oracleScripts, oracleScript)
	}
	return oracleScripts
}

// AddOracleScriptFile compiles Wasm code (see go-owasm), adds the compiled file to filecache,
// and returns its sha256 reference name. Returns do-not-modify symbol if input is do-not-modify.
func (k Keeper) AddOracleScriptFile(file []byte) (string, error) {
	if bytes.Equal(file, oracletypes.DoNotModifyBytes) {
		return oracletypes.DoNotModify, nil
	}
	compiledFile, err := k.owasmVM.Compile(file, oracletypes.MaxCompiledWasmCodeSize)
	if err != nil {
		return "", sdkerrors.Wrapf(oracletypes.ErrOwasmCompilation, "with error: %s", err.Error())
	}
	return k.fileCache.AddFile(compiledFile), nil
}
