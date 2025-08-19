package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
	"yuki_zpm.org/utils"
)

func InitCmd() *cobra.Command {
	var autoYes bool
	
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Yuki project",
		Long:  "Create a new Yuki project with manifest file, build script, and source files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, autoYes)
		},
	}

	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Skip prompts and use default values")
	
	return cmd
}

func runInit(cmd *cobra.Command, args []string, autoYes bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if manifest.Exists(cwd) {
		logger.Error("Project already initialized (yuki.toml exists)")
		return fmt.Errorf("project already initialized")
	}

	logger.Info("Initializing new Yuki project...")

	projectInfo, err := promptProjectInfo(cwd, autoYes)
	if err != nil {
		return err
	}

	scripts := map[string]string{
		"test":   fmt.Sprintf("zig test %s", projectInfo.RootFile),
		"format": "zig fmt src/",
	}

	m := &manifest.Manifest{
		Package: *projectInfo,
		Features: map[string][]string{
			"default": {},
		},
		Scripts: scripts,
	}

	if err := m.Save(cwd); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	if err := utils.CreateProjectStructure(cwd, projectInfo.Name, projectInfo.RootFile); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	logger.Success("Successfully initialized project '%s'", projectInfo.Name)
	logger.Info("Next steps:")
	logger.Info("  - Edit %s to implement your project", projectInfo.RootFile)
	logger.Info("  - Add dependencies with: yuki add <package>@<version>")
	logger.Info("  - Build your project with: yuki build")
	
	return nil
}

func promptProjectInfo(cwd string, autoYes bool) (*manifest.PackageInfo, error) {
	reader := bufio.NewReader(os.Stdin)
	defaultName := filepath.Base(cwd)
	
	detectedZigVersion := utils.DetectZigVersion()
	defaultZigVersion := "0.12.0"
	if detectedZigVersion != "" {
		defaultZigVersion = detectedZigVersion
	}

	var name, version, description, author, license, username, zigVersion, rootFile string

	if autoYes {
		name = defaultName
		version = "0.1.0"
		description = ""
		author = ""
		license = "MIT"
		username = ""
		zigVersion = defaultZigVersion
		rootFile = "src/main.zig"
		
		logger.Info("Using auto-init mode with defaults:")
		logger.Info("  Package name: %s", name)
		logger.Info("  Version: %s", version)
		logger.Info("  License: %s", license)
		logger.Info("  Zig version: %s", zigVersion)
		logger.Info("  Root file: %s", rootFile)
	} else {
		fmt.Printf("Package name (%s): ", defaultName)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		name = strings.TrimSpace(input)
		if name == "" {
			name = defaultName
		}

		fmt.Print("Version (0.1.0): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		version = strings.TrimSpace(input)
		if version == "" {
			version = "0.1.0"
		}

		fmt.Print("Description: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		description = strings.TrimSpace(input)

		fmt.Print("Author (Your name <youremail@example.com>): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		author = strings.TrimSpace(input)

		fmt.Print("License (MIT): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		license = strings.TrimSpace(input)
		if license == "" {
			license = "MIT"
		}

		fmt.Print("GitHub username: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		username = strings.TrimSpace(input)

		if detectedZigVersion != "" {
			fmt.Printf("Minimum Zig version (detected: %s): ", detectedZigVersion)
		} else {
			fmt.Printf("Minimum Zig version (%s): ", defaultZigVersion)
		}
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		zigVersion = strings.TrimSpace(input)
		if zigVersion == "" {
			zigVersion = defaultZigVersion
		}

		fmt.Print("Root source file (src/main.zig): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		rootFile = strings.TrimSpace(input)
		if rootFile == "" {
			rootFile = "src/main.zig"
		}

		rootFile, err = utils.ValidateAndSanitizeRootFile(rootFile)
		if err != nil {
			return nil, fmt.Errorf("invalid root file: %w", err)
		}
	}

	var authors []string
	if author != "" {
		authors = []string{author}
	}

	var homepage, repository string
	if username != "" {
		homepage = fmt.Sprintf("https://github.com/%s/%s", username, name)
		repository = fmt.Sprintf("https://github.com/%s/%s", username, name)
	}

	return &manifest.PackageInfo{
		Name:        name,
		Version:     version,
		Authors:     authors,
		License:     license,
		Description: description,
		Homepage:    homepage,
		Repository:  repository,
		ZigVersion:  zigVersion,
		RootFile:    rootFile,
	}, nil
}
