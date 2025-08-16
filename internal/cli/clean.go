package cli

import (
	"github.com/spf13/cobra"
	"yuki/internal/build"
	"yuki/internal/logger"
	"yuki/internal/vendor"
)

func CleanCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean build artifacts",
		Long:  "Remove build artifacts and optionally dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(all)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Also remove dependencies (yuki_modules)")

	return cmd
}

func runClean(all bool) error {
	builder := build.New()
	
	if err := builder.Clean("."); err != nil {
		return err
	}

	if all {
		vendorer := vendor.New()
		if err := vendorer.Clean("."); err != nil {
			return err
		}
		logger.Info("Removed dependencies and generated files")
	}

	return nil
}
