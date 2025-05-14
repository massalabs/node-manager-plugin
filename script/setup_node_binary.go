package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	massaVersion = "MAIN.2.5"
	buildnetConfigURL = "https://raw.githubusercontent.com/massalabs/massa/buildnet/massa-node/base_config/config.toml"
)

func main() {
	// Determine OS-specific download URL
	var downloadURL string
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "arm64" {
			downloadURL = "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_linux_arm64.tar.gz"
		} else {
			downloadURL = "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_linux.tar.gz"
		}
	case "darwin":
		if runtime.GOARCH == "arm64" {
			downloadURL = "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_macos_aarch64.tar.gz"
		} else {
			downloadURL = "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_macos.tar.gz"
		}
	case "windows":
		downloadURL = "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_windows.zip"
	default:
		fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	// Log the detected OS and architecture
	fmt.Printf("Detected OS: %s, Architecture: %s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Downloading from: %s\n", downloadURL)

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "massa-setup-*")
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	// Download and extract Massa
	fmt.Println("Downloading Massa...")
	if err := downloadAndExtract(downloadURL, tempDir); err != nil {
		fmt.Printf("Error downloading/extracting Massa: %v\n", err)
		os.Exit(1)
	}

	// Setup node directories
	nodeMassaDir := filepath.Join(tempDir, "node-massa")
	if err := setupNodeDirectories(nodeMassaDir); err != nil {
		fmt.Printf("Error setting up node directories: %v\n", err)
		os.Exit(1)
	}

	// Move node-massa to current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	targetDir := filepath.Join(currentDir, "node-massa")
	if err := os.Rename(nodeMassaDir, targetDir); err != nil {
		fmt.Printf("Error moving node-massa directory: %v\n", err)
		os.Exit(1)
	}

	// Download buildnet config
	fmt.Println("Downloading buildnet configuration...")
	if err := downloadBuildnetConfig(targetDir); err != nil {
		fmt.Printf("Error downloading buildnet config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Setup completed successfully!")
}

func downloadAndExtract(url, tempDir string) error {
	// Download file
	resp, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("error downloading file: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create a temporary file for the download
	tempFile := filepath.Join(tempDir, "massa.tar.gz")
	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("error creating temp file: %v", err)
	}
	
	defer out.Close()

	// Copy the downloaded content to the file
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("error saving downloaded file: %v", err)
	}

	// Close the file before extracting
	out.Close()

	// Extract the tar.gz file
	if err := extractTarGz(tempFile, tempDir); err != nil {
		return fmt.Errorf("error extracting file: %v", err)
	}

	// Find the extracted directory
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("error reading temp directory: %v", err)
	}

	// log entries
	for _, entry := range entries {
		fmt.Println(entry.Name())
	}

	var extractedDir string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "massa") {
			extractedDir = filepath.Join(tempDir, entry.Name())
			break
		}
	}

	if extractedDir == "" {
		return fmt.Errorf("could not find extracted massa directory")
	}

	// Rename the extracted directory
	nodeMassaDir := filepath.Join(tempDir, "node-massa")
	if err := os.Rename(extractedDir, nodeMassaDir); err != nil {
		return fmt.Errorf("error renaming directory: %v", err)
	}

	return nil
}

func extractTarGz(tarGzPath, destDir string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func setupNodeDirectories(nodeMassaDir string) error {
	// Remove massa-client
	clientDir := filepath.Join(nodeMassaDir, "massa-client")
	if err := os.RemoveAll(clientDir); err != nil {
		return fmt.Errorf("error removing massa-client: %v", err)
	}

	// Create mainnet and buildnet directories
	nodeDir := filepath.Join(nodeMassaDir, "massa-node")
	mainnetDir := filepath.Join(nodeMassaDir, "mainnet")
	buildnetDir := filepath.Join(nodeMassaDir, "buildnet")

	// Copy massa-node to buildnet
	if err := copyDir(nodeDir, buildnetDir); err != nil {
		return fmt.Errorf("error copying to buildnet: %v", err)
	}

	// rename nodeDir to mainnet
	if err := os.Rename(nodeDir, mainnetDir); err != nil {
		return fmt.Errorf("error renaming nodeDir to mainnet: %v", err)
	}

	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func downloadBuildnetConfig(nodeMassaDir string) error {
	resp, err := http.Get(buildnetConfigURL)
	if err != nil {
		return fmt.Errorf("error downloading buildnet config: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	configPath := filepath.Join(nodeMassaDir, "buildnet", "base_config", "config.toml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	out, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("error creating config file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
