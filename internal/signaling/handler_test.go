package signaling

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleWebRTCMessageWithAllowedGuest(t *testing.T) {
	server := NewServer()
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
		Status: StatusKnocking,
	}

	room.AddParticipant(host)
	room.AddParticipant(guest)

	// Step 1: Host allows guest
	allowMessage := &Message{
		Type: MessageTypeAllow,
		Data: "guest1",
	}

	mockGuestConn.On("WriteJSON", mock.Anything).Return(nil).Times(2) // Allow + Participants
	mockHostConn.On("WriteJSON", mock.Anything).Return(nil).Times(1)  // Participants

	server.handleAllow(room, host, allowMessage)

	// Verify guest is now in room
	assert.Equal(t, StatusInRoom, guest.Status)

	// Step 2: Guest sends WebRTC message
	webRTCMessage := &Message{
		Type: MessageTypeOffer,
		To:   "host1",
		Data: map[string]interface{}{"sdp": "test-offer"},
	}

	mockHostConn.On("WriteJSON", webRTCMessage).Return(nil).Once()

	server.handleWebRTCMessage(room, guest, webRTCMessage)

	mockHostConn.AssertExpectations(t)
	mockGuestConn.AssertExpectations(t)
}

func TestHandleWebRTCMessageNotInRoom(t *testing.T) {
	server := NewServer()
	room := NewRoom("test-room")

	mockGuestConn := &MockWebSocketConn{}
	guest := &Participant{
		ID:     "guest1",
		Conn:   mockGuestConn,
		Role:   RoleGuest,
		Status: StatusKnocking, // Guest is NOT in the room
	}

	room.AddParticipant(guest)

	message := &Message{
		Type: MessageTypeOffer,
		Data: map[string]interface{}{"sdp": "test-offer"},
	}

	server.handleWebRTCMessage(room, guest, message)

	// No messages should have been sent
	mockGuestConn.AssertNotCalled(t, "WriteJSON")
}

func TestHandleAllow(t *testing.T) {
	server := NewServer()
	room := NewRoom("test-room")
	server.rooms["test-room"] = room

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
		Status: StatusKnocking,
	}

	room.AddParticipant(host)
	room.AddParticipant(guest)

	message := &Message{
		Type: MessageTypeAllow,
		Data: "guest1",
	}

	mockGuestConn.On("WriteJSON", mock.Anything).Return(nil).Times(2) // Allow + Participants
	mockHostConn.On("WriteJSON", mock.Anything).Return(nil).Once()    // Participants

	server.handleAllow(room, host, message)

	assert.Equal(t, StatusInRoom, guest.Status)
	mockGuestConn.AssertExpectations(t)
	mockHostConn.AssertExpectations(t)
}
func TestHandleDeny(t *testing.T) {
	server := NewServer()
	room := NewRoom("test-room")

	mockHostConn := &MockWebSocketConn{}
	mockGuestConn := &MockWebSocketConn{}

	host := &Participant{
		ID:   "host1",
		Conn: mockHostConn,
		Role: RoleHost,
	}

	guest := &Participant{
		ID:     "guest1",
		Conn:   mockGuestConn,
		Role:   RoleGuest,
		Status: StatusKnocking,
	}

	room.AddParticipant(host)
	room.AddParticipant(guest)

	message := &Message{
		Type: MessageTypeDeny,
		Data: "guest1",
	}

	// Ожидаем отправку сообщения гостю и закрытие соединения
	mockGuestConn.On("WriteJSON", mock.AnythingOfType("*signaling.Message")).Return(nil).Once()
	mockGuestConn.On("Close").Return(nil).Once()

	server.handleDeny(room, host, message)

	assert.Equal(t, 0, len(room.Guests))
	mockGuestConn.AssertExpectations(t)
}

func TestHandleWebRTCMessageRejectsKnockingGuest(t *testing.T) {
	server := NewServer()
	room := NewRoom("test-room")

	mockGuestConn := &MockWebSocketConn{}
	guest := &Participant{
		ID:     "guest1",
		Conn:   mockGuestConn,
		Role:   RoleGuest,
		Status: StatusKnocking, // Not allowed to send messages
	}

	room.AddParticipant(guest)

	message := &Message{
		Type: MessageTypeOffer,
		Data: map[string]interface{}{"sdp": "test-offer"},
	}

	server.handleWebRTCMessage(room, guest, message)

	// No messages should have been sent
	mockGuestConn.AssertNotCalled(t, "WriteJSON")
}
