package main

import (
	"os"

	"github.com/spf13/cobra"
	"yuki_zpm.org/cli"
	"yuki_zpm.org/logger"
)

var (
	verbose bool
	quiet   bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "yuki",
		Short: "Yuki - A package manager for Zig",
		Long:  "Yuki is a comprehensive package manager for Zig projects with GitHub integration and semantic versioning support.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.SetVerbose(verbose)
			logger.SetQuiet(quiet)
		},
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Enable quiet mode")

	rootCmd.AddCommand(cli.InitCmd())
	rootCmd.AddCommand(cli.BuildCmd())
	rootCmd.AddCommand(cli.TestCmd())
	rootCmd.AddCommand(cli.RunCmd())
	rootCmd.AddCommand(cli.CheckCmd())
	rootCmd.AddCommand(cli.AddCmd())
	rootCmd.AddCommand(cli.InstallCmd())
	rootCmd.AddCommand(cli.UpdateCmd())
	rootCmd.AddCommand(cli.RemoveCmd())
	rootCmd.AddCommand(cli.SyncCmd())
	rootCmd.AddCommand(cli.OutdatedCmd())
	rootCmd.AddCommand(cli.ListCmd())
	rootCmd.AddCommand(cli.WhyCmd())
	rootCmd.AddCommand(cli.SearchCmd())
	rootCmd.AddCommand(cli.InfoCmd())
	rootCmd.AddCommand(cli.DoctorCmd())
	rootCmd.AddCommand(cli.CacheCmd())
	rootCmd.AddCommand(cli.ConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)	
	}
}
