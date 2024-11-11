package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "config [key] [value]",
		Aliases: []string{"c"},
		Short:   "Set Grogu's configuration environment",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			viper.Set(args[0], args[1])
			return viper.WriteConfig()
		},
	}
}
