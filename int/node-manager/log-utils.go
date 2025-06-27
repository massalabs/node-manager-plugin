package nodeManager

import (
	"fmt"
	"io"
)

func (nodeMana *NodeManager) getLogger(isMainnet bool) (io.WriteCloser, error) {
	if isMainnet && nodeMana.mainnetLogger != nil {
		return nodeMana.mainnetLogger, nil
	}

	if !isMainnet && nodeMana.buildnetLogger != nil {
		return nodeMana.buildnetLogger, nil
	}

	version, err := nodeMana.nodeDirManager.GetVersion(isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to get node version: %v", err)
	}

	// Set the node logger as the stdout and stderr of the node process
	nodeLogger, err := nodeMana.NodeLogManager.newLogger(version)
	if err != nil {
		return nil, fmt.Errorf("failed to create node logger: %v", err)
	}

	if isMainnet {
		nodeMana.mainnetLogger = nodeLogger
	} else {
		nodeMana.buildnetLogger = nodeLogger
	}

	return nodeLogger, nil
}

func (nodeMana *NodeManager) closeLoggers() error {
	if nodeMana.mainnetLogger != nil {
		if err := nodeMana.mainnetLogger.Close(); err != nil {
			return fmt.Errorf("failed to close mainnet logger: %v", err)
		}
	}

	if nodeMana.buildnetLogger != nil {
		if err := nodeMana.buildnetLogger.Close(); err != nil {
			return fmt.Errorf("failed to close buildnet logger: %v", err)
		}
	}

	return nil
}
