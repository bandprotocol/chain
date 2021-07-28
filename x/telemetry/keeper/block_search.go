package keeper

import (
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/rpc/core"
	rpctypes "github.com/tendermint/tendermint/rpc/jsonrpc/types"
	tendermint "github.com/tendermint/tendermint/types"
	"time"
)

const (
	MaxCountPerPage = 100

	AllBlocksQuery = "block.height > 1"
	OrderByAsc     = "asc"
)

func (k Keeper) GetBlocksByDates(startDate, endDate time.Time) (map[time.Time][]*tendermint.Block, error) {
	ok := startDate.Before(endDate)
	if !ok {
		return nil, sdkerrors.Wrapf(telemetrytypes.ErrInvalidDateInterval, "invalid dates order")
	}

	blocksPerDay := make(map[time.Time][]*tendermint.Block, 0)
	maxBlocksPerPage := MaxCountPerPage
	page := 1
	blocksParsed := 0

	for {
		blocks, err := core.BlockSearch(
			&rpctypes.Context{},
			AllBlocksQuery,
			&page,
			&maxBlocksPerPage,
			OrderByAsc,
		)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to find the blocks")
		}

		blocksCount := len(blocks.Blocks)

		for _, r := range blocks.Blocks {
			blockDate := telemetrytypes.TimeToUTCDate(r.Block.Header.Time)
			if blockDate.Equal(startDate) || blockDate.After(startDate) && blockDate.Equal(endDate) || blockDate.Before(endDate) {
				blocksPerDay[blockDate] = append(blocksPerDay[blockDate], r.Block)
			}
		}

		blocksParsed += blocksCount

		if blocks.TotalCount == blocksCount {
			break
		}

		page++
	}

	return blocksPerDay, nil
}

func (k Keeper) GetBlockValidators(blockHeight int64) ([]tendermint.Validator, error) {
	var validators []tendermint.Validator
	maxValidatorsPerPage := MaxCountPerPage
	page := 1

	for {
		resultValidators, err := core.Validators(&rpctypes.Context{}, &blockHeight, &page, &maxValidatorsPerPage)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "failed to get the validators")
		}

		for _, val := range resultValidators.Validators {
			validators = append(validators, *val)
		}

		if resultValidators.Total == len(validators) {
			break
		}

		page++
	}

	return validators, nil
}
