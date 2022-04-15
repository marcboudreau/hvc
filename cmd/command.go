package cmd

import (
	"github.com/marcboudreau/hvc/cmd/copy"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hvc",
	Short: "hvc efficiently copies secrets from one or more source Vaults to a target Vault",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	rootCmd.AddCommand(copy.CopyCmd)
}

// Execute executes the rootCmd's Run function.
func Execute() {
	rootCmd.Execute()
}
