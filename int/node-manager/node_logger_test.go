package nodeManager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeLogger(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test configuration
	testConfig := config.PluginConfig{
		NodeLogPath:    tempDir,
		NodeLogMaxSize: 10,
		MaxLogBackups:  5,
	}

	// Test NewNodeLogger
	t.Run("NewNodeLogger", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)
		assert.NotNil(t, logger)
		assert.NotNil(t, logger.re)
	})

	// Test Init
	t.Run("Init", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		version := "test-version"
		logger.Init(version)

		expectedLogPath := filepath.Join(tempDir, version)
		assert.Equal(t, expectedLogPath, logger.logFilesFolder)

		// Check if directory was created
		_, err = os.Stat(expectedLogPath)
		assert.NoError(t, err)

		// write in the current log file
		_, err = logger.getLogger().Write([]byte("test"))
		assert.NoError(t, err)

		// Check if current log file was created
		currentLogPath := filepath.Join(expectedLogPath, NodeLogFileBaseName+NodeLogFileExtension)
		_, err = os.Stat(currentLogPath)
		assert.NoError(t, err)
	})

	// Test getLogs
	t.Run("getLogs", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		version := "test-version"
		logger.Init(version)

		// Create test log files with different timestamps
		testFiles := []struct {
			name    string
			content string
		}{
			{
				name:    "node-2024-01-01T10-00-00.000.log",
				content: "Oldest log content\n",
			},
			{
				name:    "node-2024-01-02T10-00-00.000.log",
				content: "Middle log content\n",
			},
			{
				name:    "node.log",
				content: "Current log content\n",
			},
		}

		// Create test files
		for _, tf := range testFiles {
			filePath := filepath.Join(logger.logFilesFolder, tf.name)
			// 0644 represents read/write for owner, read-only for group and others
			err := os.WriteFile(filePath, []byte(tf.content), 0o644)
			require.NoError(t, err)
		}

		// Get logs
		logs, err := logger.getLogs()
		require.NoError(t, err)

		// Verify the content is concatenated in the correct order
		expectedContent := "Oldest log content\nMiddle log content\nCurrent log content\n"
		assert.Equal(t, expectedContent, logs)
	})

	// Test getLogs with no files
	t.Run("getLogs with no files", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		version := "empty-version"
		logger.Init(version)

		// Try to get logs from empty directory
		_, err = logger.getLogs()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no log files found")
	})

	// Test getLogs with invalid file names
	t.Run("getLogs with invalid file names", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		version := "invalid-version"
		logger.Init(version)

		// Create a file with invalid name
		invalidFilePath := filepath.Join(logger.logFilesFolder, "invalid.log")
		err = os.WriteFile(invalidFilePath, []byte("Invalid content\n"), 0o644)
		require.NoError(t, err)

		// Try to get logs
		_, err = logger.getLogs()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no log files found")
	})
}
