# Outbox Pattern - Core Abstraction

Core interfaces and types for the Outbox Pattern implementation.

## Installation

```bash
go get github.com/ali63yavari/outbox-abstraction
```

## Usage

```go
import "github.com/ali63yavari/outbox-abstraction/abstraction"

// Create event
event := abstraction.CreateNewEvent(...)

// Create manager
manager := abstraction.NewOutboxEventManager()
```

## Implementations

- [PostgreSQL](https://github.com/ali63yavari/outbox-pgsql) - Official implementation
- [NATS](https://github.com/ali63yavari/outbox-nats) - Community implementation
- See [EXTENSION_GUIDE.md](EXTENSION_GUIDE.md) for creating your own

## Documentation

- [Architecture](ARCHITECTURE.md) - Design principles
- [Extension Guide](EXTENSION_GUIDE.md) - Creating implementations
