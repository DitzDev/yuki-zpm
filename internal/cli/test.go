package cli

import (
	"github.com/spf13/cobra"
	"yuki/internal/build"
)

func TestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run project tests",
		Long:  "Run all tests with dependencies",
		RunE:  runTest,
	}
}

func runTest(cmd *cobra.Command, args []string) error {
	builder := build.New()
	return builder.Test(".")
}
