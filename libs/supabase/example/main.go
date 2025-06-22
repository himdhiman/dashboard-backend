package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/models"
)

// User represents a user in our application
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Age       int       `json:"age"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Post represents a blog post
type Post struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
	Published bool      `json:"published"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	// Initialize logger
	ctx := context.Background()
	logger := logger.New(logger.DefaultConfig("supabase-example")).WithContext(ctx)

	// Create Supabase configuration
	cfg := config.NewConfig().
		WithURL("https://your-project.supabase.co").
		WithKey("your-anon-key").
		WithServiceKey("your-service-key").
		WithTimeout(30 * time.Second).
		WithLogging(true)

	// Create Supabase client
	client, err := supabase.NewSupabaseClient(cfg, logger)
	if err != nil {
		log.Fatal("Failed to create Supabase client:", err)
	}
	defer client.Close()

	// Test connection
	if err := client.Ping(ctx); err != nil {
		log.Fatal("Failed to ping Supabase:", err)
	}

	fmt.Println("✅ Connected to Supabase successfully!")

	// Run examples
	runAuthExamples(ctx, client)
	runDatabaseExamples(ctx, client)
	runStorageExamples(ctx, client)
	runRealtimeExamples(ctx, client)
	runEdgeFunctionExamples(ctx, client)
}

func runAuthExamples(ctx context.Context, client supabase.ISupabaseClient) {
	fmt.Println("\n🔐 Authentication Examples")

	// Example 1: Sign up a new user
	signUpReq := &models.SignUpRequest{
		Email:    "john.doe@example.com",
		Password: "securepassword123",
		Data: map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		},
	}

	authResp, err := client.Auth().SignUp(ctx, signUpReq)
	if err != nil {
		fmt.Printf("❌ Sign up failed: %v\n", err)
	} else {
		fmt.Printf("✅ User signed up: %s\n", authResp.User.Email)
	}

	// Example 2: Sign in a user
	signInReq := &models.SignInRequest{
		Email:    "john.doe@example.com",
		Password: "securepassword123",
	}

	authResp, err = client.Auth().SignIn(ctx, signInReq)
	if err != nil {
		fmt.Printf("❌ Sign in failed: %v\n", err)
	} else {
		fmt.Printf("✅ User signed in: %s\n", authResp.User.Email)
	}

	// Example 3: Get current user
	user, err := client.Auth().GetUser(ctx)
	if err != nil {
		fmt.Printf("❌ Get user failed: %v\n", err)
	} else {
		fmt.Printf("✅ Current user: %s (%s)\n", user.Email, user.ID)
	}

	// Example 4: Update user
	updateReq := &models.UserUpdateRequest{
		Data: map[string]interface{}{
			"age": 31,
		},
	}

	updatedUser, err := client.Auth().UpdateUser(ctx, updateReq)
	if err != nil {
		fmt.Printf("❌ Update user failed: %v\n", err)
	} else {
		fmt.Printf("✅ User updated: %s\n", updatedUser.Email)
	}

	// Example 5: Reset password
	resetReq := &models.PasswordResetRequest{
		Email: "john.doe@example.com",
	}

	err = client.Auth().ResetPassword(ctx, resetReq)
	if err != nil {
		fmt.Printf("❌ Password reset failed: %v\n", err)
	} else {
		fmt.Println("✅ Password reset email sent")
	}

	// Example 6: OAuth sign in
	authURL, err := client.Auth().SignInWithOAuth(ctx, "google", map[string]interface{}{
		"redirectTo": "https://yourapp.com/callback",
	})
	if err != nil {
		fmt.Printf("❌ OAuth sign in failed: %v\n", err)
	} else {
		fmt.Printf("✅ OAuth URL: %s\n", authURL)
	}

	// Example 7: Magic link
	err = client.Auth().SignInWithMagicLink(ctx, "john.doe@example.com", map[string]interface{}{
		"redirectTo": "https://yourapp.com/callback",
	})
	if err != nil {
		fmt.Printf("❌ Magic link failed: %v\n", err)
	} else {
		fmt.Println("✅ Magic link sent")
	}
}

