package signaling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBroadcastToAll(t *testing.T) {
	room := NewRoom("test-room")
	mockHostConn := &MockWebSocketConn{}
	mockGuestConn := &MockWebSocketConn{}

	host := &Participant{
		ID:     "host1",
		Conn:   mockHostConn,
		Role:   RoleHost,
		Status: StatusInRoom, // Важно: статус должен быть StatusInRoom
	}

	guest := &Participant{
		ID:     "guest1",
		Conn:   mockGuestConn,
		Role:   RoleGuest,
		Status: StatusInRoom, // Важно: статус должен быть StatusInRoom для broadcast
	}

	room.AddParticipant(host)
	room.AddParticipant(guest)
	// Изменяем статус гостя на StatusInRoom вручную
	guest.Status = StatusInRoom

	message := &Message{
		Type: MessageTypeOffer,
		From: "sender1",
	}

	// Ожидаем вызовы у обоих участников
	mockHostConn.On("WriteJSON", mock.Anything).Return(nil).Once()
	mockGuestConn.On("WriteJSON", mock.Anything).Return(nil).Once()

	room.BroadcastToAll(message, "exclude1")

	mockHostConn.AssertExpectations(t)
	mockGuestConn.AssertExpectations(t)
}

func TestBroadcastToAllExcludeSender(t *testing.T) {
	room := NewRoom("test-room")
	mockHostConn := &MockWebSocketConn{}
	mockGuestConn := &MockWebSocketConn{}

	host := &Participant{
		ID:     "host1",
		Conn:   mockHostConn,
		Role:   RoleHost,
		Status: StatusInRoom,
	}

	guest := &Participant{
		ID:     "guest1",
		Conn:   mockGuestConn,
		Role:   RoleGuest,
		Status: StatusInRoom,
	}

	room.AddParticipant(host)
	room.AddParticipant(guest)
	// Изменяем статус гостя на StatusInRoom вручную
	guest.Status = StatusInRoom

	message := &Message{
		Type: MessageTypeOffer,
		From: "host1",
	}

	// Только гость должен получить сообщение
	mockGuestConn.On("WriteJSON", mock.Anything).Return(nil).Once()

	room.BroadcastToAll(message, "host1")

	mockGuestConn.AssertExpectations(t)
	mockHostConn.AssertNotCalled(t, "WriteJSON")
}

func TestBroadcastToHost(t *testing.T) {
	room := NewRoom("test-room")
	mockHostConn := &MockWebSocketConn{}

	host := &Participant{
		ID:   "host1",
		Conn: mockHostConn,
		Role: RoleHost,
	}

	room.AddParticipant(host)

	message := &Message{
		Type: MessageTypeKnock,
		From: "guest1",
	}

	mockHostConn.On("WriteJSON", mock.Anything).Return(nil).Once()

	room.BroadcastToHost(message)
	mockHostConn.AssertExpectations(t)
}

func TestBroadcastToGuest(t *testing.T) {
	room := NewRoom("test-room")
	mockGuestConn := &MockWebSocketConn{}

	guest := &Participant{
		ID:   "guest1",
		Conn: mockGuestConn,
		Role: RoleGuest,
	}

	room.AddParticipant(guest)

	message := &Message{
		Type: MessageTypeAllow,
		To:   "guest1",
	}

	mockGuestConn.On("WriteJSON", mock.Anything).Return(nil).Once()

	room.BroadcastToGuest("guest1", message)
	mockGuestConn.AssertExpectations(t)
}

// ... остальные тесты остаются без изменений

func TestAllowGuest(t *testing.T) {
	room := NewRoom("test-room")
	mockConn := &MockWebSocketConn{}

	guest := &Participant{
		ID:     "guest1",
		Conn:   mockConn,
		Role:   RoleGuest,
		Status: StatusKnocking,
	}

	room.AddParticipant(guest)

	err := room.AllowGuest("guest1")
	assert.NoError(t, err)
	assert.Equal(t, StatusInRoom, guest.Status)
}

func TestIsEmpty(t *testing.T) {
	room := NewRoom("test-room")
	assert.True(t, room.IsEmpty())

	mockConn := &MockWebSocketConn{}
	host := &Participant{
		ID:   "host1",
		Conn: mockConn,
		Role: RoleHost,
	}

	room.AddParticipant(host)
	assert.False(t, room.IsEmpty())

	room.RemoveParticipant("host1")
	assert.True(t, room.IsEmpty())
}

func TestGetParticipantsData(t *testing.T) {
	room := NewRoom("test-room")
	mockConn := &MockWebSocketConn{}

	host := &Participant{
		ID:   "host1",
		Conn: mockConn,
		Role: RoleHost,
	}

	guest := &Participant{
		ID:   "guest1",
		Conn: mockConn,
		Role: RoleGuest,
	}

	room.AddParticipant(host)
	room.AddParticipant(guest)

	data := room.GetParticipantsData()
	assert.Equal(t, host, data.Host)
	assert.Equal(t, 1, len(data.Guests))
	assert.Equal(t, guest, data.Guests["guest1"])
	assert.Equal(t, 2, data.Count)
}
