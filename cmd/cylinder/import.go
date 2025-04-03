package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v3/cylinder/context"
	"github.com/bandprotocol/chain/v3/cylinder/store"
)

// importCmd returns a Cobra command for importing data from store.
func importCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import data in cylinder's store",
	}

	cmd.AddCommand(
		importGroupsCmd(ctx),
		importDKGsCmd(ctx),
		importDEsCmd(ctx),
	)

	return cmd
}

// importGroupsCmd returns a Cobra command for importing groups data
func importGroupsCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups [path_to_json_file]",
		Short: "Import groups data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// open the file
			jsonFile, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			// read the file as bytes
			bytes, err := io.ReadAll(jsonFile)
			if err != nil {
				return err
			}

			// unmarshal json to data
			var groups []store.Group
			err = json.Unmarshal(bytes, &groups)
			if err != nil {
				return err
			}

			// loop to set each group to store
			for _, group := range groups {
				if err = ctx.Store.SetGroup(group); err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}

// importDKGsCmd returns a Cobra command for importing dkgs data
func importDKGsCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dkgs [path_to_json_file]",
		Short: "Import DKGs data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// open the file
			jsonFile, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			// read the file as bytes
			bytes, err := io.ReadAll(jsonFile)
			if err != nil {
				return err
			}

			// unmarshal json to data
			var dkgs []store.DKG
			err = json.Unmarshal(bytes, &dkgs)
			if err != nil {
				return err
			}

			// loop to set each dkg to store
			for _, dkg := range dkgs {
				err = ctx.Store.SetDKG(dkg)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}

// importDEsCmd returns a Cobra command for importing des data
func importDEsCmd(ctx *context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "des [path_to_json_file]",
		Short: "Import DEs data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, err = ctx.WithGoLevelDB()
			if err != nil {
				return err
			}

			// open the file
			jsonFile, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer jsonFile.Close()

			// read the file as bytes
			bytes, err := io.ReadAll(jsonFile)
			if err != nil {
				return err
			}

			// unmarshal json to data
			var des []store.DE
			if err := json.Unmarshal(bytes, &des); err != nil {
				panic(err)
			}

			// loop to set each de to store
			for _, de := range des {
				err = ctx.Store.SetDE(de)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}
