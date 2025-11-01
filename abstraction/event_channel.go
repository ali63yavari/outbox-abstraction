package abstraction

import (
	"context"
	"time"
)

type ChannelConfig struct {
	MaxRetries int
	BatchSize  int
	Interval   time.Duration
	EventType  OutboxEventType
	Ctx        context.Context
}

type OutboxEventChannel interface {
	RegisterEvent(event OutboxEvent) error
}
