package signaling

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestShutdown(t *testing.T) {
	server := NewServer()
	slug := "test-room"

	mockHostConn := &MockWebSocketConn{}
	mockGuestConn := &MockWebSocketConn{}

	host := &Participant{
		ID:   "host1",
		Conn: mockHostConn,
		Role: RoleHost,
	}

	guest := &Participant{
		ID:   "guest1",
		Conn: mockGuestConn,
		Role: RoleGuest,
	}

	// Create room and add participants
	server.rooms[slug] = NewRoom(slug)
	server.rooms[slug].AddParticipant(host)
	server.rooms[slug].AddParticipant(guest)

	assert.Equal(t, 1, len(server.rooms))

	// Wait for Close calls on both connections
	mockHostConn.On("Close").Return(nil).Once()
	mockGuestConn.On("Close").Return(nil).Once()

	server.Shutdown()

	// Check that all rooms are cleared after shutdown
	assert.Equal(t, 0, len(server.rooms))

	mockHostConn.AssertExpectations(t)
	mockGuestConn.AssertExpectations(t)
}

func TestWebSocketConnection(t *testing.T) {
	server := NewServer()

	testServer := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") +
		"?slug=test-room&role=host&name=TestHost"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	stats := server.GetRoomStats("test-room")
	assert.NotNil(t, stats)
	assert.Equal(t, "test-room", stats["slug"])
	assert.Equal(t, true, stats["has_host"])
}

func TestNewServer(t *testing.T) {
	server := NewServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.rooms)
	assert.Equal(t, 0, len(server.rooms))
}

func TestWebSocketUpgrade(t *testing.T) {
	server := NewServer()
	testServer := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer testServer.Close()

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") +
		"?slug=test-room&role=host&name=TestHost"

	// Connect via WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	// Check room created
	stats := server.GetRoomStats("test-room")
	assert.NotNil(t, stats)
	assert.Equal(t, "test-room", stats["slug"])
	assert.Equal(t, true, stats["has_host"])
}

func TestJoinRoomAsGuest(t *testing.T) {
	server := NewServer()

	testServer := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer testServer.Close()

	// Connect as host
	hostURL := "ws" + strings.TrimPrefix(testServer.URL, "http") +
		"?slug=test-room&role=host&name=Host"
	hostConn, _, err := websocket.DefaultDialer.Dial(hostURL, nil)
	assert.NoError(t, err)
	defer hostConn.Close()

	time.Sleep(50 * time.Millisecond)

	// Connect as guest
	guestURL := "ws" + strings.TrimPrefix(testServer.URL, "http") +
		"?slug=test-room&role=guest&name=Guest"
	guestConn, _, err := websocket.DefaultDialer.Dial(guestURL, nil)
	assert.NoError(t, err)
	defer guestConn.Close()

	time.Sleep(50 * time.Millisecond)

	// Check room stats
	stats := server.GetRoomStats("test-room")
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats["guests_count"])
}

func TestInvalidWebSocketParams(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"missing slug", "role=host&name=Test", http.StatusBadRequest},
		{"missing role", "slug=room&name=Test", http.StatusBadRequest},
		{"invalid role", "slug=room&role=invalid&name=Test", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ws?"+tt.query, nil)
			w := httptest.NewRecorder()

			server.HandleWebSocket(w, req)
			assert.Equal(t, tt.expected, w.Code)
		})
	}
}

func TestGetRoomStats(t *testing.T) {
	server := NewServer()

	// Room does not exist
	stats := server.GetRoomStats("nonexistent")
	assert.Nil(t, stats)

	// Create room via WebSocket
	testServer := httptest.NewServer(http.HandlerFunc(server.HandleWebSocket))
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http") +
		"?slug=test-room&role=host&name=Host"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	stats = server.GetRoomStats("test-room")
	assert.NotNil(t, stats)
	assert.Equal(t, "test-room", stats["slug"])
	assert.Equal(t, true, stats["has_host"])
	assert.Equal(t, 0, stats["guests_count"])
}
