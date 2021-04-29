package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config [key] [value]",
		Aliases: []string{"c"},
		Short:   "Set faucet configuration environment",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "coins" {
				viper.Set(args[0], args[1:])
			} else {
				viper.Set(args[0], args[1])
			}
			return viper.WriteConfig()
		},
	}
	return cmd
}
