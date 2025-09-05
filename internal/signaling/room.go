package signaling

import (
	"fmt"
	"time"
)

func NewRoom(slug string) *Room {
	return &Room{
		Slug:       slug,
		Guests:     make(map[string]*Participant),
		PublicKeys: make(map[string]string),
		CreatedAt:  time.Now(),
	}
}

func (r *Room) SavePublicKey(participantID, publicKey string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if err := ValidatePublicKey(publicKey); err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	if r.PublicKeys == nil {
		r.PublicKeys = make(map[string]string)
	}
	r.PublicKeys[participantID] = publicKey

	if r.Host != nil && r.Host.ID == participantID {
		r.Host.Keys.PublicKey = publicKey
	}
	if guest, exists := r.Guests[participantID]; exists {
		guest.Keys.PublicKey = publicKey
	}

	return nil
}

func (r *Room) GetPublicKey(participantID string) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key, exists := r.PublicKeys[participantID]
	return key, exists
}

func (r *Room) GetAllPublicKeys() map[string]string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	keys := make(map[string]string)
	for id, key := range r.PublicKeys {
		keys[id] = key
	}
	return keys
}

func (r *Room) BroadcastPublicKeys(excludeID string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	message := &Message{
		Type: MessageTypePublicKeys,
		Data: PublicKeysData{
			Keys: r.PublicKeys,
		},
		Timestamp: time.Now(),
	}

	if r.Host != nil && r.Host.ID != excludeID && r.Host.Status == StatusInRoom {
		r.Host.Conn.WriteJSON(message)
	}

	for _, guest := range r.Guests {
		if guest.ID != excludeID && guest.Status == StatusInRoom {
			guest.Conn.WriteJSON(message)
		}
	}
}

func (r *Room) RemovePublicKey(participantID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.PublicKeys, participantID)
}

func (r *Room) AddParticipant(participant *Participant) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if participant.Role == RoleHost {
		if r.Host != nil {
			return fmt.Errorf("room already has a host")
		}
		r.Host = participant
		participant.Status = StatusInRoom
	} else {
		participant.Status = StatusKnocking
		r.Guests[participant.ID] = participant
	}

	return nil
}

func (r *Room) RemoveParticipant(participantID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.Host != nil && r.Host.ID == participantID {
		r.Host = nil
	} else {
		delete(r.Guests, participantID)
	}
}

func (r *Room) GetParticipant(participantID string) *Participant {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.Host != nil && r.Host.ID == participantID {
		return r.Host
	}

	return r.Guests[participantID]
}

func (r *Room) AllowGuest(guestID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	guest, exists := r.Guests[guestID]
	if !exists {
		return fmt.Errorf("guest not found")
	}

	guest.Status = StatusInRoom
	return nil
}

func (r *Room) DenyGuest(guestID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	guest, exists := r.Guests[guestID]
	if !exists {
		return fmt.Errorf("guest not found")
	}

	guest.Status = StatusDisconnected
	delete(r.Guests, guestID)
	return nil
}

func (r *Room) GetParticipantsData() *ParticipantsData {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := len(r.Guests)
	if r.Host != nil {
		count++
	}

	return &ParticipantsData{
		Host:   r.Host,
		Guests: r.Guests,
		Count:  count,
	}
}

func (r *Room) IsEmpty() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.Host == nil && len(r.Guests) == 0
}

func (r *Room) BroadcastToAll(message *Message, excludeID string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.Host != nil && r.Host.ID != excludeID {
		r.Host.Conn.WriteJSON(message)
	}

	for _, guest := range r.Guests {
		if guest.ID != excludeID && guest.Status == StatusInRoom {
			guest.Conn.WriteJSON(message)
		}
	}
}

func (r *Room) BroadcastToHost(message *Message) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.Host != nil {
		r.Host.Conn.WriteJSON(message)
	}
}

func (r *Room) BroadcastToGuest(guestID string, message *Message) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if guest, exists := r.Guests[guestID]; exists {
		guest.Conn.WriteJSON(message)
	}
}
