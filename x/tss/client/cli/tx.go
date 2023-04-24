package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/x/tss/types"
)

// NewTxCmd returns a root CLI command handler for all x/tss transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "TSS transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(MsgCreateGroupCmd())
	txCmd.AddCommand(MsgSubmitDKGRound1Cmd())

	return txCmd
}

func MsgCreateGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-group [member1,member2,...] [threshold]",
		Args:  cobra.ExactArgs(2),
		Short: "Make a new group for tss module",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a new group for sign tx in tss module.
Example:
$ %s tx tss create-group band15mxunzureevrg646khnunhrl6nxvrj3eree5tz,band1p2t43jx3rz84y4z05xk8dcjjhzzeqnfrt9ua9v,band18f55l8hf4l7zvy8tx28n4r4nksz79p6lp4z305 2 --from mykey
`,
				version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members := strings.Split(args[0], ",")

			threshold, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			msg := &types.MsgCreateGroup{
				Members:   members,
				Threshold: uint32(threshold),
				Sender:    clientCtx.GetFromAddress().String(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgSubmitDKGRound1Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-dkg-round1 [group_id] [one_time_pub_key] [a0_sing] [one_time_sign] [coefficients-commit-json-file]",
		Args:  cobra.ExactArgs(5),
		Short: "submit tss round 1 containing group_id, one_time_pub_key, a0_sing, one_time_sign and coefficients_commit",
		Example: fmt.Sprintf(`
		%s tx tss submit-dkg-round1 [group_id] [one_time_pub_key] [a0_sing] [one_time_sign] coefficients-commit.json
		
		where coefficients-commit.json contains:
		
		{
			"points": [
			  	{
					"x": "d74bf844b0862475103d96a611cf2d898447e288d34b360bc885cb8ce7c00575",
					"y": "131c670d414c4546b88ac3ff664611b1c38ceb1c21d76369d7a7a0969d61d97d"
				},
				{
					"x": "d74bf844b0862475103d96a611cf2d898447e288d34b360bc885cb8ce7c00575",
					"y": "131c670d414c4546b88ac3ff664611b1c38ceb1c21d76369d7a7a0969d61d97d"
				}
			]
		  }
		`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			oneTimePubKey, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			a0Sign, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			oneTimeSign, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			coefficientsCommit, err := parsePoints(args[4])
			if err != nil {
				return err
			}

			msg := &types.MsgSubmitDKGRound1{
				GroupId:            groupID,
				CoefficientsCommit: coefficientsCommit,
				OneTimePubKey:      oneTimePubKey,
				A0Sing:             a0Sign,
				OneTimeSign:        oneTimeSign,
				Sender:             clientCtx.GetFromAddress().String(),
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
