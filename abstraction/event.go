package abstraction

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxEventType interface {
	GetName() string
	GetID() int
}

type OutboxEvent struct {
	EventID       string      `json:"event_id"`
	EventType     string      `json:"event_type"`
	EventStatus   string      `json:"event_status"`
	AggregateType string      `json:"aggregate_type"`
	AggregateID   string      `json:"aggregate_id"`
	Payload       interface{} `json:"payload"`

	CreatedAt time.Time
}

func CreateNewEvent(eventType OutboxEventType, AggregateType, AggregateID string, payload interface{}) OutboxEvent {
	tm := time.Now()
	return OutboxEvent{
		EventID:       uuid.New().String(),
		EventType:     eventType.GetName(),
		EventStatus:   string(OutboxEventStatusPending),
		AggregateType: AggregateType,
		AggregateID:   AggregateID,
		Payload:       payload,
		CreatedAt:     tm,
	}
}

type OutboxEventHandler func(ctx context.Context, event *OutboxEvent) error

type OutboxEventStatus string

const (
	OutboxEventStatusPending OutboxEventStatus = "pending"
	OutboxEventStatusClosed  OutboxEventStatus = "closed"
	OutboxEventStatusFailed  OutboxEventStatus = "failed"
)

func (e OutboxEventStatus) String() string {
	return string(e)
}

func (n *OutboxEventStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*n = OutboxEventStatus(s)
	return nil
}

func (n OutboxEventStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}
