package nodeStatus

import (
	"sync"

	"github.com/massalabs/station/pkg/logger"
)

type NodeStatusSubscriber struct {
	ch   chan NodeStatus
	name string
}

type NodeStatusDispatcher interface {
	Publish(status NodeStatus)
	GetCurrentStatus() NodeStatus
	SubscribeAll(subscriberName string) (chan NodeStatus, func())
	Subscribe(status []NodeStatus, subscriberName string) (chan NodeStatus, func())
}

type NodeStatusDispatcherImpl struct {
	status                    NodeStatus
	specificStatusSubscribers map[NodeStatus][]NodeStatusSubscriber
	allStatusSubscribers      []NodeStatusSubscriber
	mu                        sync.RWMutex
}

func NewNodeStatusDispatcher() NodeStatusDispatcher {
	return &NodeStatusDispatcherImpl{
		status:                    NodeStatusOff, // Default status
		specificStatusSubscribers: make(map[NodeStatus][]NodeStatusSubscriber),
		allStatusSubscribers:      make([]NodeStatusSubscriber, 0),
	}
}

func (n *NodeStatusDispatcherImpl) Publish(status NodeStatus) {
	n.mu.Lock()
	n.status = status
	n.mu.Unlock()

	n.mu.RLock()
	defer n.mu.RUnlock()

	// Notify specific subscribers for this status
	if subscribers, exists := n.specificStatusSubscribers[status]; exists {
		for _, subscriber := range subscribers {
			select {
			case subscriber.ch <- status:
			default:
				logger.Debugf("Subscriber %s Channel is full or closed, ignoring status update for %s", subscriber.name, status)
				// Channel is full or closed, ignore
			}
		}
	}

	// Notify all subscribers
	for _, subscriber := range n.allStatusSubscribers {
		select {
		case subscriber.ch <- status:
		default:
			logger.Debugf("Subscriber %s Channel is full or closed for all status changes, ignoring status update for %s", subscriber.name, status)
			// Channel is full or closed, ignore
		}
	}
}

func (n *NodeStatusDispatcherImpl) GetCurrentStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.status
}

func (n *NodeStatusDispatcherImpl) SubscribeAll(subscriberName string) (chan NodeStatus, func()) {
	eventChan := make(chan NodeStatus, 10) // Buffered channel

	n.mu.Lock()
	defer n.mu.Unlock()

	n.allStatusSubscribers = append(n.allStatusSubscribers, NodeStatusSubscriber{ch: eventChan, name: subscriberName})

	// Create unsubscribe function
	unsubscribe := func() {
		n.unsubscribeAll(subscriberName)
	}

	return eventChan, unsubscribe
}

func (n *NodeStatusDispatcherImpl) Subscribe(statusList []NodeStatus, subscriberName string) (chan NodeStatus, func()) {
	eventChan := make(chan NodeStatus, 10) // Buffered channel

	n.mu.Lock()
	defer n.mu.Unlock()

	for _, status := range statusList {
		if _, ok := n.specificStatusSubscribers[status]; !ok {
			n.specificStatusSubscribers[status] = make([]NodeStatusSubscriber, 0)
		}
		n.specificStatusSubscribers[status] = append(n.specificStatusSubscribers[status], NodeStatusSubscriber{ch: eventChan, name: subscriberName})
	}

	// Create unsubscribe function
	unsubscribe := func() {
		n.unsubscribe(statusList, subscriberName)
	}

	return eventChan, unsubscribe
}

func (n *NodeStatusDispatcherImpl) unsubscribe(statusList []NodeStatus, subscriberName string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, status := range statusList {
		if subscribers, exists := n.specificStatusSubscribers[status]; exists {
			for i, subscriber := range subscribers {
				if subscriber.name == subscriberName {
					n.specificStatusSubscribers[status] = append(subscribers[:i], subscribers[i+1:]...)
					close(subscriber.ch)
					break
				}
			}
		}
	}
}

func (n *NodeStatusDispatcherImpl) unsubscribeAll(subscriberName string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Find and remove the channel from allSubscribers
	for i, subscriber := range n.allStatusSubscribers {
		if subscriber.name == subscriberName {
			n.allStatusSubscribers = append(n.allStatusSubscribers[:i], n.allStatusSubscribers[i+1:]...)
			close(subscriber.ch)
			break
		}
	}
}
