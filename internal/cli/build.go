package cli

import (
	"github.com/spf13/cobra"
	"yuki/internal/build"
)

func BuildCmd() *cobra.Command {
	var release bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the project",
		Long:  "Compile the project with all dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(release)
		},
	}

	cmd.Flags().BoolVarP(&release, "release", "r", false, "Build in release mode")

	return cmd
}

func runBuild(release bool) error {
	builder := build.New()
	return builder.Build(".", release)
}
