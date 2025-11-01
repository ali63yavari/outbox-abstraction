package abstraction

import (
	"testing"
)

// Mock channel implementation
type mockChannel struct {
	registered bool
	events     []OutboxEvent
}

func (m *mockChannel) RegisterEvent(event OutboxEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestNewOutboxEventManager(t *testing.T) {
	manager := NewOutboxEventManager()
	if manager == nil {
		t.Fatal("NewOutboxEventManager should not return nil")
	}

	// Test that it implements the interface
	_, ok := manager.(OutboxEventManager)
	if !ok {
		t.Error("NewOutboxEventManager should return OutboxEventManager interface")
	}
}

func TestOutboxEventManager_RegisterEventChannel(t *testing.T) {
	manager := NewOutboxEventManager()
	eventType := mockEventType{name: "UserCreated", id: 1}
	channel := &mockChannel{}

	// Register the channel
	err := manager.RegisterEventChannel(eventType, channel)
	if err != nil {
		t.Fatalf("Expected successful registration, got error: %v", err)
	}

	// Verify we can retrieve it
	retrieved, err := manager.GetChannel(eventType)
	if err != nil {
		t.Fatalf("Expected to retrieve channel, got error: %v", err)
	}

	if retrieved != channel {
		t.Error("Retrieved channel should be the same as registered channel")
	}
}

func TestOutboxEventManager_RegisterEventChannel_PreventOverwrite(t *testing.T) {
	manager := NewOutboxEventManager()
	eventType := mockEventType{name: "UserCreated", id: 1}
	channel1 := &mockChannel{}
	channel2 := &mockChannel{}

	// Register first channel
	err := manager.RegisterEventChannel(eventType, channel1)
	if err != nil {
		t.Fatalf("Expected successful registration, got error: %v", err)
	}

	// Attempt to register second channel with same event type (should fail)
	err = manager.RegisterEventChannel(eventType, channel2)
	if err == nil {
		t.Error("Expected error when registering duplicate event type")
	}

	expectedError := "event channel already registered for this event type"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: '%s'", expectedError, err.Error())
	}

	// Verify we still get the first channel
	retrieved, err := manager.GetChannel(eventType)
	if err != nil {
		t.Fatalf("Expected to retrieve channel, got error: %v", err)
	}

	if retrieved != channel1 {
		t.Error("Should retrieve the first (original) channel")
	}

	if retrieved == channel2 {
		t.Error("Should not retrieve the second channel")
	}
}

func TestOutboxEventManager_GetChannel_NotFound(t *testing.T) {
	manager := NewOutboxEventManager()
	eventType := mockEventType{name: "NonExistent", id: 999}

	// Try to get a channel that was never registered
	retrieved, err := manager.GetChannel(eventType)

	if err == nil {
		t.Error("Expected error when getting non-existent channel")
	}

	if retrieved != nil {
		t.Error("Retrieved channel should be nil when not found")
	}

	expectedError := "outbox event channel not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: '%s'", expectedError, err.Error())
	}
}

func TestOutboxEventManager_MultipleChannels(t *testing.T) {
	manager := NewOutboxEventManager()

	// Create multiple event types and channels
	eventType1 := mockEventType{name: "UserCreated", id: 1}
	eventType2 := mockEventType{name: "UserUpdated", id: 2}
	eventType3 := mockEventType{name: "UserDeleted", id: 3}

	channel1 := &mockChannel{}
	channel2 := &mockChannel{}
	channel3 := &mockChannel{}

	// Register multiple channels
	if err := manager.RegisterEventChannel(eventType1, channel1); err != nil {
		t.Fatalf("Failed to register channel 1: %v", err)
	}
	if err := manager.RegisterEventChannel(eventType2, channel2); err != nil {
		t.Fatalf("Failed to register channel 2: %v", err)
	}
	if err := manager.RegisterEventChannel(eventType3, channel3); err != nil {
		t.Fatalf("Failed to register channel 3: %v", err)
	}

	// Verify all channels can be retrieved correctly
	retrieved1, err := manager.GetChannel(eventType1)
	if err != nil {
		t.Fatalf("Failed to retrieve channel 1: %v", err)
	}
	if retrieved1 != channel1 {
		t.Error("Retrieved channel 1 should match registered channel 1")
	}

	retrieved2, err := manager.GetChannel(eventType2)
	if err != nil {
		t.Fatalf("Failed to retrieve channel 2: %v", err)
	}
	if retrieved2 != channel2 {
		t.Error("Retrieved channel 2 should match registered channel 2")
	}

	retrieved3, err := manager.GetChannel(eventType3)
	if err != nil {
		t.Fatalf("Failed to retrieve channel 3: %v", err)
	}
	if retrieved3 != channel3 {
		t.Error("Retrieved channel 3 should match registered channel 3")
	}
}

func TestOutboxEventManager_ChannelIsolation(t *testing.T) {
	manager := NewOutboxEventManager()

	eventType1 := mockEventType{name: "UserCreated", id: 1}
	eventType2 := mockEventType{name: "UserUpdated", id: 2}

	channel1 := &mockChannel{}
	channel2 := &mockChannel{}

	if err := manager.RegisterEventChannel(eventType1, channel1); err != nil {
		t.Fatalf("Failed to register channel 1: %v", err)
	}
	if err := manager.RegisterEventChannel(eventType2, channel2); err != nil {
		t.Fatalf("Failed to register channel 2: %v", err)
	}

	// Add events to channels through the interface
	event1 := CreateNewEvent(eventType1, "User", "user-1", map[string]interface{}{"test": "data1"})
	event2 := CreateNewEvent(eventType2, "User", "user-2", map[string]interface{}{"test": "data2"})

	retrievedChan1, _ := manager.GetChannel(eventType1)
	retrievedChan2, _ := manager.GetChannel(eventType2)

	retrievedChan1.RegisterEvent(event1)
	retrievedChan2.RegisterEvent(event2)

	// Verify channels maintain their own events
	mock1 := retrievedChan1.(*mockChannel)
	mock2 := retrievedChan2.(*mockChannel)

	if len(mock1.events) != 1 {
		t.Errorf("Channel 1 should have 1 event, got: %d", len(mock1.events))
	}
	if len(mock2.events) != 1 {
		t.Errorf("Channel 2 should have 1 event, got: %d", len(mock2.events))
	}

	if mock1.events[0].EventType != "UserCreated" {
		t.Error("Channel 1 should have UserCreated event")
	}
	if mock2.events[0].EventType != "UserUpdated" {
		t.Error("Channel 2 should have UserUpdated event")
	}
}
