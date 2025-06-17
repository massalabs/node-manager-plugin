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

	createTestLogFolder := func(t *testing.T, folderName string) string {
		logDir := filepath.Join(tempDir, folderName)
		err := os.MkdirAll(logDir, 0o755)
		require.NoError(t, err)
		return logDir
	}

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

	// Test newLogger
	t.Run("newLogger", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		version := "test-version"
		lumberjackLogger := logger.newLogger(version)

		// Check if the logger was created with correct settings
		assert.NotNil(t, lumberjackLogger)
		assert.Equal(t, filepath.Join(tempDir, version, NodeLogFileBaseName+NodeLogFileExtension), lumberjackLogger.Filename)
		assert.Equal(t, testConfig.NodeLogMaxSize, lumberjackLogger.MaxSize)
		assert.Equal(t, testConfig.MaxLogBackups, lumberjackLogger.MaxBackups)

		// Check if directory was created
		expectedLogPath := filepath.Join(tempDir, version)
		_, err = os.Stat(expectedLogPath)
		assert.NoError(t, err)

		// Write some test content
		_, err = lumberjackLogger.Write([]byte("test log content\n"))
		assert.NoError(t, err)

		// Check if log file was created
		logFilePath := filepath.Join(expectedLogPath, NodeLogFileBaseName+NodeLogFileExtension)
		_, err = os.Stat(logFilePath)
		assert.NoError(t, err)
	})

	// Test getLogs
	t.Run("getLogs", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		logFolderTest := "test-version"
		logDir := createTestLogFolder(t, logFolderTest)

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
			filePath := filepath.Join(logDir, tf.name)
			err := os.WriteFile(filePath, []byte(tf.content), 0o644)
			require.NoError(t, err)
		}

		// Get logs
		logs, err := logger.getLogs(logFolderTest)
		require.NoError(t, err)

		// Verify the content is concatenated in the correct order
		expectedContent := "Oldest log content\nMiddle log content\nCurrent log content\n"
		assert.Equal(t, expectedContent, logs)
	})

	// Test getLogs with no files
	t.Run("getLogs with no files", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		logFolderTest := "empty-folder"
		createTestLogFolder(t, logFolderTest)

		// Try to get logs from empty directory
		_, err = logger.getLogs(logFolderTest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no log files found")
	})

	// Test getLogs with invalid file names
	t.Run("getLogs with invalid file names", func(t *testing.T) {
		logger, err := NewNodeLogger(testConfig)
		require.NoError(t, err)

		logFolderTest := "invalid-log-content"
		invalidLogDir := createTestLogFolder(t, logFolderTest)

		// Create a file with invalid name
		invalidFilePath := filepath.Join(invalidLogDir, "invalid.log")
		err = os.WriteFile(invalidFilePath, []byte("Invalid content\n"), 0o644)
		require.NoError(t, err)

		// Try to get logs
		_, err = logger.getLogs(logFolderTest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no log files found")
	})
}
