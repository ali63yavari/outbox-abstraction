package abstraction

import (
	"context"
	"errors"
)

type OutboxEventManager interface {
	RegisterEventChannel(eventType OutboxEventType, channel OutboxEventChannel) error
	GetChannel(eventType OutboxEventType) (OutboxEventChannel, error)
}

type outboxEventManagerImpl struct {
	ctx           context.Context
	eventChannels map[OutboxEventType]OutboxEventChannel
}

func NewOutboxEventManager() OutboxEventManager {
	om := &outboxEventManagerImpl{
		ctx:           context.Background(),
		eventChannels: make(map[OutboxEventType]OutboxEventChannel),
	}

	return om
}

func (m *outboxEventManagerImpl) RegisterEventChannel(eventType OutboxEventType, channel OutboxEventChannel) error {
	if _, exists := m.eventChannels[eventType]; exists {
		return errors.New("event channel already registered for this event type")
	}
	m.eventChannels[eventType] = channel
	return nil
}

func (m *outboxEventManagerImpl) GetChannel(eventType OutboxEventType) (OutboxEventChannel, error) {
	channel, ok := m.eventChannels[eventType]
	if !ok {
		return nil, errors.New("outbox event channel not found")
	}

	return channel, nil
}