func runDatabaseExamples(ctx context.Context, client supabase.ISupabaseClient) {
	fmt.Println("\n🗄️ Database Examples")

	// Get typed repositories
	usersRepo := supabase.FromTyped[User](client, "users")
	postsRepo := supabase.FromTyped[Post](client, "posts")

	// Example 1: Insert a new user
	newUser := User{
		Name:   "Jane Smith",
		Email:  "jane.smith@example.com",
		Age:    25,
		Status: "active",
	}

	insertedUser, err := usersRepo.Insert(ctx, newUser)
	if err != nil {
		fmt.Printf("❌ Insert user failed: %v\n", err)
	} else {
		fmt.Printf("✅ User inserted: %s (%s)\n", insertedUser.Name, insertedUser.ID)
	}

	// Example 2: Insert multiple users
	users := []User{
		{Name: "Alice Johnson", Email: "alice@example.com", Age: 28, Status: "active"},
		{Name: "Bob Wilson", Email: "bob@example.com", Age: 35, Status: "active"},
		{Name: "Carol Brown", Email: "carol@example.com", Age: 22, Status: "pending"},
	}

	insertedUsers, err := usersRepo.InsertMany(ctx, users)
	if err != nil {
		fmt.Printf("❌ Insert many users failed: %v\n", err)
	} else {
		fmt.Printf("✅ %d users inserted\n", len(insertedUsers))
	}

	// Example 3: Query with filters
	activeUsers, err := usersRepo.
		Select(ctx, "id", "name", "email", "age").
		Eq("status", "active").
		Gte("age", 25).
		Order("created_at", "desc").
		Limit(10).
		Execute(ctx)
	if err != nil {
		fmt.Printf("❌ Query users failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d active users (25+ years old)\n", len(activeUsers))
		for _, user := range activeUsers {
			fmt.Printf("   - %s (%s) - %d years old\n", user.Name, user.Email, user.Age)
		}
	}

	// Example 4: Text search
	searchResults, err := usersRepo.
		TextSearch("name", "john").
		Execute(ctx)
	if err != nil {
		fmt.Printf("❌ Text search failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d users matching 'john'\n", len(searchResults))
	}

	// Example 5: Update user
	updatedUser, err := usersRepo.
		Eq("email", "jane.smith@example.com").
		Update(ctx, User{
			Age:    26,
			Status: "verified",
		})
	if err != nil {
		fmt.Printf("❌ Update user failed: %v\n", err)
	} else {
		fmt.Printf("✅ User updated: %s (age: %d, status: %s)\n", updatedUser.Name, updatedUser.Age, updatedUser.Status)
	}

	// Example 6: Update multiple records
	err = usersRepo.
		Eq("status", "pending").
		UpdateMany(ctx, map[string]interface{}{
			"status":     "active",
			"updated_at": time.Now(),
		})
	if err != nil {
		fmt.Printf("❌ Update many failed: %v\n", err)
	} else {
		fmt.Println("✅ All pending users activated")
	}

	// Example 7: Count records
	count, err := usersRepo.Count(ctx)
	if err != nil {
		fmt.Printf("❌ Count failed: %v\n", err)
	} else {
		fmt.Printf("✅ Total users: %d\n", count)
	}

	// Example 8: Get single record
	user, err := usersRepo.
		Eq("email", "jane.smith@example.com").
		Single(ctx)
	if err != nil {
		fmt.Printf("❌ Get single user failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found user: %s (%s)\n", user.Name, user.Email)
	}

	// Example 9: Insert a post
	newPost := Post{
		Title:     "My First Blog Post",
		Content:   "This is the content of my first blog post...",
		AuthorID:  user.ID,
		Published: true,
	}

	insertedPost, err := postsRepo.Insert(ctx, newPost)
	if err != nil {
		fmt.Printf("❌ Insert post failed: %v\n", err)
	} else {
		fmt.Printf("✅ Post inserted: %s\n", insertedPost.Title)
	}

	// Example 10: Join-like query (using foreign key)
	posts, err := postsRepo.
		Select(ctx, "id", "title", "content", "author_id").
		Eq("author_id", user.ID).
		Eq("published", true).
		Order("created_at", "desc").
		Execute(ctx)
	if err != nil {
		fmt.Printf("❌ Query posts failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d published posts by %s\n", len(posts), user.Name)
		for _, post := range posts {
			fmt.Printf("   - %s\n", post.Title)
		}
	}

	// Example 11: Delete records
	err = usersRepo.
		Eq("status", "inactive").
		Delete(ctx)
	if err != nil {
		fmt.Printf("❌ Delete failed: %v\n", err)
	} else {
		fmt.Println("✅ Inactive users deleted")
	}
}

