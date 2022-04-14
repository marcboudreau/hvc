package copy

import (
	"fmt"
	"os"

	"github.com/marcboudreau/hvc"
	"github.com/marcboudreau/hvc/spec"
	"github.com/spf13/cobra"
)

var CopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copies secrets according to copy job specification",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "missing copy job specification filename")
			os.Exit(1)
		}

		file, err := os.Open(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open copy job specification file: %s", err.Error())
		}

		copyJobSpec, err := spec.LoadSpec(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load copy job specification file: %s", err.Error())
		}

		copyJob, err := hvc.NewCopyJob(copyJobSpec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to resolve copy job specification: %s", err.Error())
		}

		errorSlice := copyJob.Execute()
		if errorSlice != nil {
			fmt.Fprintf(os.Stderr, "failed to copy secrets: %s", errorSlice)
		}
	},
}
