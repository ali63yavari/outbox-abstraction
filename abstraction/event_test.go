package abstraction

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Mock OutboxEventType for testing
type mockEventType struct {
	name string
	id   int
}

func (m mockEventType) GetName() string {
	return m.name
}

func (m mockEventType) GetID() int {
	return m.id
}

func TestCreateNewEvent(t *testing.T) {
	eventType := mockEventType{name: "UserCreated", id: 1}
	aggregateType := "User"
	aggregateID := "user-123"
	payload := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	event := CreateNewEvent(eventType, aggregateType, aggregateID, payload)

	// Verify EventID is generated
	if event.EventID == "" {
		t.Error("EventID should not be empty")
	}

	// Verify EventID is a valid UUID
	if _, err := uuid.Parse(event.EventID); err != nil {
		t.Errorf("EventID should be a valid UUID, got: %s, error: %v", event.EventID, err)
	}

	// Verify EventType
	if event.EventType != "UserCreated" {
		t.Errorf("Expected EventType 'UserCreated', got: %s", event.EventType)
	}

	// Verify EventStatus
	if event.EventStatus != string(OutboxEventStatusPending) {
		t.Errorf("Expected EventStatus 'pending', got: %s", event.EventStatus)
	}

	// Verify AggregateType
	if event.AggregateType != aggregateType {
		t.Errorf("Expected AggregateType '%s', got: %s", aggregateType, event.AggregateType)
	}

	// Verify AggregateID
	if event.AggregateID != aggregateID {
		t.Errorf("Expected AggregateID '%s', got: %s", aggregateID, event.AggregateID)
	}

	// Verify Payload
	if event.Payload == nil {
		t.Error("Payload should not be nil")
	}

	// Verify CreatedAt
	if event.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// Verify CreatedAt is recent (within last second)
	if time.Since(event.CreatedAt) > time.Second {
		t.Error("CreatedAt should be recent")
	}
}

func TestCreateNewEvent_UniqueEventIDs(t *testing.T) {
	eventType := mockEventType{name: "UserCreated", id: 1}
	aggregateType := "User"
	aggregateID := "user-123"
	payload := map[string]interface{}{"test": "data"}

	// Create multiple events
	event1 := CreateNewEvent(eventType, aggregateType, aggregateID, payload)
	event2 := CreateNewEvent(eventType, aggregateType, aggregateID, payload)
	event3 := CreateNewEvent(eventType, aggregateType, aggregateID, payload)

	// Verify all EventIDs are unique
	if event1.EventID == event2.EventID {
		t.Error("EventIDs should be unique")
	}
	if event1.EventID == event3.EventID {
		t.Error("EventIDs should be unique")
	}
	if event2.EventID == event3.EventID {
		t.Error("EventIDs should be unique")
	}
}

func TestOutboxEventStatus_String(t *testing.T) {
	tests := []struct {
		status   OutboxEventStatus
		expected string
	}{
		{OutboxEventStatusPending, "pending"},
		{OutboxEventStatusClosed, "closed"},
		{OutboxEventStatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.status.String() != tt.expected {
				t.Errorf("Expected %s, got: %s", tt.expected, tt.status.String())
			}
		})
	}
}

func TestOutboxEventStatus_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		status   OutboxEventStatus
		expected string
	}{
		{"pending", OutboxEventStatusPending, `"pending"`},
		{"closed", OutboxEventStatusClosed, `"closed"`},
		{"failed", OutboxEventStatusFailed, `"failed"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("MarshalJSON failed: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("Expected %s, got: %s", tt.expected, string(data))
			}
		})
	}
}

func TestOutboxEventStatus_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected OutboxEventStatus
		wantErr  bool
	}{
		{"pending", `"pending"`, OutboxEventStatusPending, false},
		{"closed", `"closed"`, OutboxEventStatusClosed, false},
		{"failed", `"failed"`, OutboxEventStatusFailed, false},
		{"invalid json", `invalid`, OutboxEventStatus(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status OutboxEventStatus
			err := json.Unmarshal([]byte(tt.input), &status)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("UnmarshalJSON failed: %v", err)
			}

			if status != tt.expected {
				t.Errorf("Expected %s, got: %s", tt.expected, status)
			}
		})
	}
}

func TestOutboxEvent_JSONMarshaling(t *testing.T) {
	eventType := mockEventType{name: "UserCreated", id: 1}
	payload := map[string]interface{}{
		"email": "test@example.com",
		"name":  "Test User",
	}

	event := CreateNewEvent(eventType, "User", "user-123", payload)

	// Marshal to JSON
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Unmarshal back
	var unmarshaled OutboxEvent
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify fields
	if unmarshaled.EventID != event.EventID {
		t.Errorf("EventID mismatch: expected %s, got %s", event.EventID, unmarshaled.EventID)
	}
	if unmarshaled.EventType != event.EventType {
		t.Errorf("EventType mismatch: expected %s, got %s", event.EventType, unmarshaled.EventType)
	}
	if unmarshaled.EventStatus != event.EventStatus {
		t.Errorf("EventStatus mismatch: expected %s, got %s", event.EventStatus, unmarshaled.EventStatus)
	}
	if unmarshaled.AggregateType != event.AggregateType {
		t.Errorf("AggregateType mismatch: expected %s, got %s",
			event.AggregateType, unmarshaled.AggregateType)
	}
	if unmarshaled.AggregateID != event.AggregateID {
		t.Errorf("AggregateID mismatch: expected %s, got %s", event.AggregateID, unmarshaled.AggregateID)
	}
}
