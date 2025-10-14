package nodeManager

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/station/pkg/logger"
	"gopkg.in/natefinch/lumberjack.v2"
)

type NodeLogManager struct {
	config *config.PluginConfig
	re     *regexp.Regexp
}

type logFile struct {
	path      string
	timestamp *time.Time // nil for the current file
}

const (
	NodeLogFileBaseName  = "node"
	NodeLogFileExtension = ".log"
)

func NewNodeLogManager(config *config.PluginConfig) (*NodeLogManager, error) {
	// Exemple : node-2024-06-07T12-34-56.789.log
	re, err := regexp.Compile(regexp.QuoteMeta(NodeLogFileBaseName) + `-(\\d{4}-\\d{2}-\\d{2}T\\d{2}-\\d{2}-\\d{2}\\.\\d{3})` + regexp.QuoteMeta(NodeLogFileExtension))
	if err != nil {
		return nil, err
	}
	nodeLogManager := &NodeLogManager{
		config: config,
		re:     re,
	}
	if err := nodeLogManager.cleanOldVersionsLogs(); err != nil {
		return nil, err
	}
	return nodeLogManager, nil
}

func (nodeLog *NodeLogManager) cleanOldVersionsLogs() error {
	if _, err := os.Stat(nodeLog.config.NodeLogPath); os.IsNotExist(err) {
		return nil
	}

	if err := config.GlobalPluginInfo.RemoveOldNodeVersionsArtifacts(nodeLog.config.NodeLogPath); err != nil {
		return fmt.Errorf("failed to remove logs of old node versions: %v", err)
	}

	return nil
}

func (nodeLog *NodeLogManager) newLogger(logDirName string) (*lumberjack.Logger, error) {
	logFilesFolderPath := filepath.Join(nodeLog.config.NodeLogPath, logDirName)

	// Create the log files folder for the given logDirName if it doesn't exist
	if _, err := os.Stat(logFilesFolderPath); os.IsNotExist(err) {
		if err := os.MkdirAll(logFilesFolderPath, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create log files folder: %v", err)
		}
	}

	return &lumberjack.Logger{
		Filename:   filepath.Join(logFilesFolderPath, NodeLogFileBaseName+NodeLogFileExtension),
		MaxSize:    nodeLog.config.NodeLogMaxSize, // megabytes
		MaxBackups: nodeLog.config.MaxLogBackups,
	}, nil
}

func (nodeLog *NodeLogManager) getLogs(logDirName string) (string, error) {
	logFilesFolderPath := filepath.Join(nodeLog.config.NodeLogPath, logDirName)

	// Check if the log files folder exists
	if _, err := os.Stat(logFilesFolderPath); os.IsNotExist(err) {
		// Create the log files folder for the given logDirName if it doesn't exist
		if err := os.MkdirAll(logFilesFolderPath, 0o755); err != nil {
			logger.Error("Failed to create log files folder: %v", err)
		}
		return "", nil
	}

	// Get all log files in the directory
	files, err := os.ReadDir(logFilesFolderPath)
	if err != nil {
		return "", fmt.Errorf("failed to read log files folder %v: %w", logFilesFolderPath, err)
	}

	// Filter only log files and sort them by modification time
	var logFiles []logFile
	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || !strings.HasPrefix(fileName, NodeLogFileBaseName) || !strings.HasSuffix(fileName, NodeLogFileExtension) {
			continue
		}
		fullPath := filepath.Join(logFilesFolderPath, fileName)
		matches := nodeLog.re.FindStringSubmatch(fileName)
		if len(matches) == 2 {
			// Parse the timestamp
			t, err := time.Parse("2006-01-02T15-04-05.000", matches[1])
			if err == nil {
				logFiles = append(logFiles, logFile{path: fullPath, timestamp: &t})
				continue
			}
		}
		// If no timestamp, it's the current file (the one that is being written to) i.e. node.log
		logFiles = append(logFiles, logFile{path: fullPath, timestamp: nil})
	}

	if len(logFiles) == 0 {
		return "", nil
	}

	// If there's only one file, read it directly
	if len(logFiles) == 1 {
		content, err := os.ReadFile(logFiles[0].path)
		if err != nil {
			return "", fmt.Errorf("failed to read log file %v: %w", logFiles[0].path, err)
		}
		return string(content), nil
	}

	// Sort files by rotation time (oldest first)
	sort.Slice(logFiles, func(i, j int) bool {
		if logFiles[i].timestamp == nil {
			return false
		}
		if logFiles[j].timestamp == nil {
			return true
		}
		return logFiles[i].timestamp.Before(*logFiles[j].timestamp)
	})

	// Create a channel to receive file contents
	type fileContentRead struct {
		content string
		index   int
		err     error
	}
	contentChan := make(chan fileContentRead, len(logFiles))

	// Read files concurrently
	for i, file := range logFiles {
		go func(file logFile, index int) {
			content, err := os.ReadFile(file.path)
			if err != nil {
				contentChan <- fileContentRead{content: "", index: index, err: err}
				return
			}
			contentChan <- fileContentRead{content: string(content), index: index, err: nil}
		}(file, i)
	}

	// Collect results in order
	var errors []error
	results := make([]string, len(logFiles))
	for i := 0; i < len(logFiles); i++ {
		result := <-contentChan
		if result.err != nil {
			errors = append(errors, fmt.Errorf("failed to read log file %v: %w", logFiles[result.index].path, result.err))
			continue
		}
		results[result.index] = result.content
	}

	if len(errors) > 0 {
		return "", fmt.Errorf("failed to read some log files: %v", errors)
	}

	// Concatenate all contents
	return strings.Join(results, ""), nil
}
