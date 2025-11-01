# Extension Guide: Creating New Event Channel Implementations

This guide shows how to create custom implementations of the Outbox Pattern for different mediums (NATS, Redis, Kafka, etc.).

## Architecture Benefits

The abstraction/implementation separation provides:

✅ **Dependency Inversion** - Depend on interfaces, not implementations  
✅ **Open/Closed Principle** - Open for extension, closed for modification  
✅ **Easy Testing** - Mock implementations for testing  
✅ **Pluggable** - Switch implementations without changing business logic  
✅ **No Vendor Lock-in** - Not tied to specific technology  

## Creating a New Implementation

### Step 1: Create Module Structure

```bash
mkdir -p outbox_uow_nats/nats_channel
cd outbox_uow_nats
go mod init outbox_uow_nats
```

### Step 2: Add Dependencies

```go
// go.mod
module outbox_uow_nats

go 1.23

require (
    github.com/arash/outbox_abstraction v0.0.0
    github.com/nats-io/nats.go v1.31.0
)

replace github.com/arash/outbox_abstraction => ../outbox_uow_abstraction
```

### Step 3: Implement the Interface

```go
// nats_channel/channel.go
package natschannel

import (
    "context"
    "encoding/json"
    "log"
    "time"

    "github.com/arash/outbox_abstraction/abstraction"
    "github.com/nats-io/nats.go"
)

// Ensure interface compliance at compile time
var _ abstraction.OutboxEventChannel = (*NatsEventChannel)(nil)

type NatsEventChannel struct {
    conn      *nats.Conn
    subject   string
    eventType abstraction.OutboxEventType
}

func NewNatsEventChannel(
    natsURL string,
    eventType abstraction.OutboxEventType,
) (*NatsEventChannel, error) {
    // Connect to NATS
    conn, err := nats.Connect(natsURL)
    if err != nil {
        return nil, err
    }

    return &NatsEventChannel{
        conn:      conn,
        subject:   eventType.GetName(),
        eventType: eventType,
    }, nil
}

// RegisterEvent implements OutboxEventChannel
func (c *NatsEventChannel) RegisterEvent(event abstraction.OutboxEvent) error {
    // Serialize event
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // Publish to NATS
    return c.conn.Publish(c.subject, data)
}

func (c *NatsEventChannel) Close() error {
    c.conn.Close()
    return nil
}
```

### Step 4: Usage

```go
package main

import (
    "github.com/arash/outbox_abstraction/abstraction"
    "outbox_uow_nats/nats_channel"
)

type UserCreated struct{}
func (e UserCreated) GetName() string { return "UserCreated" }
func (e UserCreated) GetID() int      { return 1 }

func main() {
    // Create NATS channel
    natsChannel, err := natschannel.NewNatsEventChannel(
        "nats://localhost:4222",
        UserCreated{},
    )
    if err != nil {
        panic(err)
    }
    defer natsChannel.Close()

    // Register with manager
    manager := abstraction.NewOutboxEventManager()
    err = manager.RegisterEventChannel(UserCreated{}, natsChannel)
    if err != nil {
        panic(err)
    }

    // Use it!
    event := abstraction.CreateNewEvent(
        UserCreated{},
        "User",
        "user-123",
        map[string]interface{}{"email": "user@example.com"},
    )

    channel, _ := manager.GetChannel(UserCreated{})
    channel.RegisterEvent(event) // Publishes to NATS
}
```

## Example Implementations

### Redis Streams Implementation

```go
// outbox_uow_redis/redis_channel/channel.go
package redischannel

import (
    "context"
    "encoding/json"

    "github.com/arash/outbox_abstraction/abstraction"
    "github.com/redis/go-redis/v9"
)

type RedisEventChannel struct {
    client    *redis.Client
    stream    string
    eventType abstraction.OutboxEventType
}

func NewRedisEventChannel(
    redisURL string,
    eventType abstraction.OutboxEventType,
) (*RedisEventChannel, error) {
    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        return nil, err
    }

    client := redis.NewClient(opts)

    return &RedisEventChannel{
        client:    client,
        stream:    "events:" + eventType.GetName(),
        eventType: eventType,
    }, nil
}

func (c *RedisEventChannel) RegisterEvent(event abstraction.OutboxEvent) error {
    ctx := context.Background()
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    // Add to Redis Stream
    return c.client.XAdd(ctx, &redis.XAddArgs{
        Stream: c.stream,
        Values: map[string]interface{}{
            "event_id": event.EventID,
            "data":     string(data),
        },
    }).Err()
}
```

### Kafka Implementation

```go
// outbox_uow_kafka/kafka_channel/channel.go
package kafkachannel

import (
    "encoding/json"

    "github.com/arash/outbox_abstraction/abstraction"
    "github.com/segmentio/kafka-go"
)

type KafkaEventChannel struct {
    writer    *kafka.Writer
    eventType abstraction.OutboxEventType
}

func NewKafkaEventChannel(
    brokers []string,
    eventType abstraction.OutboxEventType,
) *KafkaEventChannel {
    return &KafkaEventChannel{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Topic:    eventType.GetName(),
            Balancer: &kafka.LeastBytes{},
        },
        eventType: eventType,
    }
}

func (c *KafkaEventChannel) RegisterEvent(event abstraction.OutboxEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    return c.writer.WriteMessages(context.Background(),
        kafka.Message{
            Key:   []byte(event.EventID),
            Value: data,
        },
    )
}
```

