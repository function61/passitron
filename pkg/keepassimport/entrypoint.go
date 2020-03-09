package keepassimport

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func Entrypoint() *cobra.Command {
	return &cobra.Command{
		Use:   "keepassimport [csvpath] [userId]",
		Short: "Imports data from Keepass format",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			exitIfError(Run(args[0], args[1]))
		},
	}
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
