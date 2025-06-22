package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/errors"
	"github.com/himdhiman/dashboard-backend/libs/supabase/models"
)

// IAuthClient defines the interface for authentication operations
type IAuthClient interface {
	// User management
	SignUp(ctx context.Context, req *models.SignUpRequest) (*models.AuthResponse, error)
	SignIn(ctx context.Context, req *models.SignInRequest) (*models.AuthResponse, error)
	SignOut(ctx context.Context) error
	SignOutAll(ctx context.Context) error

	// Password management
	ResetPassword(ctx context.Context, req *models.PasswordResetRequest) error
	UpdatePassword(ctx context.Context, req *models.PasswordUpdateRequest) error

	// User management
	GetUser(ctx context.Context) (*models.User, error)
	UpdateUser(ctx context.Context, req *models.UserUpdateRequest) (*models.User, error)
	DeleteUser(ctx context.Context) error

	// Session management
	GetSession(ctx context.Context) (*models.Session, error)
	RefreshSession(ctx context.Context, refreshToken string) (*models.AuthResponse, error)

	// OAuth
	SignInWithOAuth(ctx context.Context, provider string, options map[string]interface{}) (string, error)

	// Magic link
	SignInWithMagicLink(ctx context.Context, email string, options map[string]interface{}) error

	// Phone auth
	SignInWithPhone(ctx context.Context, phone, password string) (*models.AuthResponse, error)
	SignUpWithPhone(ctx context.Context, phone, password string, options map[string]interface{}) (*models.AuthResponse, error)
}

// AuthClient represents an authentication client
type AuthClient struct {
	IAuthClient
	config     *config.Config
	httpClient *http.Client
	logger     logger.ILogger
}

// NewAuthClient creates a new authentication client
func NewAuthClient(cfg *config.Config, httpClient *http.Client, logger logger.ILogger) (IAuthClient, error) {
	return &AuthClient{
		config:     cfg,
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

// SignUp registers a new user
func (a *AuthClient) SignUp(ctx context.Context, req *models.SignUpRequest) (*models.AuthResponse, error) {
	url := a.config.URL + "/auth/v1/signup"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.NewError(err, "failed to marshal signup request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewError(err, "failed to create signup request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute signup request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("signup failed with status: %d", resp.StatusCode)
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, errors.NewError(err, "failed to decode signup response")
	}

	a.logger.Info("User signed up successfully", "email", req.Email)
	return &authResp, nil
}

// SignIn authenticates a user
func (a *AuthClient) SignIn(ctx context.Context, req *models.SignInRequest) (*models.AuthResponse, error) {
	url := a.config.URL + "/auth/v1/token?grant_type=password"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.NewError(err, "failed to marshal signin request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewError(err, "failed to create signin request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute signin request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("signin failed with status: %d", resp.StatusCode)
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, errors.NewError(err, "failed to decode signin response")
	}

	a.logger.Info("User signed in successfully", "email", req.Email)
	return &authResp, nil
}

// SignOut signs out the current user
func (a *AuthClient) SignOut(ctx context.Context) error {
	url := a.config.URL + "/auth/v1/logout"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return errors.NewError(err, "failed to create signout request")
	}

	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute signout request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("signout failed with status: %d", resp.StatusCode)
	}

	a.logger.Info("User signed out successfully")
	return nil
}

// SignOutAll signs out all sessions for the current user
func (a *AuthClient) SignOutAll(ctx context.Context) error {
	url := a.config.URL + "/auth/v1/logout?scope=global"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return errors.NewError(err, "failed to create signout all request")
	}

	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute signout all request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("signout all failed with status: %d", resp.StatusCode)
	}

	a.logger.Info("User signed out from all sessions successfully")
	return nil
}

// ResetPassword sends a password reset email
func (a *AuthClient) ResetPassword(ctx context.Context, req *models.PasswordResetRequest) error {
	url := a.config.URL + "/auth/v1/recover"

	body, err := json.Marshal(req)
	if err != nil {
		return errors.NewError(err, "failed to marshal password reset request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create password reset request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute password reset request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("password reset failed with status: %d", resp.StatusCode)
	}

	a.logger.Info("Password reset email sent successfully", "email", req.Email)
	return nil
}

// UpdatePassword updates the user's password
func (a *AuthClient) UpdatePassword(ctx context.Context, req *models.PasswordUpdateRequest) error {
	url := a.config.URL + "/auth/v1/user"

	body, err := json.Marshal(req)
	if err != nil {
		return errors.NewError(err, "failed to marshal password update request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create password update request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute password update request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("password update failed with status: %d", resp.StatusCode)
	}

	a.logger.Info("Password updated successfully")
	return nil
}

// GetUser retrieves the current user
func (a *AuthClient) GetUser(ctx context.Context) (*models.User, error) {
	url := a.config.URL + "/auth/v1/user"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create get user request")
	}

	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute get user request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("get user failed with status: %d", resp.StatusCode)
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.NewError(err, "failed to decode user response")
	}

	return &user, nil
}

