package keeper

import (
	telemetrytypes "github.com/GeoDB-Limited/odin-core/x/telemetry/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"sort"
	"time"
)

type Keeper struct {
	cdc            codec.BinaryMarshaler
	bankKeeper     bankkeeper.ViewKeeper
	stakingQuerier stakingkeeper.Querier
	txDecoder      sdk.TxDecoder
}

func NewKeeper(
	cdc codec.BinaryMarshaler,
	txDecoder sdk.TxDecoder,
	bk bankkeeper.ViewKeeper,
	sk stakingkeeper.Keeper,
) Keeper {
	return Keeper{
		cdc:        cdc,
		bankKeeper: bk,
		stakingQuerier: stakingkeeper.Querier{
			Keeper: sk,
		},
		txDecoder: txDecoder,
	}
}

func (k Keeper) GetPaginatedBalances(
	ctx sdk.Context,
	denom string,
	desc bool,
	pagination *query.PageRequest,
) ([]banktypes.Balance, uint64) {

	balances := k.bankKeeper.GetAccountsBalances(ctx)

	sort.Slice(balances, func(i, j int) bool {
		if desc {
			return balances[j].GetCoins().AmountOf(denom).LT(balances[i].GetCoins().AmountOf(denom))
		}
		return balances[i].GetCoins().AmountOf(denom).LT(balances[j].GetCoins().AmountOf(denom))
	})

	if pagination.GetOffset() >= uint64(len(balances)) {
		return []banktypes.Balance{}, 0
	}

	maxLimit := pagination.GetLimit()
	if pagination.GetOffset()+pagination.GetLimit() >= uint64(len(balances)) {
		maxLimit = uint64(len(balances)) - pagination.GetOffset()
	}

	return balances[pagination.GetOffset() : pagination.GetOffset()+maxLimit], uint64(len(balances))
}

func (k Keeper) GetBalances(ctx sdk.Context, addrs ...sdk.AccAddress) []banktypes.Balance {
	balances := make([]banktypes.Balance, len(addrs))
	for i, addr := range addrs {
		balances[i] = banktypes.Balance{
			Address: addr.String(),
			Coins:   k.bankKeeper.GetAllBalances(ctx, addr),
		}
	}
	return balances
}

func (k Keeper) GetAvgBlockSizePerDay(startDate, endDate time.Time) ([]telemetrytypes.AverageBlockSizePerDay, error) {
	blocksByDates, err := k.GetBlocksByDates(startDate, endDate)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get the blocks by date")
	}

	avgBlockSizePerDay := make([]telemetrytypes.AverageBlockSizePerDay, 0, len(blocksByDates))

	for key, value := range blocksByDates {
		totalSize := 0
		for _, block := range value {
			totalSize += block.Size()
		}
		averageSize := totalSize / len(value)
		avgBlockSizePerDay = append(avgBlockSizePerDay, telemetrytypes.AverageBlockSizePerDay{
			Date:  key,
			Bytes: uint64(averageSize),
		})
	}

	return avgBlockSizePerDay, nil
}

func (k Keeper) GetAvgBlockTimePerDay(startDate, endDate time.Time) ([]telemetrytypes.AverageBlockTimePerDay, error) {
	blocksByDates, err := k.GetBlocksByDates(startDate, endDate)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get the blocks by date")
	}

	avgBlockTimePerDay := make([]telemetrytypes.AverageBlockTimePerDay, 0, len(blocksByDates))

	for key, value := range blocksByDates {
		blocksCount := int64(len(value))
		commonSeconds := value[blocksCount-1].Header.Time.Unix() - value[0].Header.Time.Unix()
		averageTime := commonSeconds / blocksCount

		avgBlockTimePerDay = append(avgBlockTimePerDay, telemetrytypes.AverageBlockTimePerDay{
			Date:    key,
			Seconds: uint64(averageTime),
		})
	}

	return avgBlockTimePerDay, nil
}

