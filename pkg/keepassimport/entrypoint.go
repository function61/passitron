package keepassimport

import (
	"github.com/spf13/cobra"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "keepassimport [csvpath] [userId]",
		Short: "Imports data from Keepass format",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			Run(args[0], args[1])
		},
	}
}
