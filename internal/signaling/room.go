package signaling

import (
	"fmt"
	"time"
)

// NewRoom создает новую комнату
func NewRoom(slug string) *Room {
	return &Room{
		Slug:      slug,
		Guests:    make(map[string]*Participant),
		CreatedAt: time.Now(),
	}
}

// AddParticipant добавляет участника в комнату
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
		// Гости начинают со статуса "knocking"
		participant.Status = StatusKnocking
		r.Guests[participant.ID] = participant
	}

	return nil
}

// RemoveParticipant удаляет участника из комнаты
func (r *Room) RemoveParticipant(participantID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.Host != nil && r.Host.ID == participantID {
		r.Host = nil
	} else {
		delete(r.Guests, participantID)
	}
}

// GetParticipant возвращает участника по ID
func (r *Room) GetParticipant(participantID string) *Participant {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.Host != nil && r.Host.ID == participantID {
		return r.Host
	}

	return r.Guests[participantID]
}

// AllowGuest разрешает гостю войти в комнату
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

// DenyGuest отклоняет запрос гостя на вход
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

// GetParticipantsData возвращает данные об участниках
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

// IsEmpty проверяет, пуста ли комната
func (r *Room) IsEmpty() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.Host == nil && len(r.Guests) == 0
}

// BroadcastToAll рассылает сообщение всем участникам
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

// BroadcastToHost отправляет сообщение только хосту
func (r *Room) BroadcastToHost(message *Message) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if r.Host != nil {
		r.Host.Conn.WriteJSON(message)
	}
}

// BroadcastToGuest отправляет сообщение конкретному гостю
func (r *Room) BroadcastToGuest(guestID string, message *Message) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if guest, exists := r.Guests[guestID]; exists {
		guest.Conn.WriteJSON(message)
	}
}
