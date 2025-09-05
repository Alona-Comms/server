package signaling

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	rooms    map[string]*Room
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		rooms: make(map[string]*Room),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO : check origin in production
				return true
			},
		},
	}
}

func generateParticipantID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	roleStr := r.URL.Query().Get("role")
	name := r.URL.Query().Get("name")

	if slug == "" || roleStr == "" {
		http.Error(w, "Missing slug or role", http.StatusBadRequest)
		return
	}

	var role ParticipantRole
	switch roleStr {
	case "host":
		role = RoleHost
	case "guest":
		role = RoleGuest
	default:
		http.Error(w, "Invalid role", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	participant := &Participant{
		ID:       generateParticipantID(),
		Conn:     conn,
		Role:     role,
		Status:   StatusConnected,
		Name:     name,
		JoinedAt: time.Now(),
	}

	s.joinRoom(slug, participant)

	go s.handleConnection(slug, participant)
}

func (s *Server) joinRoom(slug string, participant *Participant) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[slug]; !exists {
		s.rooms[slug] = NewRoom(slug)
	}

	room := s.rooms[slug]
	err := room.AddParticipant(participant)
	if err != nil {
		log.Printf("Failed to add participant: %v", err)
		participant.Conn.WriteJSON(&Message{
			Type: MessageTypeError,
			Data: ErrorData{
				Code:    "JOIN_FAILED",
				Message: err.Error(),
			},
			Timestamp: time.Now(),
		})
		participant.Conn.Close()
		return
	}

	joinMessage := &Message{
		Type:      MessageTypeJoin,
		From:      participant.ID,
		Slug:      slug,
		Data:      participant,
		Timestamp: time.Now(),
	}

	if participant.Role == RoleGuest {
		knockMessage := &Message{
			Type:      MessageTypeKnock,
			From:      participant.ID,
			Slug:      slug,
			Data:      participant,
			Timestamp: time.Now(),
		}
		room.BroadcastToHost(knockMessage)
	} else {
		room.BroadcastToAll(joinMessage, participant.ID)
	}

	participant.Conn.WriteJSON(&Message{
		Type:      MessageTypeParticipants,
		Slug:      slug,
		Data:      room.GetParticipantsData(),
		Timestamp: time.Now(),
	})
}

func (s *Server) handleConnection(slug string, participant *Participant) {
	defer s.leaveRoom(slug, participant)

	for {
		var message Message
		err := participant.Conn.ReadJSON(&message)
		if err != nil {
			log.Printf("Read message error: %v", err)
			break
		}

		message.From = participant.ID
		message.Slug = slug
		message.Timestamp = time.Now()

		s.handleMessage(slug, participant, &message)
	}
}

func (s *Server) leaveRoom(slug string, participant *Participant) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, exists := s.rooms[slug]
	if !exists {
		return
	}

	room.RemovePublicKey(participant.ID)

	room.RemoveParticipant(participant.ID)
	participant.Conn.Close()

	leaveMessage := &Message{
		Type:      MessageTypeLeave,
		From:      participant.ID,
		Slug:      slug,
		Timestamp: time.Now(),
	}
	room.BroadcastToAll(leaveMessage, participant.ID)

	room.BroadcastPublicKeys(participant.ID)

	if room.IsEmpty() {
		delete(s.rooms, slug)
		log.Printf("Room %s deleted (empty)", slug)
	}
}

func (s *Server) GetRoomStats(slug string) map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	room, exists := s.rooms[slug]
	if !exists {
		return nil
	}

	return map[string]interface{}{
		"slug":         room.Slug,
		"participants": room.GetParticipantsData(),
		"created_at":   room.CreatedAt,
		"has_host":     room.Host != nil,
		"guests_count": len(room.Guests),
	}
}

func (s *Server) Shutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, room := range s.rooms {
		if room.Host != nil {
			room.Host.Conn.Close()
		}
		for _, guest := range room.Guests {
			guest.Conn.Close()
		}
	}
	s.rooms = make(map[string]*Room)
}
