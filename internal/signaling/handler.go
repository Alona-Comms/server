package signaling

import (
	"log"
	"time"
)

func (s *Server) handleMessage(slug string, participant *Participant, message *Message) {
	s.mutex.RLock()
	room, exists := s.rooms[slug]
	s.mutex.RUnlock()

	if !exists {
		return
	}

	switch message.Type {
	case MessageTypeAllow:
		s.handleAllow(room, participant, message)
	case MessageTypeDeny:
		s.handleDeny(room, participant, message)
	case MessageTypeOffer, MessageTypeAnswer, MessageTypeICECandidate:
		s.handleWebRTCMessage(room, participant, message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

func (s *Server) handleAllow(room *Room, participant *Participant, message *Message) {
	// Only the host can allow
	if participant.Role != RoleHost {
		return
	}

	guestID, ok := message.Data.(string)
	if !ok {
		return
	}

	err := room.AllowGuest(guestID)
	if err != nil {
		log.Printf("Failed to allow guest: %v", err)
		return
	}

	// Notify the guest about the allowance
	allowMessage := &Message{
		Type:      MessageTypeAllow,
		From:      participant.ID,
		To:        guestID,
		Slug:      room.Slug,
		Timestamp: time.Now(),
	}
	room.BroadcastToGuest(guestID, allowMessage)

	// Notify all participants about the updated participant list
	participantsMessage := &Message{
		Type:      MessageTypeParticipants,
		Slug:      room.Slug,
		Data:      room.GetParticipantsData(),
		Timestamp: time.Now(),
	}
	room.BroadcastToAll(participantsMessage, "")
}

func (s *Server) handleDeny(room *Room, participant *Participant, message *Message) {
	// Only the host can deny
	if participant.Role != RoleHost {
		return
	}

	guestID, ok := message.Data.(string)
	if !ok {
		return
	}

	guest := room.GetParticipant(guestID)
	if guest == nil {
		return
	}

	// Notify the guest about the denial
	denyMessage := &Message{
		Type:      MessageTypeDeny,
		From:      participant.ID,
		To:        guestID,
		Slug:      room.Slug,
		Timestamp: time.Now(),
	}
	room.BroadcastToGuest(guestID, denyMessage)

	// Remove the guest from the room
	room.DenyGuest(guestID)
	guest.Conn.Close()
}

func (s *Server) handleWebRTCMessage(room *Room, participant *Participant, message *Message) {
	// Only participants with "in_room" status can exchange WebRTC signals
	if participant.Status != StatusInRoom {
		return
	}

	// If a recipient is specified, send only to them
	if message.To != "" {
		targetParticipant := room.GetParticipant(message.To)
		if targetParticipant != nil && targetParticipant.Status == StatusInRoom {
			if targetParticipant.Role == RoleHost {
				room.BroadcastToHost(message)
			} else {
				room.BroadcastToGuest(message.To, message)
			}
		}
		return
	}

	// Otherwise, broadcast to all other participants in the room
	room.BroadcastToAll(message, participant.ID)
}
