package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// ResolveAndCheckPath resolves the symlinks and checks if the parent directory is within the trusted root
func resolveAndCheckPath(trustedPath string, path BindPath) (string, error) {
	// Clean the path
	cleanedPath := filepath.Clean(path.Path)
	fullPath, err := filepath.Abs(cleanedPath)
	if err != nil {
		return fullPath, errors.New("cannot resolve full path")
	}

	// Verify if the resolved parent directory is within the trusted root
	err = inTrustedRoot(fullPath, trustedPath)
	if err != nil {
		return cleanedPath, errors.New("path is outside of trusted root")
	}

	log.Printf("checking existence of %s path: %s\n", path.Label, fullPath)
	// Check if the directory exists, and create it if not
	err = createDirectoryIfNotExists(fullPath)
	if err != nil {
		return fullPath, err
	}

	if path.Label == "serverfiles" {
		log.Printf("checking existence of minecraft data folder")
		err = createDirectoryIfNotExists(fmt.Sprintf("%s/mcdata", fullPath))
		if err != nil {
			return fullPath, err
		}
	}
	return fullPath, nil
}

func inTrustedRoot(path string, trustedRoot string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.New("error determining absolute path")
	}

	// Ensure the resolved path starts with the trusted root
	if !filepath.HasPrefix(absPath, trustedRoot) {
		return errors.New("path is outside of trusted root")
	}

	return nil
}

// CreateDirectoryIfNotExists checks if the directory exists, and creates it if not
func createDirectoryIfNotExists(path string) error {
	// Check if the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Directory does not exist, create it
		err := os.MkdirAll(path, 0755) // 0755 is the permission mode for the directory
		if err != nil {
			return fmt.Errorf("failed to create directory: %s, error: %v", path, err)
		}
		log.Printf("created directory: %s\n", path)
	} else if err != nil {
		return fmt.Errorf("error checking directory: %s, error: %v", path, err)
	}
	return nil
}

// VerifyPath checks if the path is valid based on the OS
func verifyPath(trustedPath string, path BindPath) (string, error) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		return resolveAndCheckPath(trustedPath, path)
	}

	if runtime.GOOS == "windows" {
		return path.Path, fmt.Errorf("unimplemented")
	}

	return path.Path, fmt.Errorf("runtime not implemented: %s", runtime.GOOS)
}
