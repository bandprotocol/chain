package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/store"
	"github.com/bandprotocol/chain/v3/pkg/tss"
)

const (
	flagOutput = "output"
)

// exportCmd returns a Cobra command for exporting data from store.
func exportCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data in cylinder's store",
	}

	cmd.AddCommand(
		exportGroupsCmd(ctx),
		exportDKGsCmd(ctx),
		exportDEsCmd(ctx),
	)

	return cmd
}

// exportGroupsCmd returns a Cobra command for exporting groups data
func exportGroupsCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups [public-key-1] [public-key-2] [public-key-3] [...]",
		Short: "Export groups data",
		Long:  "Export groups data by its public key, if no public key is provided, all groups will be exported",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// get 'output' flag
			output, err := cmd.Flags().GetString(flagOutput)
			if err != nil {
				return err
			}

			// get groups information
			groups := []store.Group{}
			if len(args) == 0 {
				groups, err = ctx.Store.GetAllGroups()
				if err != nil {
					return err
				}
			} else {
				for i := 0; i < len(args); i++ {
					pubKey, err := hex.DecodeString(args[i])
					if err != nil {
						return err
					}

					group, err := ctx.Store.GetGroup(pubKey)
					if err != nil {
						return err
					}

					groups = append(groups, group)
				}
			}

			// marshal data of groups to json
			bytes, err := json.Marshal(groups)
			if err != nil {
				return err
			}

			// create file
			f, err := os.Create(output)
			if err != nil {
				return err
			}
			defer f.Close()

			// write data to the file
			_, err = f.Write(bytes)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flagOutput, "", "Specific output filename")

	_ = cmd.MarkFlagRequired(flagOutput)

	return cmd
}

// exportDKGsCmd returns a Cobra command for exporting DKGs data
func exportDKGsCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dkgs [group-id-1] [group-id-2] [group-id-3] [...]",
		Short: "Export DKGs data",
		Long:  "Export DKGs data by group id. If not specify any group id, it will export all DKGs data",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// get 'output' flag
			output, err := cmd.Flags().GetString(flagOutput)
			if err != nil {
				return err
			}

			// get DKGs information
			var dkgs []store.DKG
			if len(args) == 0 {
				dkgs, err = ctx.Store.GetAllDKGs()
				if err != nil {
					return err
				}
			} else {
				for i := 0; i < len(args); i++ {
					gid, err := strconv.ParseUint(args[i], 10, 64)
					if err != nil {
						return err
					}

					dkg, err := ctx.Store.GetDKG(tss.GroupID(gid))
					if err != nil {
						return err
					}

					dkgs = append(dkgs, dkg)
				}
			}

			// marshal data of dkgs to json
			bytes, err := json.Marshal(dkgs)
			if err != nil {
				return err
			}

			// create file
			f, err := os.Create(output)
			if err != nil {
				return err
			}
			defer f.Close()

			// write data to the file
			_, err = f.Write(bytes)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flagOutput, "", "Specific output filename")

	_ = cmd.MarkFlagRequired(flagOutput)

	return cmd
}

// exportDEsCmd returns a Cobra command for exporting DEs data
func exportDEsCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "des",
		Short: "Export DEs data",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// get 'output' flag
			output, err := cmd.Flags().GetString(flagOutput)
			if err != nil {
				return err
			}

			// get DEs information
			des, err := ctx.Store.GetAllDEs()
			if err != nil {
				return err
			}

			// marshal data of DEs to json
			bytes, err := json.Marshal(des)
			if err != nil {
				return err
			}

			// create file
			f, err := os.Create(output)
			if err != nil {
				return err
			}
			defer f.Close()

			// write data to the file
			_, err = f.Write(bytes)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flagOutput, "", "Specific output filename")

	_ = cmd.MarkFlagRequired(flagOutput)

	return cmd
}
