package stakingManager

import (
	"sync"

	"github.com/massalabs/station/pkg/logger"
)

type AddressChangedSubscriber struct {
	ch   chan []StakingAddress
	name string
}

type AddressChangedDispatcher interface {
	Publish(addresses []StakingAddress)
	Subscribe(subscriberName string) (chan []StakingAddress, func())
	HasSubscribers() bool
}

type AddressChangedDispatcherImpl struct {
	subscribers []AddressChangedSubscriber
	mu          sync.RWMutex
}

func NewAddressChangedDispatcher() AddressChangedDispatcher {
	return &AddressChangedDispatcherImpl{
		subscribers: make([]AddressChangedSubscriber, 0),
	}
}

func (a *AddressChangedDispatcherImpl) Publish(addresses []StakingAddress) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Notify all subscribers
	for _, subscriber := range a.subscribers {
		select {
		case subscriber.ch <- addresses:
		default:
			logger.Debugf("Subscriber %s Channel is full or closed, ignoring address update", subscriber.name)
			// Channel is full or closed, ignore
		}
	}
}

func (a *AddressChangedDispatcherImpl) Subscribe(subscriberName string) (chan []StakingAddress, func()) {
	eventChan := make(chan []StakingAddress, 10) // Buffered channel

	a.mu.Lock()
	defer a.mu.Unlock()

	a.subscribers = append(a.subscribers, AddressChangedSubscriber{ch: eventChan, name: subscriberName})

	// Create unsubscribe function
	unsubscribe := func() {
		a.unsubscribe(subscriberName)
	}

	return eventChan, unsubscribe
}

func (a *AddressChangedDispatcherImpl) unsubscribe(subscriberName string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Find and remove the channel from subscribers
	for i, subscriber := range a.subscribers {
		if subscriber.name == subscriberName {
			a.subscribers = append(a.subscribers[:i], a.subscribers[i+1:]...)
			close(subscriber.ch)
			break
		}
	}
}

func (a *AddressChangedDispatcherImpl) HasSubscribers() bool {
	return len(a.subscribers) > 0
}
