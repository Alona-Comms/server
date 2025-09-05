package signaling

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType определяет типы сообщений в signaling
type MessageType string

const (
	MessageTypeJoin         MessageType = "join"
	MessageTypeLeave        MessageType = "leave"
	MessageTypeKnock        MessageType = "knock"         // Гость "стучится"
	MessageTypeAllow        MessageType = "allow"         // Хост разрешает
	MessageTypeDeny         MessageType = "deny"          // Хост отклоняет
	MessageTypeOffer        MessageType = "offer"         // WebRTC offer
	MessageTypeAnswer       MessageType = "answer"        // WebRTC answer
	MessageTypeICECandidate MessageType = "ice_candidate" // ICE candidate
	MessageTypeParticipants MessageType = "participants"  // Список участников
	MessageTypeError        MessageType = "error"
)

// ParticipantRole определяет роль участника
type ParticipantRole string

const (
	RoleHost  ParticipantRole = "host"
	RoleGuest ParticipantRole = "guest"
)

// ParticipantStatus определяет статус участника
type ParticipantStatus string

const (
	StatusConnected    ParticipantStatus = "connected"
	StatusKnocking     ParticipantStatus = "knocking"
	StatusInRoom       ParticipantStatus = "in_room"
	StatusDisconnected ParticipantStatus = "disconnected"
)

type WebSocketConnInterface interface {
	WriteJSON(v interface{}) error
	ReadJSON(v interface{}) error
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}

// WebSocketConnWrapper оборачивает *websocket.Conn для реализации интерфейса
type WebSocketConnWrapper struct {
	*websocket.Conn
}

// NewWebSocketConnWrapper создает wrapper для websocket.Conn
func NewWebSocketConnWrapper(conn *websocket.Conn) *WebSocketConnWrapper {
	return &WebSocketConnWrapper{Conn: conn}
}

type Participant struct {
	ID       string                 `json:"id"`
	Conn     WebSocketConnInterface `json:"-"` // Изменен тип на интерфейс
	Role     ParticipantRole        `json:"role"`
	Status   ParticipantStatus      `json:"status"`
	Name     string                 `json:"name,omitempty"`
	JoinedAt time.Time              `json:"joined_at"`
}

// Room представляет комнату для звонков
type Room struct {
	Slug      string                  `json:"slug"`
	Host      *Participant            `json:"host,omitempty"`
	Guests    map[string]*Participant `json:"guests"`
	CreatedAt time.Time               `json:"created_at"`
	mutex     sync.RWMutex
}

// Message представляет сообщение в signaling
type Message struct {
	Type      MessageType `json:"type"`
	From      string      `json:"from,omitempty"`
	To        string      `json:"to,omitempty"`
	Slug      string      `json:"slug,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// JoinData содержит данные для присоединения
type JoinData struct {
	Name string          `json:"name"`
	Role ParticipantRole `json:"role"`
}

// ParticipantsData содержит список участников
type ParticipantsData struct {
	Host   *Participant            `json:"host,omitempty"`
	Guests map[string]*Participant `json:"guests"`
	Count  int                     `json:"count"`
}

// ErrorData содержит информацию об ошибке
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
