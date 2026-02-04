# Paanj Client SDK for Go

Official Go SDK for Paanj - Real-time communication platform.

[![Go Reference](https://pkg.go.dev/badge/github.com/paanj-cloud/client-go.svg)](https://pkg.go.dev/github.com/paanj-cloud/client-go)

## Installation

```bash
go get github.com/paanj-cloud/client-go@latest
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    client "github.com/paanj-cloud/client-go"
)

func main() {
    // Initialize client
    paanjClient := client.NewClient(client.ClientOptions{
        ApiKey: "your-api-key",
        ApiUrl: "https://api1.paanj.com",
        WsUrl:  "wss://ws1.paanj.com",
    })

    // Authenticate anonymously
    auth, err := paanjClient.AuthenticateAnonymous(map[string]interface{}{
        "name": "John Doe",
    }, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Authenticated! User ID: %s\\n", auth.UserId)

    // Connect to WebSocket
    if err := paanjClient.Connect(); err != nil {
        log.Fatal(err)
    }
    defer paanjClient.Disconnect()

    fmt.Println("Connected to Paanj!")
}
```

## Features

- ✅ Anonymous authentication
- ✅ WebSocket real-time connections
- ✅ Automatic reconnection
- ✅ Event-based architecture
- ✅ Type-safe API

## API Reference

### Client Initialization

```go
client := client.NewClient(client.ClientOptions{
    ApiKey:               "your-api-key",
    ApiUrl:               "https://api1.paanj.com",    // Optional
    WsUrl:                "wss://ws1.paanj.com",       // Optional
    AutoReconnect:        true,                        // Optional
    ReconnectInterval:    5 * time.Second,             // Optional
    MaxReconnectAttempts: 10,                          // Optional
})
```

### Authentication

#### Anonymous Authentication

```go
auth, err := client.AuthenticateAnonymous(map[string]interface{}{
    "name":  "User Name",
    "email": "user@example.com",
}, map[string]interface{}{
    "customField": "value",
})
```

#### Token Authentication

```go
client.AuthenticateWithToken(accessToken, userId, refreshToken)
```

### WebSocket Connection

#### Connect

```go
err := client.Connect()
```

#### Disconnect

```go
client.Disconnect()
```

#### Check Connection Status

```go
isConnected := client.IsConnected()
```

### Event Handling

```go
client.On("event.name", func(data interface{}) {
    fmt.Printf("Event received: %+v\\n", data)
})
```

### Subscriptions

```go
err := client.Subscribe(map[string]interface{}{
    "resource": "users",
    "id":       userId,
    "events":   []string{"user.updated"},
})
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    client "github.com/paanj-cloud/client-go"
)

func main() {
    // Initialize
    paanjClient := client.NewClient(client.ClientOptions{
        ApiKey:            "pk_live_your_api_key",
        ApiUrl:            "https://api1.paanj.com",
        WsUrl:             "wss://ws1.paanj.com",
        AutoReconnect:     true,
        ReconnectInterval: 5 * time.Second,
    })

    // Authenticate
    auth, err := paanjClient.AuthenticateAnonymous(map[string]interface{}{
        "name": "Alice",
    }, nil)
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }

    fmt.Printf("✅ Authenticated as User ID: %s\\n", auth.UserId)

    // Connect to WebSocket
    if err := paanjClient.Connect(); err != nil {
        log.Fatalf("Connection failed: %v", err)
    }
    defer paanjClient.Disconnect()

    fmt.Println("✅ Connected to WebSocket")

    // Listen for events
    paanjClient.On("user.created", func(data interface{}) {
        fmt.Printf("📢 User created event: %+v\\n", data)
    })

    // Subscribe to user events
    err = paanjClient.Subscribe(map[string]interface{}{
        "resource": "users",
        "id":       auth.UserId,
        "events":   []string{"user.updated", "user.deleted"},
    })
    if err != nil {
        log.Printf("Subscription failed: %v", err)
    }

    // Keep alive
    fmt.Println("Listening for events... (Press Ctrl+C to exit)")
    select {}
}
```

## Error Handling

```go
auth, err := client.AuthenticateAnonymous(userData, nil)
if err != nil {
    // Handle authentication error
    log.Printf("Auth error: %v", err)
    return
}

if err := client.Connect(); err != nil {
    // Handle connection error
    log.Printf("Connection error: %v", err)
    return
}
```

## Advanced Usage

### Custom HTTP Client

The SDK uses a standard `http.Client` internally. For advanced use cases, you can access the HTTP client:

```go
httpClient := client.GetHttpClient()
```

### WebSocket Client Access

```go
wsClient := client.GetWebSocket()
```

## License

MIT License - see LICENSE file for details.

## Support

- Documentation: https://docs.paanj.com
- Issues: https://github.com/paanj-cloud/client-go/issues
- Email: support@paanj.com
