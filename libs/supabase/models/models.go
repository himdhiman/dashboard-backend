package models

import (
	"time"
)

// User represents a Supabase user
type User struct {
	ID                     string                 `json:"id"`
	Email                  string                 `json:"email"`
	Phone                  string                 `json:"phone"`
	EmailConfirmedAt       *time.Time             `json:"email_confirmed_at"`
	PhoneConfirmedAt       *time.Time             `json:"phone_confirmed_at"`
	LastSignInAt           *time.Time             `json:"last_sign_in_at"`
	AppMetadata            map[string]interface{} `json:"app_metadata"`
	UserMetadata           map[string]interface{} `json:"user_metadata"`
	Identities             []Identity             `json:"identities"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
	ConfirmationSentAt     *time.Time             `json:"confirmation_sent_at"`
	RecoverySentAt         *time.Time             `json:"recovery_sent_at"`
	EmailChange            string                 `json:"email_change"`
	EmailChangeSentAt      *time.Time             `json:"email_change_sent_at"`
	PhoneChange            string                 `json:"phone_change"`
	PhoneChangeSentAt      *time.Time             `json:"phone_change_sent_at"`
	ReauthenticationSentAt *time.Time             `json:"reauthentication_sent_at"`
	Role                   string                 `json:"role"`
	Aud                    string                 `json:"aud"`
	BannedUntil            *time.Time             `json:"banned_until"`
	ConfirmedAt            *time.Time             `json:"confirmed_at"`
	InvitedAt              *time.Time             `json:"invited_at"`
}

// Identity represents a user identity
type Identity struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	IdentityData map[string]interface{} `json:"identity_data"`
	Provider     string                 `json:"provider"`
	LastSignInAt time.Time              `json:"last_sign_in_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Session represents a user session
type Session struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	User         User      `json:"user"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Session *Session `json:"session"`
	User    *User    `json:"user"`
}

// SignUpRequest represents a sign up request
type SignUpRequest struct {
	Email    string                 `json:"email"`
	Password string                 `json:"password"`
	Phone    string                 `json:"phone,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// SignInRequest represents a sign in request
type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
}

// PasswordResetRequest represents a password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordUpdateRequest represents a password update request
type PasswordUpdateRequest struct {
	Password string `json:"password"`
}

// UserUpdateRequest represents a user update request
type UserUpdateRequest struct {
	Email    string                 `json:"email,omitempty"`
	Phone    string                 `json:"phone,omitempty"`
	Password string                 `json:"password,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// DatabaseResponse represents a database response
type DatabaseResponse struct {
	Data       interface{} `json:"data"`
	Error      *Error      `json:"error"`
	Count      *int        `json:"count"`
	Status     int         `json:"status"`
	StatusText string      `json:"statusText"`
}

// Error represents a Supabase error
type Error struct {
	Message string `json:"message"`
	Details string `json:"details"`
	Hint    string `json:"hint"`
	Code    string `json:"code"`
}

// QueryOptions represents query options for database operations
type QueryOptions struct {
	Select  string            `json:"select,omitempty"`
	Order   string            `json:"order,omitempty"`
	Range   string            `json:"range,omitempty"`
	Limit   int               `json:"limit,omitempty"`
	Offset  int               `json:"offset,omitempty"`
	Filters map[string]string `json:"filters,omitempty"`
}

// StorageFile represents a storage file
type StorageFile struct {
	Name           string                 `json:"name"`
	BucketID       string                 `json:"bucket_id"`
	Owner          string                 `json:"owner"`
	ID             string                 `json:"id"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedAt      time.Time              `json:"created_at"`
	LastAccessedAt time.Time              `json:"last_accessed_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// StorageBucket represents a storage bucket
type StorageBucket struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Owner            string    `json:"owner"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Public           bool      `json:"public"`
	FileSizeLimit    *int      `json:"file_size_limit,omitempty"`
	AllowedMimeTypes []string  `json:"allowed_mime_types,omitempty"`
}

// RealtimeMessage represents a realtime message
type RealtimeMessage struct {
	Type            string      `json:"type"`
	Table           string      `json:"table"`
	Schema          string      `json:"schema"`
	Record          interface{} `json:"record"`
	OldRecord       interface{} `json:"old_record"`
	Columns         []string    `json:"columns"`
	CommitTimestamp string      `json:"commit_timestamp"`
	EventType       string      `json:"event_type"`
	Errors          []string    `json:"errors"`
}

// FunctionResponse represents a function response
type FunctionResponse struct {
	Data       interface{} `json:"data"`
	Error      *Error      `json:"error"`
	Status     int         `json:"status"`
	StatusText string      `json:"statusText"`
}

// Pagination represents pagination information
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
}

// Filter represents a database filter
type Filter struct {
	Column   string      `json:"column"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Sort represents a database sort
type Sort struct {
	Column    string `json:"column"`
	Direction string `json:"direction"`
}