func runStorageExamples(ctx context.Context, client supabase.ISupabaseClient) {
	fmt.Println("\n📁 Storage Examples")

	// Example 1: Create a bucket
	bucket, err := client.Storage().CreateBucket(ctx, "avatars", map[string]interface{}{
		"public":          true,
		"file_size_limit": 5242880, // 5MB
	})
	if err != nil {
		fmt.Printf("❌ Create bucket failed: %v\n", err)
	} else {
		fmt.Printf("✅ Bucket created: %s (public: %t)\n", bucket.Name, bucket.Public)
	}

	// Example 2: List buckets
	buckets, err := client.Storage().ListBuckets(ctx)
	if err != nil {
		fmt.Printf("❌ List buckets failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d buckets:\n", len(buckets))
		for _, bucket := range buckets {
			fmt.Printf("   - %s (public: %t)\n", bucket.Name, bucket.Public)
		}
	}

	// Example 3: Upload a file
	fileData := []byte("Hello, this is a test file content!")
	file, err := client.Storage().Upload(ctx, "avatars", "user-123/avatar.txt", fileData, map[string]interface{}{
		"upsert": true,
	})
	if err != nil {
		fmt.Printf("❌ Upload failed: %v\n", err)
	} else {
		fmt.Printf("✅ File uploaded: %s\n", file.Name)
	}

	// Example 4: Get bucket-specific client
	avatarsBucket := client.Storage().From("avatars")

	// Example 5: Upload to specific bucket
	imageData := []byte("fake image data")
	imageFile, err := avatarsBucket.Upload(ctx, "user-456/profile.jpg", imageData, nil)
	if err != nil {
		fmt.Printf("❌ Upload to bucket failed: %v\n", err)
	} else {
		fmt.Printf("✅ Image uploaded: %s\n", imageFile.Name)
	}

	// Example 6: Get public URL
	publicURL := avatarsBucket.GetPublicURL("user-456/profile.jpg")
	fmt.Printf("✅ Public URL: %s\n", publicURL)

	// Example 7: Create signed URL
	signedURL, err := avatarsBucket.CreateSignedURL(ctx, "user-123/avatar.txt", 3600, nil)
	if err != nil {
		fmt.Printf("❌ Create signed URL failed: %v\n", err)
	} else {
		fmt.Printf("✅ Signed URL: %s\n", signedURL)
	}

	// Example 8: Download a file
	downloadedData, err := client.Storage().Download(ctx, "avatars", "user-123/avatar.txt")
	if err != nil {
		fmt.Printf("❌ Download failed: %v\n", err)
	} else {
		fmt.Printf("✅ File downloaded: %d bytes\n", len(downloadedData))
		fmt.Printf("   Content: %s\n", string(downloadedData))
	}

	// Example 9: List files in a bucket
	files, err := client.Storage().List(ctx, "avatars", "user-123/", nil)
	if err != nil {
		fmt.Printf("❌ List files failed: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d files in user-123/:\n", len(files))
		for _, file := range files {
			fmt.Printf("   - %s\n", file.Name)
		}
	}

	// Example 10: Move a file
	err = client.Storage().Move(ctx, "avatars", "user-123/avatar.txt", "user-123/avatar_backup.txt")
	if err != nil {
		fmt.Printf("❌ Move file failed: %v\n", err)
	} else {
		fmt.Println("✅ File moved successfully")
	}

	// Example 11: Copy a file
	err = client.Storage().Copy(ctx, "avatars", "user-456/profile.jpg", "user-456/profile_backup.jpg")
	if err != nil {
		fmt.Printf("❌ Copy file failed: %v\n", err)
	} else {
		fmt.Println("✅ File copied successfully")
	}

	// Example 12: Delete files
	err = client.Storage().Delete(ctx, "avatars", []string{"user-123/avatar_backup.txt", "user-456/profile_backup.jpg"})
	if err != nil {
		fmt.Printf("❌ Delete files failed: %v\n", err)
	} else {
		fmt.Println("✅ Files deleted successfully")
	}
}