### Hybrid Implementation (PostgreSQL + NATS)

```go
// Combines PostgreSQL persistence with NATS publishing
type HybridEventChannel struct {
    pgChannel   *pgsqlchannel.PgSqlEventChannel
    natsChannel *natschannel.NatsEventChannel
}

func (c *HybridEventChannel) RegisterEvent(event abstraction.OutboxEvent) error {
    // Save to PostgreSQL first (reliable storage)
    if err := c.pgChannel.RegisterEvent(event); err != nil {
        return err
    }

    // Then publish to NATS (fast distribution)
    // Errors here don't fail the registration
    if err := c.natsChannel.RegisterEvent(event); err != nil {
        log.Printf("Warning: NATS publish failed: %v", err)
    }

    return nil
}
```

## Mixed Implementation Strategy

Use different implementations for different event types:

```go
manager := abstraction.NewOutboxEventManager()

// Critical events → PostgreSQL (reliable, with retry)
pgChannel := pgsqlchannel.NewPgSqlEventChannel(...)
manager.RegisterEventChannel(OrderPlaced{}, pgChannel)
manager.RegisterEventChannel(PaymentProcessed{}, pgChannel)

// High-volume events → NATS (fast, lightweight)
natsChannel, _ := natschannel.NewNatsEventChannel(...)
manager.RegisterEventChannel(UserViewed{}, natsChannel)
manager.RegisterEventChannel(SearchPerformed{}, natsChannel)

// Analytics events → Kafka (streaming)
kafkaChannel := kafkachannel.NewKafkaEventChannel(...)
manager.RegisterEventChannel(AnalyticsEvent{}, kafkaChannel)
```

## Testing Your Implementation

```go
package natschannel

import (
    "testing"
    "github.com/arash/outbox_abstraction/abstraction"
    "github.com/nats-io/nats-server/v2/server"
    "github.com/nats-io/nats.go"
)

func TestNatsEventChannel_RegisterEvent(t *testing.T) {
    // Start embedded NATS server for testing
    opts := &server.Options{
        Host: "127.0.0.1",
        Port: -1, // Random port
    }
    ns, _ := server.NewServer(opts)
    go ns.Start()
    defer ns.Shutdown()

    // Wait for server to be ready
    if !ns.ReadyForConnections(time.Second) {
        t.Fatal("Server not ready")
    }

    // Create channel
    channel, err := NewNatsEventChannel(
        ns.ClientURL(),
        mockEventType{name: "Test"},
    )
    if err != nil {
        t.Fatal(err)
    }
    defer channel.Close()

    // Test event registration
    event := abstraction.CreateNewEvent(
        mockEventType{name: "Test"},
        "Test",
        "test-1",
        map[string]interface{}{"key": "value"},
    )

    err = channel.RegisterEvent(event)
    if err != nil {
        t.Errorf("Failed to register event: %v", err)
    }
}

// Ensure interface compliance
func TestNatsEventChannel_ImplementsInterface(t *testing.T) {
    var _ abstraction.OutboxEventChannel = (*NatsEventChannel)(nil)
}
```

## Best Practices

### 1. Interface Compliance Check

```go
// Add this at the top of your implementation file
var _ abstraction.OutboxEventChannel = (*YourChannel)(nil)
```

This ensures compile-time verification that your type implements the interface.

### 2. Error Handling

```go
func (c *YourChannel) RegisterEvent(event abstraction.OutboxEvent) error {
    // Validate input
    if event.EventID == "" {
        return errors.New("event ID is required")
    }

    // Your implementation
    if err := c.doSomething(event); err != nil {
        return fmt.Errorf("failed to register event: %w", err)
    }

    return nil
}
```

### 3. Configuration

```go
type ChannelConfig struct {
    URL         string
    MaxRetries  int
    Timeout     time.Duration
    // ... other options
}

func NewChannel(config ChannelConfig, eventType abstraction.OutboxEventType) (*Channel, error) {
    // Validate config
    // Create channel
}
```

### 4. Resource Cleanup

```go
type YourChannel struct {
    // ... fields
    done chan struct{}
}

func (c *YourChannel) Close() error {
    close(c.done)
    // Cleanup resources
    return nil
}
```

## Migration Example

### Phase 1: Start with PostgreSQL

```go
manager := abstraction.NewOutboxEventManager()
pgChannel := pgsqlchannel.NewPgSqlEventChannel(...)
manager.RegisterEventChannel(UserCreated{}, pgChannel)
```

### Phase 2: Add NATS (Dual Publishing)

```go
// Keep PostgreSQL for reliability
manager.RegisterEventChannel(UserCreated{}, hybridChannel)
// Hybrid publishes to both PostgreSQL and NATS
```

### Phase 3: Move to NATS Only

```go
// Replace with pure NATS
natsChannel, _ := natschannel.NewNatsEventChannel(...)
manager.RegisterEventChannel(UserCreated{}, natsChannel)
```

## Contribution Guide

To contribute a new implementation:

1. Create new module: `outbox_uow_<technology>/`
2. Implement `OutboxEventChannel` interface
3. Add comprehensive tests
4. Document configuration options
5. Add usage examples
6. Submit pull request

## Summary

Your abstraction is **excellent** because:

✅ Follows SOLID principles  
✅ Easy to extend (new implementations)  
✅ Easy to test (mock implementations)  
✅ Easy to migrate (swap implementations)  
✅ Technology agnostic  
✅ No vendor lock-in  
✅ Supports mixed strategies  
