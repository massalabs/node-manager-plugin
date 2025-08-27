package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetExecDirPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}

	return filepath.Dir(execPath), nil
}
