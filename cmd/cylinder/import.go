package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/bandprotocol/chain/v2/cylinder"
	"github.com/bandprotocol/chain/v2/cylinder/store"
)

// importCmd returns a Cobra command for importing data from store.
func importCmd(ctx *Context) *cobra.Command {
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
func importGroupsCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups [path_to_json_file]",
		Short: "Import groups data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			json.Unmarshal(bytes, &groups)

			// create context
			c, err := cylinder.NewContext(ctx.config, ctx.keyring, ctx.home)
			if err != nil {
				return err
			}

			// loop to set each group to store
			for _, group := range groups {
				err = c.Store.SetGroup(group)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}

// importDKGsCmd returns a Cobra command for importing dkgs data
func importDKGsCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dkgs [path_to_json_file]",
		Short: "Import DKGs data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			json.Unmarshal(bytes, &dkgs)

			// create context
			c, err := cylinder.NewContext(ctx.config, ctx.keyring, ctx.home)
			if err != nil {
				return err
			}

			// loop to set each dkg to store
			for _, dkg := range dkgs {
				err = c.Store.SetDKG(dkg)
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
func importDEsCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "des [path_to_json_file]",
		Short: "Import DEs data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			json.Unmarshal(bytes, &des)

			// create context
			c, err := cylinder.NewContext(ctx.config, ctx.keyring, ctx.home)
			if err != nil {
				return err
			}

			// loop to set each de to store
			for _, de := range des {
				err = c.Store.SetDE(de)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	return cmd
}
