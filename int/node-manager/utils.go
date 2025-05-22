package nodeManager

func IsRunning(nodeStatus NodeStatus) bool {
	return nodeStatus == NodeStatusOn || nodeStatus == NodeStatusBootstrapping || nodeStatus == NodeStatusStopping
}
