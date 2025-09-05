package signaling

import "github.com/stretchr/testify/mock"

type MockWebSocketConn struct {
	mock.Mock
}

func (m *MockWebSocketConn) WriteJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockWebSocketConn) ReadJSON(v interface{}) error {
	args := m.Called(v)
	return args.Error(0)
}

func (m *MockWebSocketConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWebSocketConn) ReadMessage() (messageType int, p []byte, err error) {
	args := m.Called()
	return args.Int(0), args.Get(1).([]byte), args.Error(2)
}

func (m *MockWebSocketConn) WriteMessage(messageType int, data []byte) error {
	args := m.Called(messageType, data)
	return args.Error(0)
}
