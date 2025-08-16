package integrity

import (
        "crypto/sha256"
        "fmt"
        "io"
        "os"
        "path/filepath"
        "sort"
        "strings"
)

func CalculateFileChecksum(filePath string) (string, error) {
        file, err := os.Open(filePath)
        if err != nil {
                return "", err
        }
        defer file.Close()

        hash := sha256.New()
        if _, err := io.Copy(hash, file); err != nil {
                return "", err
        }

        return fmt.Sprintf("%x", hash.Sum(nil)), nil
}


func CalculateDirectoryChecksum(dirPath string) (string, error) {
        var files []string
        
        
        err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
                if err != nil {
                        return err
                }
                
                
                if info.IsDir() || filepath.Base(path) == ".git" {
                        return nil
                }
                
                
                if contains(strings.Split(path, string(filepath.Separator)), ".git") {
                        return nil
                }
                
                files = append(files, path)
                return nil
        })
        
        if err != nil {
                return "", err
        }
        
        
        sort.Strings(files)
        
        hash := sha256.New()
        
        
        for _, filePath := range files {
                relPath, err := filepath.Rel(dirPath, filePath)
                if err != nil {
                        return "", err
                }
                
                
                hash.Write([]byte(relPath))
                hash.Write([]byte{0}) 
                
                
                fileHash, err := CalculateFileChecksum(filePath)
                if err != nil {
                        return "", fmt.Errorf("failed to hash file %s: %w", filePath, err)
                }
                
                hash.Write([]byte(fileHash))
                hash.Write([]byte{0}) 
        }
        
        return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func contains(slice []string, item string) bool {
        for _, s := range slice {
                if s == item {
                        return true
                }
        }
        return false
}


func VerifyChecksum(path, expectedChecksum string) error {
        info, err := os.Stat(path)
        if err != nil {
                return fmt.Errorf("failed to stat path: %w", err)
        }
        
        var actualChecksum string
        if info.IsDir() {
                actualChecksum, err = CalculateDirectoryChecksum(path)
        } else {
                actualChecksum, err = CalculateFileChecksum(path)
        }
        
        if err != nil {
                return fmt.Errorf("failed to calculate checksum: %w", err)
        }
        
        if actualChecksum != expectedChecksum {
                return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
        }
        
        return nil
}
