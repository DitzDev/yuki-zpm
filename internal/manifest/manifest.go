package manifest

import (
        "fmt"
        "os"
        "path/filepath"

        "github.com/BurntSushi/toml"
)

type Manifest struct {
        Package      PackageInfo            `toml:"package"`
        Dependencies map[string]Dependency  `toml:"dependencies,omitempty"`
        DevDeps      map[string]Dependency  `toml:"dev-dependencies,omitempty"`
        BuildDeps    map[string]Dependency  `toml:"build-dependencies,omitempty"`
        Features     map[string][]string    `toml:"features,omitempty"`
        Scripts      map[string]string      `toml:"scripts,omitempty"`
}

type PackageInfo struct {
        Name        string   `toml:"name"`
        Version     string   `toml:"version"`
        Edition     string   `toml:"edition"`
        Authors     []string `toml:"authors"`
        License     string   `toml:"license"`
        Description string   `toml:"description"`
        Homepage    string   `toml:"homepage,omitempty"`
        Repository  string   `toml:"repository,omitempty"`
        Keywords    []string `toml:"keywords,omitempty"`
        ZigVersion  string   `toml:"zig_version"`
        RootFile string `toml:"root_file,omitempty"`
}

type Dependency struct {
        Git     string `toml:"git"`
        Version string `toml:"version"`
        Branch  string `toml:"branch,omitempty"`
        Tag     string `toml:"tag,omitempty"`
        Rev     string `toml:"rev,omitempty"`
        RootFile string `toml:"root_file,omitempty"`
}

type LockFile struct {
        Metadata LockMetadata             `toml:"metadata"`
        Package  []LockedPackage          `toml:"package"`
}

type LockMetadata struct {
        Version string `toml:"version"`
}

type LockedPackage struct {
        Name     string `toml:"name"`
        Version  string `toml:"version"`
        Source   string `toml:"source"`
        Checksum string `toml:"checksum"`
        Deps     []string `toml:"dependencies,omitempty"`
}

const ManifestFile = "yuki.toml"
const LockFileName = "yuki.lock"


func Load(path string) (*Manifest, error) {
        manifestPath := filepath.Join(path, ManifestFile)
        
        data, err := os.ReadFile(manifestPath)
        if err != nil {
                return nil, fmt.Errorf("failed to read manifest file: %w", err)
        }

        var manifest Manifest
        if err := toml.Unmarshal(data, &manifest); err != nil {
                return nil, fmt.Errorf("failed to parse manifest file: %w", err)
        }

        return &manifest, nil
}


func (m *Manifest) Save(path string) error {
        manifestPath := filepath.Join(path, ManifestFile)
        
        file, err := os.Create(manifestPath)
        if err != nil {
                return fmt.Errorf("failed to create manifest file: %w", err)
        }
        defer file.Close()

        encoder := toml.NewEncoder(file)
        if err := encoder.Encode(m); err != nil {
                return fmt.Errorf("failed to encode manifest: %w", err)
        }

        return nil
}


func LoadLockFile(path string) (*LockFile, error) {
        lockPath := filepath.Join(path, LockFileName)
        
        data, err := os.ReadFile(lockPath)
        if err != nil {
                if os.IsNotExist(err) {
                        return &LockFile{
                                Metadata: LockMetadata{Version: "1"},
                                Package:  []LockedPackage{},
                        }, nil
                }
                return nil, fmt.Errorf("failed to read lock file: %w", err)
        }

        var lockFile LockFile
        if err := toml.Unmarshal(data, &lockFile); err != nil {
                return nil, fmt.Errorf("failed to parse lock file: %w", err)
        }

        return &lockFile, nil
}


func (l *LockFile) Save(path string) error {
        lockPath := filepath.Join(path, LockFileName)
        
        file, err := os.Create(lockPath)
        if err != nil {
                return fmt.Errorf("failed to create lock file: %w", err)
        }
        defer file.Close()

        encoder := toml.NewEncoder(file)
        if err := encoder.Encode(l); err != nil {
                return fmt.Errorf("failed to encode lock file: %w", err)
        }

        return nil
}


func (m *Manifest) Validate() error {
        if m.Package.Name == "" {
                return fmt.Errorf("package name is required")
        }
        if m.Package.Version == "" {
                return fmt.Errorf("package version is required")
        }
        if m.Package.ZigVersion == "" {
                return fmt.Errorf("zig_version is required")
        }

        
        for name, dep := range m.Dependencies {
                if err := validateDependency(name, dep); err != nil {
                        return err
                }
        }

        return nil
}

func validateDependency(name string, dep Dependency) error {
        if dep.Git == "" {
                return fmt.Errorf("dependency '%s' must have a git URL", name)
        }
        
        
        hasVersionSpec := dep.Version != "" || dep.Branch != "" || dep.Tag != "" || dep.Rev != ""
        if !hasVersionSpec {
                return fmt.Errorf("dependency '%s' must specify version, branch, tag, or rev", name)
        }

        return nil
}


func (m *Manifest) GetAllDependencies() map[string]Dependency {
        all := make(map[string]Dependency)
        
        for name, dep := range m.Dependencies {
                all[name] = dep
        }
        for name, dep := range m.DevDeps {
                all[name] = dep
        }
        for name, dep := range m.BuildDeps {
                all[name] = dep
        }
        
        return all
}


func Exists(path string) bool {
        manifestPath := filepath.Join(path, ManifestFile)
        _, err := os.Stat(manifestPath)
        return !os.IsNotExist(err)
}
