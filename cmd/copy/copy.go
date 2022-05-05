package copy

import (
	"fmt"
	"os"

	"github.com/marcboudreau/hvc"
	"github.com/marcboudreau/hvc/spec"
	"github.com/spf13/cobra"
)

// CopyCmd is the cobra.Command that handles the copy option of this
// application.
var CopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copies secrets according to copy job specification",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing copy job specification filename")
		}

		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open copy job specification file %s: %w", args[0], err)
		}

		copyJobSpec, err := spec.LoadSpec(file)
		if err != nil {
			return fmt.Errorf("failed to load copy job specification file %s: %w", file.Name(), err)
		}

		copyJob, err := hvc.NewCopyJob(copyJobSpec)
		if err != nil {
			return fmt.Errorf("failed to resolve copy job specification: %w", err)
		}

		errorSlice := copyJob.Execute()
		if len(errorSlice) > 0 {
			return fmt.Errorf("failed to copy secrets: %s", errorSlice)
		}

		return nil
	},
}
