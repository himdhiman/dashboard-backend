package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/himdhiman/dashboard-backend/libs/logger"
	"github.com/himdhiman/dashboard-backend/libs/supabase/config"
	"github.com/himdhiman/dashboard-backend/libs/supabase/errors"
	"github.com/himdhiman/dashboard-backend/libs/supabase/models"
)

// IRealtimeClient defines the interface for realtime operations
type IRealtimeClient interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	Close() error
	IsConnected() bool

	// Channel operations
	Channel(channel string) IRealtimeChannel
	RemoveChannel(channel string) error
	RemoveAllChannels() error

	// Subscription management
	Subscribe(channel string, callback func(*models.RealtimeMessage)) error
	Unsubscribe(channel string) error

	// Broadcasting
	Broadcast(channel string, event string, payload interface{}) error
	BroadcastToSelf(channel string, event string, payload interface{}) error

	// Presence
	Track(payload interface{}) error
	Untrack() error
	Presence(channel string) (map[string]interface{}, error)
}

// IRealtimeChannel defines the interface for channel-specific operations
type IRealtimeChannel interface {
	// Subscription
	On(event string, callback func(*models.RealtimeMessage)) IRealtimeChannel
	Subscribe(callback func(*models.RealtimeMessage)) error
	Unsubscribe() error

	// Broadcasting
	Send(event string, payload interface{}) error
	SendToSelf(event string, payload interface{}) error

	// Presence
	Track(payload interface{}) error
	Untrack() error
	Presence() (map[string]interface{}, error)
}

// RealtimeClient represents a realtime client
type RealtimeClient struct {
	IRealtimeClient
	config    *config.Config
	logger    logger.ILogger
	conn      *websocket.Conn
	channels  map[string]*RealtimeChannel
	mutex     sync.RWMutex
	connected bool
	messageID int
	callbacks map[string]func(*models.RealtimeMessage)
}

// RealtimeChannel represents a realtime channel
type RealtimeChannel struct {
	IRealtimeChannel
	client *RealtimeClient
	name   string
	events map[string]func(*models.RealtimeMessage)
	mutex  sync.RWMutex
}

// NewRealtimeClient creates a new realtime client
func NewRealtimeClient(cfg *config.Config, logger logger.ILogger) (IRealtimeClient, error) {
	return &RealtimeClient{
		config:    cfg,
		logger:    logger,
		channels:  make(map[string]*RealtimeChannel),
		callbacks: make(map[string]func(*models.RealtimeMessage)),
	}, nil
}

// Connect establishes a WebSocket connection
func (r *RealtimeClient) Connect(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.connected {
		return nil
	}

	// Parse the URL and convert to WebSocket URL
	parsedURL, err := url.Parse(r.config.URL)
	if err != nil {
		return errors.NewError(err, "failed to parse URL")
	}

	// Convert to WebSocket URL
	wsURL := fmt.Sprintf("wss://%s/realtime/v1/websocket", parsedURL.Host)

	// Add query parameters
	query := url.Values{}
	query.Set("apikey", r.config.Key)
	wsURL += "?" + query.Encode()

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return errors.NewError(err, "failed to connect to WebSocket")
	}

	r.conn = conn
	r.connected = true

	// Start message handling goroutine
	go r.handleMessages()

	r.logger.Info("Realtime client connected successfully")
	return nil
}

// Disconnect disconnects from the WebSocket
func (r *RealtimeClient) Disconnect() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.connected {
		return nil
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return errors.NewError(err, "failed to close WebSocket connection")
		}
	}

	r.connected = false
	r.logger.Info("Realtime client disconnected")
	return nil
}

// Close closes the realtime client
func (r *RealtimeClient) Close() error {
	// Remove all channels
	if err := r.RemoveAllChannels(); err != nil {
		return err
	}

	// Disconnect
	return r.Disconnect()
}

// IsConnected returns whether the client is connected
func (r *RealtimeClient) IsConnected() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.connected
}

// Channel returns a channel instance
func (r *RealtimeClient) Channel(channel string) IRealtimeChannel {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if ch, exists := r.channels[channel]; exists {
		return ch
	}

	ch := &RealtimeChannel{
		client: r,
		name:   channel,
		events: make(map[string]func(*models.RealtimeMessage)),
	}

	r.channels[channel] = ch
	return ch
}

// RemoveChannel removes a channel
func (r *RealtimeClient) RemoveChannel(channel string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if ch, exists := r.channels[channel]; exists {
		// Unsubscribe from the channel
		if err := ch.Unsubscribe(); err != nil {
			return err
		}

		delete(r.channels, channel)
		r.logger.Info("Channel removed", "channel", channel)
	}

	return nil
}

// RemoveAllChannels removes all channels
func (r *RealtimeClient) RemoveAllChannels() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for channel := range r.channels {
		if err := r.channels[channel].Unsubscribe(); err != nil {
			return err
		}
	}

	r.channels = make(map[string]*RealtimeChannel)
	r.logger.Info("All channels removed")
	return nil
}

// Subscribe subscribes to a channel
func (r *RealtimeClient) Subscribe(channel string, callback func(*models.RealtimeMessage)) error {
	ch := r.Channel(channel).(*RealtimeChannel)
	return ch.Subscribe(callback)
}