// UpdateUser updates the current user
func (a *AuthClient) UpdateUser(ctx context.Context, req *models.UserUpdateRequest) (*models.User, error) {
	url := a.config.URL + "/auth/v1/user"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.NewError(err, "failed to marshal user update request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewError(err, "failed to create user update request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute user update request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("user update failed with status: %d", resp.StatusCode)
	}

	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, errors.NewError(err, "failed to decode user update response")
	}

	a.logger.Info("User updated successfully")
	return &user, nil
}

// DeleteUser deletes the current user
func (a *AuthClient) DeleteUser(ctx context.Context) error {
	url := a.config.URL + "/auth/v1/user"

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return errors.NewError(err, "failed to create delete user request")
	}

	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute delete user request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("delete user failed with status: %d", resp.StatusCode)
	}

	a.logger.Info("User deleted successfully")
	return nil
}

// GetSession retrieves the current session
func (a *AuthClient) GetSession(ctx context.Context) (*models.Session, error) {
	url := a.config.URL + "/auth/v1/user"

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewError(err, "failed to create get session request")
	}

	httpReq.Header.Set("apikey", a.config.Key)
	// Note: Authorization header should be set with the user's access token

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute get session request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("get session failed with status: %d", resp.StatusCode)
	}

	// Extract session info from response headers or body
	// This is a simplified implementation
	session := &models.Session{
		AccessToken: resp.Header.Get("Authorization"),
		TokenType:   "bearer",
		ExpiresIn:   3600,
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	return session, nil
}

// RefreshSession refreshes the current session
func (a *AuthClient) RefreshSession(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	url := a.config.URL + "/auth/v1/token?grant_type=refresh_token"

	req := map[string]string{
		"refresh_token": refreshToken,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.NewError(err, "failed to marshal refresh session request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.NewError(err, "failed to create refresh session request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewError(err, "failed to execute refresh session request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewErrorf("refresh session failed with status: %d", resp.StatusCode)
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, errors.NewError(err, "failed to decode refresh session response")
	}

	a.logger.Info("Session refreshed successfully")
	return &authResp, nil
}

// SignInWithOAuth initiates OAuth sign in
func (a *AuthClient) SignInWithOAuth(ctx context.Context, provider string, options map[string]interface{}) (string, error) {
	url := a.config.URL + "/auth/v1/authorize"

	params := map[string]interface{}{
		"provider": provider,
	}

	// Merge options with default params
	for k, v := range options {
		params[k] = v
	}

	body, err := json.Marshal(params)
	if err != nil {
		return "", errors.NewError(err, "failed to marshal OAuth request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", errors.NewError(err, "failed to create OAuth request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", errors.NewError(err, "failed to execute OAuth request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.NewErrorf("OAuth signin failed with status: %d", resp.StatusCode)
	}

	// Extract authorization URL from response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", errors.NewError(err, "failed to decode OAuth response")
	}

	authURL, ok := result["url"].(string)
	if !ok {
		return "", errors.NewError(err, "invalid OAuth response format")
	}

	a.logger.Info("OAuth signin initiated successfully", "provider", provider)
	return authURL, nil
}

// SignInWithMagicLink sends a magic link to the user's email
func (a *AuthClient) SignInWithMagicLink(ctx context.Context, email string, options map[string]interface{}) error {
	url := a.config.URL + "/auth/v1/magiclink"

	params := map[string]interface{}{
		"email": email,
	}

	// Merge options with default params
	for k, v := range options {
		params[k] = v
	}

	body, err := json.Marshal(params)
	if err != nil {
		return errors.NewError(err, "failed to marshal magic link request")
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return errors.NewError(err, "failed to create magic link request")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apikey", a.config.Key)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return errors.NewError(err, "failed to execute magic link request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.NewErrorf("magic link signin failed with status: %d", resp.StatusCode)
	}

	a.logger.Info("Magic link sent successfully", "email", email)
	return nil
}

// SignInWithPhone signs in a user with phone number
func (a *AuthClient) SignInWithPhone(ctx context.Context, phone, password string) (*models.AuthResponse, error) {
	req := &models.SignInRequest{
		Phone:    phone,
		Password: password,
	}

	return a.SignIn(ctx, req)
}

// SignUpWithPhone signs up a user with phone number
func (a *AuthClient) SignUpWithPhone(ctx context.Context, phone, password string, options map[string]interface{}) (*models.AuthResponse, error) {
	req := &models.SignUpRequest{
		Phone:    phone,
		Password: password,
		Data:     options,
	}

	return a.SignUp(ctx, req)
}
