package keepassimport

import (
	"github.com/spf13/cobra"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "keepassimport [csvpath]",
		Short: "Imports data from Keepass format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			Run(args[0])
		},
	}
}