func runRealtimeExamples(ctx context.Context, client supabase.ISupabaseClient) {
	fmt.Println("\n⚡ Realtime Examples")

	// Example 1: Connect to realtime
	err := client.Realtime().Connect(ctx)
	if err != nil {
		fmt.Printf("❌ Realtime connect failed: %v\n", err)
		return
	}
	fmt.Println("✅ Connected to realtime")

	// Example 2: Subscribe to table changes
	err = client.Realtime().Subscribe("users", func(message *models.RealtimeMessage) {
		fmt.Printf("📡 User change: %s - %s\n", message.EventType, message.Table)
		fmt.Printf("   Record: %+v\n", message.Record)
	})
	if err != nil {
		fmt.Printf("❌ Subscribe to users failed: %v\n", err)
	} else {
		fmt.Println("✅ Subscribed to users table changes")
	}

	// Example 3: Get a channel and subscribe to specific events
	channel := client.Realtime().Channel("posts")

	channel.On("INSERT", func(message *models.RealtimeMessage) {
		fmt.Println("📝 New post inserted!")
	})

	channel.On("UPDATE", func(message *models.RealtimeMessage) {
		fmt.Println("✏️ Post updated!")
	})

	channel.On("DELETE", func(message *models.RealtimeMessage) {
		fmt.Println("🗑️ Post deleted!")
	})

	err = channel.Subscribe(func(message *models.RealtimeMessage) {
		fmt.Printf("📡 Received post message: %+v\n", message)
	})
	if err != nil {
		fmt.Printf("❌ Subscribe to posts channel failed: %v\n", err)
	} else {
		fmt.Println("✅ Subscribed to posts channel")
	}

	// Example 4: Broadcast a message
	err = client.Realtime().Broadcast("chat", "message", map[string]interface{}{
		"user":      "john",
		"text":      "Hello, everyone!",
		"timestamp": time.Now(),
	})
	if err != nil {
		fmt.Printf("❌ Broadcast failed: %v\n", err)
	} else {
		fmt.Println("✅ Message broadcasted to chat channel")
	}

	// Example 5: Broadcast to self only
	err = client.Realtime().BroadcastToSelf("chat", "typing", map[string]interface{}{
		"user":   "john",
		"typing": true,
	})
	if err != nil {
		fmt.Printf("❌ Broadcast to self failed: %v\n", err)
	} else {
		fmt.Println("✅ Typing indicator sent to self")
	}

	// Example 6: Track presence
	err = client.Realtime().Track(map[string]interface{}{
		"user_id":   "user-123",
		"status":    "online",
		"last_seen": time.Now(),
		"name":      "John Doe",
	})
	if err != nil {
		fmt.Printf("❌ Track presence failed: %v\n", err)
	} else {
		fmt.Println("✅ Presence tracked")
	}

	// Example 7: Get presence for a channel
	presence, err := client.Realtime().Presence("chat")
	if err != nil {
		fmt.Printf("❌ Get presence failed: %v\n", err)
	} else {
		fmt.Printf("✅ Online users: %+v\n", presence)
	}

	// Example 8: Untrack presence
	err = client.Realtime().Untrack()
	if err != nil {
		fmt.Printf("❌ Untrack presence failed: %v\n", err)
	} else {
		fmt.Println("✅ Presence untracked")
	}

	// Clean up
	time.Sleep(2 * time.Second) // Give some time for messages to be processed
	client.Realtime().RemoveAllChannels()
	client.Realtime().Disconnect()
	fmt.Println("✅ Realtime connection closed")
}

func runEdgeFunctionExamples(ctx context.Context, client supabase.ISupabaseClient) {
	fmt.Println("\n🚀 Edge Function Examples")

	// Example 1: Call an edge function
	result, err := client.EdgeFunction("hello-world", map[string]interface{}{
		"name": "John",
		"age":  30,
		"city": "New York",
	})
	if err != nil {
		fmt.Printf("❌ Edge function call failed: %v\n", err)
	} else {
		fmt.Printf("✅ Edge function result: %+v\n", result)
	}

	// Example 2: Call a function with different parameters
	result, err = client.Function("process-data", map[string]interface{}{
		"data": []string{"item1", "item2", "item3"},
		"options": map[string]interface{}{
			"sort":   true,
			"filter": "active",
			"limit":  10,
		},
	})
	if err != nil {
		fmt.Printf("❌ Function call failed: %v\n", err)
	} else {
		fmt.Printf("✅ Function result: %+v\n", result)
	}
}
