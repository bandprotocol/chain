package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"strings"
)

// Parameter store keys
var (
	KeyMintDenom             = []byte("MintDenom")
	KeyInflationRateChange   = []byte("InflationRateChange")
	KeyInflationMax          = []byte("InflationMax")
	KeyInflationMin          = []byte("InflationMin")
	KeyGoalBonded            = []byte("GoalBonded")
	KeyBlocksPerYear         = []byte("BlocksPerYear")
	KeyMintAir               = []byte("MintAir")
	KeyEthIntegrationAddress = []byte("EthIntegrationAddress")
	KeyMaxWithdrawalPerTime  = []byte("MaxWithdrawalPerTime")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, inflationRateChange, inflationMax, inflationMin, goalBonded sdk.Dec, MaxWithdrawalPerTime sdk.Coins, blocksPerYear uint64, mintAir bool, ethIntegrationAddress string,
) Params {

	return Params{
		MintDenom:             mintDenom,
		InflationRateChange:   inflationRateChange,
		InflationMax:          inflationMax,
		InflationMin:          inflationMin,
		GoalBonded:            goalBonded,
		BlocksPerYear:         blocksPerYear,
		MintAir:               mintAir,
		EthIntegrationAddress: ethIntegrationAddress,
		MaxWithdrawalPerTime:  MaxWithdrawalPerTime,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:             sdk.DefaultBondDenom,
		InflationRateChange:   sdk.NewDecWithPrec(13, 2),
		InflationMax:          sdk.NewDecWithPrec(20, 2),
		InflationMin:          sdk.NewDecWithPrec(7, 2),
		GoalBonded:            sdk.NewDecWithPrec(67, 2),
		BlocksPerYear:         uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
		MintAir:               false,
		EthIntegrationAddress: "0xa19Df1199CeEfd7831576f1D055E454364337633", // default value (might be invalid for actual use)
		MaxWithdrawalPerTime:  sdk.Coins{},
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationRateChange(p.InflationRateChange); err != nil {
		return err
	}
	if err := validateInflationMax(p.InflationMax); err != nil {
		return err
	}
	if err := validateInflationMin(p.InflationMin); err != nil {
		return err
	}
	if err := validateGoalBonded(p.GoalBonded); err != nil {
		return err
	}
	if err := validateBlocksPerYear(p.BlocksPerYear); err != nil {
		return err
	}
	if err := validateMintAir(p.MintAir); err != nil {
		return err
	}
	if err := validateEthIntegarionAddress(p.EthIntegrationAddress); err != nil {
		return err
	}
	if err := validateMaxWithdrawalPerTime(p.MaxWithdrawalPerTime); err != nil {
		return err
	}
	if p.InflationMax.LT(p.InflationMin) {
		return fmt.Errorf(
			"max inflation (%s) must be greater than or equal to min inflation (%s)",
			p.InflationMax, p.InflationMin,
		)
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:             %s
  Inflation Rate Change:  %s
  Inflation Max:          %s
  Inflation Min:          %s
  Goal Bonded:            %s
  Blocks Per Year:        %d
  Eth Integration Address: %s
  Max Withdrawal Per Time:	%s
`,
		p.MintDenom, p.InflationRateChange, p.InflationMax,
		p.InflationMin, p.GoalBonded, p.BlocksPerYear, p.EthIntegrationAddress, p.MaxWithdrawalPerTime,
	)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyInflationRateChange, &p.InflationRateChange, validateInflationRateChange),
		paramtypes.NewParamSetPair(KeyInflationMax, &p.InflationMax, validateInflationMax),
		paramtypes.NewParamSetPair(KeyInflationMin, &p.InflationMin, validateInflationMin),
		paramtypes.NewParamSetPair(KeyGoalBonded, &p.GoalBonded, validateGoalBonded),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
		paramtypes.NewParamSetPair(KeyMintAir, &p.MintAir, validateMintAir),
		paramtypes.NewParamSetPair(KeyEthIntegrationAddress, &p.EthIntegrationAddress, validateEthIntegarionAddress),
		paramtypes.NewParamSetPair(KeyMaxWithdrawalPerTime, &p.MaxWithdrawalPerTime, validateMaxWithdrawalPerTime),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return sdkerrors.Wrap(ErrInvalidMintDenom, "mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

func validateInflationMax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateInflationMin(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("min inflation too large: %s", v)
	}

	return nil
}

func validateGoalBonded(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("goal bonded cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}

func validateMaxWithdrawalPerTime(i interface{}) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !v.IsValid() {
		return fmt.Errorf("max withdrawal per time parameter is not valid: %s", v)
	}
	if v.IsAnyNegative() {
		return fmt.Errorf("max withdrawal per time cannot be negative: %s", v)
	}

	return nil
}

func validateEthIntegarionAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if !ethcommon.IsHexAddress(v) {
		return fmt.Errorf("value is not a valid eth hex address: %s", v)
	}
	return nil
}

func validateMintAir(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}
