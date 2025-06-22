# Supabase Go Library

A comprehensive Golang library for connecting to Supabase with all its functionalities including Authentication, Database operations, Storage, Realtime subscriptions, and Edge Functions.

## Features

- 🔐 **Authentication**: Complete user management with email, phone, OAuth, and magic link support
- 🗄️ **Database**: Full CRUD operations with query building, filtering, and pagination
- 📁 **Storage**: File upload, download, management with buckets and signed URLs
- ⚡ **Realtime**: WebSocket-based real-time subscriptions and broadcasting
- 🚀 **Edge Functions**: Call Supabase Edge Functions
- 🛡️ **Error Handling**: Comprehensive error handling with custom error types
- 📝 **Logging**: Integrated logging support
- ⚙️ **Configuration**: Flexible configuration management

## Installation

```bash
go get github.com/himdhiman/dashboard-backend/libs/supabase
```

## Quick Start

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/himdhiman/dashboard-backend/libs/logger"
    "github.com/himdhiman/dashboard-backend/libs/supabase"
    "github.com/himdhiman/dashboard-backend/libs/supabase/config"
)

func main() {
    // Initialize logger
    logger := logger.NewLogger()
    
    // Create Supabase configuration
    cfg := config.NewConfig().
        WithURL("https://your-project.supabase.co").
        WithKey("your-anon-key").
        WithServiceKey("your-service-key").
        WithTimeout(30 * time.Second)
    
    // Create Supabase client
    client, err := supabase.NewSupabaseClient(cfg, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Test connection
    if err := client.Ping(context.Background()); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Connected to Supabase!")
}
```

### Environment Variables

You can also load configuration from environment variables:

```bash
export SUPABASE_URL="https://your-project.supabase.co"
export SUPABASE_ANON_KEY="your-anon-key"
export SUPABASE_SERVICE_KEY="your-service-key"
export SUPABASE_JWT_SECRET="your-jwt-secret"
```

```go
// Load from environment
client, err := supabase.NewSupabaseClientFromEnv(logger)
```

## Authentication

### User Sign Up

```go
// Sign up a new user
signUpReq := &models.SignUpRequest{
    Email:    "user@example.com",
    Password: "securepassword",
    Data: map[string]interface{}{
        "name": "John Doe",
        "age":  30,
    },
}

authResp, err := client.Auth().SignUp(ctx, signUpReq)
if err != nil {
    log.Fatal(err)
}

log.Printf("User created: %s", authResp.User.ID)
```

### User Sign In

```go
// Sign in a user
signInReq := &models.SignInRequest{
    Email:    "user@example.com",
    Password: "securepassword",
}

authResp, err := client.Auth().SignIn(ctx, signInReq)
if err != nil {
    log.Fatal(err)
}

log.Printf("User signed in: %s", authResp.User.Email)
```

### OAuth Authentication

```go
// Sign in with OAuth
authURL, err := client.Auth().SignInWithOAuth(ctx, "google", map[string]interface{}{
    "redirectTo": "https://yourapp.com/callback",
})
if err != nil {
    log.Fatal(err)
}

log.Printf("OAuth URL: %s", authURL)
```

### Magic Link

```go
// Send magic link
err := client.Auth().SignInWithMagicLink(ctx, "user@example.com", map[string]interface{}{
    "redirectTo": "https://yourapp.com/callback",
})
if err != nil {
    log.Fatal(err)
}

log.Println("Magic link sent!")
```

## Database Operations

### Basic CRUD Operations

```go
// Define your data model
type User struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// Get a typed repository
usersRepo := client.FromTyped[User]("users")

// Insert a new user
newUser := User{
    Name:  "John Doe",
    Email: "john@example.com",
    Age:   30,
}

insertedUser, err := usersRepo.Insert(ctx, &newUser)
if err != nil {
    log.Fatal(err)
}

log.Printf("User inserted: %s", insertedUser.ID)
```

### Query Building

```go
// Complex queries with filtering, ordering, and pagination
users, err := usersRepo.
    Select(ctx, "id", "name", "email").
    Eq("age", 30).
    Gte("created_at", "2024-01-01").
    Like("name", "%John%").
    Order("created_at", "desc").
    Limit(10).
    Offset(0).
    Execute(ctx)
if err != nil {
    log.Fatal(err)
}

for _, user := range users {
    log.Printf("User: %s - %s", user.Name, user.Email)
}
```

### Advanced Queries

```go
// Text search
users, err := usersRepo.
    TextSearch("name", "john doe").
    Execute(ctx)

// Range queries
users, err := usersRepo.
    RangeGte("age", 18).
    RangeLte("age", 65).
    Execute(ctx)

// Array operations
users, err := usersRepo.
    In("status", []interface{}{"active", "pending"}).
    Contains("tags", []string{"developer", "golang"}).
    Execute(ctx)
```

### Update Operations

```go
// Update a single record
user, err := usersRepo.
    Eq("id", "user-id").
    Update(ctx, &User{
        Name: "Jane Doe",
        Age:  31,
    })

// Update multiple records
err = usersRepo.
    Eq("status", "inactive").
    UpdateMany(ctx, map[string]interface{}{
        "status": "active",
        "updated_at": time.Now(),
    })
```

### Delete Operations

```go
// Delete records
err = usersRepo.
    Eq("status", "deleted").
    Delete(ctx)

// Delete by ID
err = usersRepo.
    Eq("id", "user-id").
    Delete(ctx)
```

## Storage Operations

### Bucket Management

```go
// Create a bucket
bucket, err := client.Storage().CreateBucket(ctx, "avatars", map[string]interface{}{
    "public": true,
    "file_size_limit": 5242880, // 5MB
})
if err != nil {
    log.Fatal(err)
}

// List buckets
buckets, err := client.Storage().ListBuckets(ctx)
if err != nil {
    log.Fatal(err)
}

for _, bucket := range buckets {
    log.Printf("Bucket: %s (public: %t)", bucket.Name, bucket.Public)
}
```

### File Operations

```go
// Upload a file
fileData := []byte("Hello, World!")
file, err := client.Storage().Upload(ctx, "avatars", "user-123/avatar.jpg", fileData, map[string]interface{}{
    "upsert": true,
})
if err != nil {
    log.Fatal(err)
}

// Upload from local file
file, err = client.Storage().UploadFromFile(ctx, "avatars", "user-123/avatar.jpg", "/path/to/local/file.jpg", nil)

// Download a file
data, err := client.Storage().Download(ctx, "avatars", "user-123/avatar.jpg")
if err != nil {
    log.Fatal(err)
}

// Download to local file
err = client.Storage().DownloadToFile(ctx, "avatars", "user-123/avatar.jpg", "/path/to/download/file.jpg")
```

### Bucket-Specific Operations

```go
// Get a bucket-specific client
avatarsBucket := client.Storage().From("avatars")

// Upload to specific bucket
file, err := avatarsBucket.Upload(ctx, "user-123/avatar.jpg", fileData, nil)

// Get public URL
publicURL := avatarsBucket.GetPublicURL("user-123/avatar.jpg")

// Create signed URL
signedURL, err := avatarsBucket.CreateSignedURL(ctx, "user-123/avatar.jpg", 3600, nil)
```

### File Management

```go
// List files in a bucket
files, err := client.Storage().List(ctx, "avatars", "user-123/", nil)
if err != nil {
    log.Fatal(err)
}

for _, file := range files {
    log.Printf("File: %s (%d bytes)", file.Name, file.Metadata["size"])
}

// Move a file
err = client.Storage().Move(ctx, "avatars", "user-123/old.jpg", "user-123/new.jpg")

// Copy a file
err = client.Storage().Copy(ctx, "avatars", "user-123/original.jpg", "user-123/backup.jpg")

// Delete files
err = client.Storage().Delete(ctx, "avatars", []string{"user-123/old.jpg", "user-123/temp.jpg"})
```

## Realtime Subscriptions

### Basic Subscription

```go
// Connect to realtime
err := client.Realtime().Connect(ctx)
if err != nil {
    log.Fatal(err)
}

// Subscribe to table changes
err = client.Realtime().Subscribe("users", func(message *models.RealtimeMessage) {
    log.Printf("User change: %s - %s", message.EventType, message.Table)
    log.Printf("Record: %+v", message.Record)
})
if err != nil {
    log.Fatal(err)
}
```

### Channel-Based Operations

```go
// Get a channel
channel := client.Realtime().Channel("users")

// Subscribe to specific events
channel.On("INSERT", func(message *models.RealtimeMessage) {
    log.Println("New user inserted!")
})

channel.On("UPDATE", func(message *models.RealtimeMessage) {
    log.Println("User updated!")
})

channel.On("DELETE", func(message *models.RealtimeMessage) {
    log.Println("User deleted!")
})

// Subscribe to the channel
err := channel.Subscribe(func(message *models.RealtimeMessage) {
    log.Printf("Received message: %+v", message)
})
```

### Broadcasting

```go
// Broadcast a message
err := client.Realtime().Broadcast("chat", "message", map[string]interface{}{
    "user": "john",
    "text": "Hello, everyone!",
})

// Broadcast to self only
err = client.Realtime().BroadcastToSelf("chat", "typing", map[string]interface{}{
    "user": "john",
    "typing": true,
})
```

### Presence

```go
// Track presence
err := client.Realtime().Track(map[string]interface{}{
    "user_id": "user-123",
    "status": "online",
    "last_seen": time.Now(),
})

// Get presence for a channel
presence, err := client.Realtime().Presence("chat")
if err != nil {
    log.Fatal(err)
}

log.Printf("Online users: %+v", presence)
```

## Edge Functions

```go
// Call an edge function
result, err := client.EdgeFunction("hello-world", map[string]interface{}{
    "name": "John",
    "age":  30,
})
if err != nil {
    log.Fatal(err)
}

log.Printf("Function result: %+v", result)
```

## Error Handling

The library provides comprehensive error handling with specific error types:

```go
import "github.com/himdhiman/dashboard-backend/libs/supabase/errors"

// Check for specific errors
if errors.IsError(err, errors.ErrUnauthorized) {
    log.Println("User is not authorized")
} else if errors.IsError(err, errors.ErrNotFound) {
    log.Println("Resource not found")
} else if errors.IsError(err, errors.ErrValidationFailed) {
    log.Println("Validation failed")
}

// Create custom errors with context
customErr := errors.NewError(err, "failed to process user data")
```

## Configuration Options

```go
cfg := config.NewConfig().
    WithURL("https://your-project.supabase.co").
    WithKey("your-anon-key").
    WithServiceKey("your-service-key").
    WithJWTSecret("your-jwt-secret").
    WithTimeout(30 * time.Second).
    WithMaxRetries(3).
    WithRetryDelay(1 * time.Second).
    WithLogging(true)

// Validate configuration
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

## Best Practices

### 1. Connection Management

```go
// Always close the client when done
defer client.Close()

// Check connection status
if !client.IsConnected() {
    log.Println("Client is not connected")
}
```

### 2. Context Usage

```go
// Use context with timeout for operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

users, err := usersRepo.Execute(ctx)
```

### 3. Error Handling

```go
// Always check for errors
if err != nil {
    // Log the error with context
    logger.Error("Failed to fetch users", "error", err, "table", "users")
    return err
}
```

### 4. Resource Cleanup

```go
// Clean up realtime subscriptions
defer func() {
    client.Realtime().RemoveAllChannels()
    client.Realtime().Disconnect()
}()
```

## Testing

```go
func TestSupabaseClient(t *testing.T) {
    // Create test configuration
    cfg := config.NewConfig().
        WithURL("https://test-project.supabase.co").
        WithKey("test-key")
    
    // Create test logger
    logger := logger.NewLogger()
    
    // Create client
    client, err := supabase.NewSupabaseClient(cfg, logger)
    require.NoError(t, err)
    defer client.Close()
    
    // Test ping
    err = client.Ping(context.Background())
    require.NoError(t, err)
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions, please open an issue on GitHub or contact the maintainers. 