func (k Keeper) GetAvgTxFeePerDay(startDate, endDate time.Time) ([]telemetrytypes.AverageTxFeePerDay, error) {
	blocksByDates, err := k.GetBlocksByDates(startDate, endDate)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get the blocks by date")
	}

	avgTxFeePerDay := make([]telemetrytypes.AverageTxFeePerDay, 0, len(blocksByDates))

	for key, value := range blocksByDates {
		totalFee := sdk.NewCoins()
		totalTxCount := 0
		for _, block := range value {
			for _, tx := range block.Data.Txs {
				decodedTx, err := k.txDecoder(tx)
				if err != nil {
					return nil, sdkerrors.Wrap(err, "failed to decode block transaction")
				}

				feeTx, ok := decodedTx.(sdk.FeeTx)
				if !ok {
					return nil, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "failed to retrieve tx fee")
				}

				txFee := feeTx.GetFee()
				totalFee = totalFee.Add(txFee...)
			}
			totalTxCount += len(block.Data.Txs)
		}

		avgFee := sdk.NewCoins()
		if totalTxCount != 0 {
			avgFee, _ = sdk.NewDecCoinsFromCoins(totalFee...).QuoDec(sdk.NewDec(int64(totalTxCount))).TruncateDecimal()
		}
		avgTxFeePerDay = append(avgTxFeePerDay, telemetrytypes.AverageTxFeePerDay{
			Date: key,
			Fee:  avgFee,
		})
	}

	return avgTxFeePerDay, nil
}

func (k Keeper) GetTxVolumePerDay(startDate, endDate time.Time) ([]telemetrytypes.TxVolumePerDay, error) {
	blocksByDates, err := k.GetBlocksByDates(startDate, endDate)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get the blocks by date")
	}

	dailyTxsVolumes := make([]telemetrytypes.TxVolumePerDay, 0, len(blocksByDates))

	for key, value := range blocksByDates {
		txsVolume := 0
		for _, block := range value {
			txsVolume += len(block.Data.Txs)
		}

		dailyTxsVolumes = append(dailyTxsVolumes, telemetrytypes.TxVolumePerDay{
			Date:   key,
			Volume: uint64(txsVolume),
		})
	}
	return dailyTxsVolumes, nil
}

func (k Keeper) GetValidatorsBlocks(
	ctx sdk.Context,
	startDate, endDate time.Time,
	desc bool,
	pagination *query.PageRequest,
) ([]telemetrytypes.ValidatorsBlocks, uint64, error) {

	blocksByDates, err := k.GetBlocksByDates(startDate, endDate)
	if err != nil {
		return nil, 0, sdkerrors.Wrap(err, "failed to get the blocks by date")
	}

	blocksCount := make(map[string]uint64, pagination.Limit)
	for _, blocks := range blocksByDates {
		for _, block := range blocks {
			validators, err := k.GetBlockValidators(block.Header.Height)
			if err != nil {
				return nil, 0, sdkerrors.Wrap(err, "failed to get block validators")
			}

			for _, val := range validators {
				blocksCount[val.Address.String()]++
			}
		}
	}

	validatorsBlocks := make([]telemetrytypes.ValidatorsBlocks, 0, len(blocksCount))
	totalBondedTokens := k.stakingQuerier.TotalBondedTokens(ctx)

	for addr, blocks := range blocksCount {
		valAddr, err := sdk.ValAddressFromHex(addr)
		if err != nil {
			return nil, 0, sdkerrors.Wrap(err, "failed to retrieve validator address from hex")
		}

		validatorRequest := &stakingtypes.QueryValidatorRequest{
			ValidatorAddr: valAddr.String(),
		}

		validatorResponse, err := k.stakingQuerier.Validator(sdk.WrapSDKContext(ctx), validatorRequest)
		if err != nil {
			return nil, 0, sdkerrors.Wrap(err, "failed to get the validator")
		}

		stakePercentage := sdk.NewDecFromIntWithPrec(
			validatorResponse.Validator.BondedTokens(),
			2,
		).QuoRoundUp(
			sdk.NewDecFromIntWithPrec(
				totalBondedTokens,
				2,
			).Mul(
				sdk.NewDecFromIntWithPrec(sdk.NewInt(100),
					2,
				),
			),
		)

		validatorsBlocks = append(
			validatorsBlocks,
			telemetrytypes.ValidatorsBlocks{
				ValidatorAddress: addr,
				BlocksCount:      blocks,
				StakePercentage:  stakePercentage,
			},
		)
	}

	sort.Slice(validatorsBlocks, func(i, j int) bool {
		if desc {
			return validatorsBlocks[j].BlocksCount < validatorsBlocks[j].BlocksCount
		}
		return validatorsBlocks[i].BlocksCount < validatorsBlocks[j].BlocksCount
	})

	validatorsBlocksLength := uint64(len(validatorsBlocks))

	if pagination.GetOffset() >= validatorsBlocksLength {
		return []telemetrytypes.ValidatorsBlocks{}, 0, nil
	}

	maxLimit := pagination.GetLimit()
	if pagination.GetOffset()+pagination.GetLimit() >= validatorsBlocksLength {
		maxLimit = validatorsBlocksLength - pagination.GetOffset()
	}

	return validatorsBlocks[pagination.GetOffset() : pagination.GetOffset()+maxLimit], validatorsBlocksLength, nil
}
