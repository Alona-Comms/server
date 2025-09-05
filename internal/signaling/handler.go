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
	case MessageTypeKeyExchange:
		s.handleKeyExchange(room, participant, message)
	case MessageTypeEncrypted:
		s.handleEncryptedData(room, participant, message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
	}
}

func (s *Server) handleKeyExchange(room *Room, participant *Participant, message *Message) {
	data, ok := message.Data.(map[string]interface{})
	if !ok {
		log.Printf("Invalid key exchange data format")
		return
	}

	publicKey, ok := data["public_key"].(string)
	if !ok {
		log.Printf("Missing public key in key exchange")
		return
	}

	if err := room.SavePublicKey(participant.ID, publicKey); err != nil {
		log.Printf("Failed to save public key for %s: %v", participant.ID, err)

		errorMsg := &Message{
			Type: MessageTypeError,
			Data: ErrorData{
				Code:    "INVALID_PUBLIC_KEY",
				Message: "Invalid public key format",
			},
			Timestamp: time.Now(),
		}
		participant.Conn.WriteJSON(errorMsg)
		return
	}

	log.Printf("Saved public key for participant %s in room %s", participant.ID, room.Slug)

	// ВАЖНО: Убедитесь что BroadcastPublicKeys вызывается
	log.Printf("Broadcasting public keys to all participants")
	room.BroadcastPublicKeys("")
	log.Printf("Broadcast completed")
}

func (s *Server) handleEncryptedData(room *Room, participant *Participant, message *Message) {
	if participant.Status != StatusInRoom {
		return
	}

	data, ok := message.Data.(map[string]interface{})
	if !ok {
		return
	}

	toParticipantID, ok := data["to"].(string)
	if !ok {
		return
	}

	message.From = participant.ID
	message.Timestamp = time.Now()

	if toParticipantID == "all" {
		room.BroadcastToAll(message, participant.ID)
	} else {
		if room.Host != nil && room.Host.ID == toParticipantID {
			room.BroadcastToHost(message)
		} else {
			room.BroadcastToGuest(toParticipantID, message)
		}
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
