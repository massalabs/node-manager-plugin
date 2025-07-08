package eventBus

import (
	"fmt"
	"sync"
)

type EventType string

type Event struct {
	Type EventType
	Data interface{}
}

type Subscriber struct {
	Name      string
	EventChan chan Event
}

type EventBus interface {
	Publish(event Event) error
	Subscribe(event EventType, subscriberName string) chan Event
	Unsubscribe(event EventType, ch chan Event)
}

type EventBusImpl struct {
	mu          sync.RWMutex
	subscribers map[EventType][]Subscriber
	chanBufSize uint32
}

func NewEventBus(chanBufSize uint32) EventBus {
	return &EventBusImpl{
		subscribers: make(map[EventType][]Subscriber),
		chanBufSize: chanBufSize,
	}
}

func (e *EventBusImpl) Publish(event Event) error {
	failedSubscribers := make([]string, 0)
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, subscriber := range e.subscribers[event.Type] {
		select {
		case subscriber.EventChan <- event:
		default:
			failedSubscribers = append(failedSubscribers, subscriber.Name)
		}
	}
	if len(failedSubscribers) > 0 {
		return fmt.Errorf("Could not send event %s to subscribers: %v", event.Type, failedSubscribers)
	}
	return nil
}

func (e *EventBusImpl) Subscribe(eventType EventType, subscriberName string) chan Event {
	eventChan := make(chan Event, e.chanBufSize)

	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.subscribers[eventType]; !ok {
		e.subscribers[eventType] = make([]Subscriber, 0)
	}
	e.subscribers[eventType] = append(e.subscribers[eventType], Subscriber{
		Name:      subscriberName,
		EventChan: eventChan,
	})
	return eventChan
}

func (e *EventBusImpl) Unsubscribe(eventType EventType, ch chan Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if subscribers, exists := e.subscribers[eventType]; exists {
		for i, subscriber := range subscribers {
			if subscriber.EventChan == ch {
				e.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
				close(subscriber.EventChan)
				break
			}
		}
	}
}
