package cli

import (
	"github.com/spf13/cobra"
	"yuki/internal/build"
)

func RunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Build and run the project",
		Long:  "Compile and execute the project with all dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(args)
		},
	}
}

func runRun(args []string) error {
	builder := build.New()
	return builder.Run(".", args)
}
