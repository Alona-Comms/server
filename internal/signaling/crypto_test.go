package signaling

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGenerateEd25519KeyPair(t *testing.T) {
	pub, priv, err := GenerateEd25519KeyPair()

	assert.NoError(t, err)
	assert.NotEmpty(t, pub)
	assert.NotEmpty(t, priv)

	pubBytes, err := base64.StdEncoding.DecodeString(pub)
	assert.NoError(t, err)
	assert.Equal(t, ed25519.PublicKeySize, len(pubBytes))

	privBytes, err := base64.StdEncoding.DecodeString(priv)
	assert.NoError(t, err)
	assert.Equal(t, ed25519.PrivateKeySize, len(privBytes))
}

func TestValidatePublicKey(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		wantErr   bool
	}{
		{
			name:      "valid key",
			publicKey: "3p6s4iEO3Z0Qo7J8k4hzJ5WrGJ8r4W7hQ8zE2kV9P8Y=",
			wantErr:   false,
		},
		{
			name:      "invalid base64",
			publicKey: "invalid-base64!",
			wantErr:   true,
		},
		{
			name:      "wrong size",
			publicKey: base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr:   true,
		},
		{
			name:      "empty key",
			publicKey: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "valid key" {
				pub, _, err := GenerateEd25519KeyPair()
				assert.NoError(t, err)
				tt.publicKey = pub
			}

			err := ValidatePublicKey(tt.publicKey)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoomPublicKeys(t *testing.T) {
	room := NewRoom("test-room")

	publicKey, _, err := GenerateEd25519KeyPair()
	assert.NoError(t, err)

	participantID := "user1"

	err = room.SavePublicKey(participantID, publicKey)
	assert.NoError(t, err)

	savedKey, exists := room.GetPublicKey(participantID)
	assert.True(t, exists)
	assert.Equal(t, publicKey, savedKey)

	allKeys := room.GetAllPublicKeys()
	assert.Equal(t, 1, len(allKeys))
	assert.Equal(t, publicKey, allKeys[participantID])

	room.RemovePublicKey(participantID)
	_, exists = room.GetPublicKey(participantID)
	assert.False(t, exists)
}

func TestKeyExchangeIntegration(t *testing.T) {
	server := NewServer()
	room := NewRoom("test-room")
	server.rooms["test-room"] = room

	mockConn1 := &MockWebSocketConn{} // user1 
	mockConn2 := &MockWebSocketConn{} // user2

	participant1 := &Participant{
		ID:     "user1",
		Conn:   mockConn1,
		Role:   RoleGuest,
		Status: StatusInRoom,
	}

	participant2 := &Participant{
		ID:     "user2", 
		Conn:   mockConn2,
		Role:   RoleHost,
		Status: StatusInRoom,
	}

	room.AddParticipant(participant1)
	room.AddParticipant(participant2)

	mockConn2.On("WriteJSON", mock.AnythingOfType("*signaling.Message")).Return(nil).Once()

	publicKey, _, err := GenerateEd25519KeyPair()
	assert.NoError(t, err)

	message := &Message{
		Type: MessageTypeKeyExchange,
		Data: map[string]interface{}{
			"public_key": publicKey,
		},
	}

	server.handleMessage("test-room", participant1, message)

	savedKey, exists := room.GetPublicKey("user1")
	assert.True(t, exists)
	assert.Equal(t, publicKey, savedKey)

	mockConn2.AssertExpectations(t)
}