// Unsubscribe unsubscribes from a channel
func (r *RealtimeClient) Unsubscribe(channel string) error {
	return r.RemoveChannel(channel)
}

// Broadcast broadcasts a message to a channel
func (r *RealtimeClient) Broadcast(channel string, event string, payload interface{}) error {
	ch := r.Channel(channel).(*RealtimeChannel)
	return ch.Send(event, payload)
}

// BroadcastToSelf broadcasts a message to self
func (r *RealtimeClient) BroadcastToSelf(channel string, event string, payload interface{}) error {
	ch := r.Channel(channel).(*RealtimeChannel)
	return ch.SendToSelf(event, payload)
}

// Track tracks presence
func (r *RealtimeClient) Track(payload interface{}) error {
	message := map[string]interface{}{
		"type":    "track",
		"payload": payload,
	}

	return r.sendMessage(message)
}

// Untrack untracks presence
func (r *RealtimeClient) Untrack() error {
	message := map[string]interface{}{
		"type": "untrack",
	}

	return r.sendMessage(message)
}

// Presence gets presence for a channel
func (r *RealtimeClient) Presence(channel string) (map[string]interface{}, error) {
	message := map[string]interface{}{
		"type":    "presence",
		"channel": channel,
	}

	// This is a simplified implementation
	// In a real implementation, you would wait for the response
	return make(map[string]interface{}), r.sendMessage(message)
}

// handleMessages handles incoming WebSocket messages
func (r *RealtimeClient) handleMessages() {
	for {
		if !r.IsConnected() {
			break
		}

		_, message, err := r.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				r.logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		var realtimeMessage models.RealtimeMessage
		if err := json.Unmarshal(message, &realtimeMessage); err != nil {
			r.logger.Error("Failed to unmarshal realtime message", "error", err)
			continue
		}

		// Handle the message
		r.handleMessage(&realtimeMessage)
	}
}

// handleMessage handles a single realtime message
func (r *RealtimeClient) handleMessage(message *models.RealtimeMessage) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Check if we have a callback for this channel
	if callback, exists := r.callbacks[message.Table]; exists {
		callback(message)
	}

	// Check if we have a channel for this table
	if ch, exists := r.channels[message.Table]; exists {
		ch.handleMessage(message)
	}
}

// sendMessage sends a message through the WebSocket
func (r *RealtimeClient) sendMessage(message map[string]interface{}) error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.connected || r.conn == nil {
		return errors.ErrConnectionClosed
	}

	// Add message ID
	r.messageID++
	message["ref"] = r.messageID

	data, err := json.Marshal(message)
	if err != nil {
		return errors.NewError(err, "failed to marshal message")
	}

	if err := r.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return errors.NewError(err, "failed to send message")
	}

	return nil
}

// RealtimeChannel methods
func (rc *RealtimeChannel) On(event string, callback func(*models.RealtimeMessage)) IRealtimeChannel {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	rc.events[event] = callback
	return rc
}

func (rc *RealtimeChannel) Subscribe(callback func(*models.RealtimeMessage)) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// Store the callback
	rc.client.callbacks[rc.name] = callback

	// Send subscription message
	message := map[string]interface{}{
		"type":    "subscribe",
		"channel": rc.name,
	}

	return rc.client.sendMessage(message)
}

func (rc *RealtimeChannel) Unsubscribe() error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	// Remove the callback
	delete(rc.client.callbacks, rc.name)

	// Send unsubscription message
	message := map[string]interface{}{
		"type":    "unsubscribe",
		"channel": rc.name,
	}

	return rc.client.sendMessage(message)
}

func (rc *RealtimeChannel) Send(event string, payload interface{}) error {
	message := map[string]interface{}{
		"type":    "broadcast",
		"channel": rc.name,
		"event":   event,
		"payload": payload,
	}

	return rc.client.sendMessage(message)
}

func (rc *RealtimeChannel) SendToSelf(event string, payload interface{}) error {
	message := map[string]interface{}{
		"type":    "broadcast",
		"channel": rc.name,
		"event":   event,
		"payload": payload,
		"self":    true,
	}

	return rc.client.sendMessage(message)
}

func (rc *RealtimeChannel) Track(payload interface{}) error {
	message := map[string]interface{}{
		"type":    "track",
		"channel": rc.name,
		"payload": payload,
	}

	return rc.client.sendMessage(message)
}

func (rc *RealtimeChannel) Untrack() error {
	message := map[string]interface{}{
		"type":    "untrack",
		"channel": rc.name,
	}

	return rc.client.sendMessage(message)
}

func (rc *RealtimeChannel) Presence() (map[string]interface{}, error) {
	message := map[string]interface{}{
		"type":    "presence",
		"channel": rc.name,
	}

	// This is a simplified implementation
	// In a real implementation, you would wait for the response
	return make(map[string]interface{}), rc.client.sendMessage(message)
}

// handleMessage handles a message for this channel
func (rc *RealtimeChannel) handleMessage(message *models.RealtimeMessage) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	// Check if we have a callback for this event
	if callback, exists := rc.events[message.EventType]; exists {
		callback(message)
	}
}
