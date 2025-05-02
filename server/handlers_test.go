package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/piftai/load-balancer/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository реализует repository.Repository для тестов
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(client repository.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockRepository) Get(id string) (*repository.Client, error) {
	args := m.Called(id)
	return args.Get(0).(*repository.Client), args.Error(1)
}

func (m *MockRepository) Update(client repository.Client) error {
	args := m.Called(client)
	return args.Error(0)
}

func (m *MockRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestClientHandler_Create(t *testing.T) {
	tests := []struct {
		name        string
		requestBody interface{}
		mockSetup   func(*MockRepository)
		wantStatus  int
	}{
		{
			name: "successful creation",
			requestBody: map[string]string{
				"id":   "test-id",
				"name": "test-client",
			},
			mockSetup: func(mr *MockRepository) {
				mr.On("Create", mock.AnythingOfType("repository.Client")).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:        "invalid request body",
			requestBody: "invalid",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "repository error",
			requestBody: map[string]string{
				"id":   "test-id",
				"name": "test-client",
			},
			mockSetup: func(mr *MockRepository) {
				mr.On("Create", mock.AnythingOfType("repository.Client")).Return(assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewClientHandler(mockRepo)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Create(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		clientID   string
		mockSetup  func(*MockRepository)
		wantStatus int
	}{
		{
			name:     "client found",
			clientID: "existing-id",
			mockSetup: func(mr *MockRepository) {
				mr.On("Get", "existing-id").Return(&repository.Client{ID: "existing-id"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "client not found",
			clientID: "non-existent-id",
			mockSetup: func(mr *MockRepository) {
				mr.On("Get", "non-existent-id").Return((*repository.Client)(nil), nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "repository error",
			clientID: "error-id",
			mockSetup: func(mr *MockRepository) {
				mr.On("Get", "error-id").Return((*repository.Client)(nil), assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			handler := NewClientHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/clients/"+tt.clientID, nil)
			req.SetPathValue("id", tt.clientID)
			rec := httptest.NewRecorder()

			handler.Get(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			mockRepo.AssertExpectations(t)

			if rec.Code == http.StatusOK {
				var client repository.Client
				err := json.NewDecoder(rec.Body).Decode(&client)
				require.NoError(t, err)
				assert.Equal(t, tt.clientID, client.ID)
			}
		})
	}
}

func TestClientHandler_Update(t *testing.T) {
	tests := []struct {
		name        string
		requestBody interface{}
		mockSetup   func(*MockRepository)
		wantStatus  int
	}{
		{
			name: "successful update",
			requestBody: map[string]string{
				"id":   "test-id",
				"name": "updated-client",
			},
			mockSetup: func(mr *MockRepository) {
				mr.On("Update", mock.AnythingOfType("repository.Client")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "invalid request body",
			requestBody: "invalid",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "missing client ID",
			requestBody: map[string]string{
				"name": "no-id-client",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "repository error",
			requestBody: map[string]string{
				"id":   "test-id",
				"name": "updated-client",
			},
			mockSetup: func(mr *MockRepository) {
				mr.On("Update", mock.AnythingOfType("repository.Client")).Return(assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			handler := NewClientHandler(mockRepo)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/clients", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.Update(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestClientHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		clientID   string
		mockSetup  func(*MockRepository)
		wantStatus int
	}{
		{
			name:     "successful deletion",
			clientID: "existing-id",
			mockSetup: func(mr *MockRepository) {
				mr.On("Delete", "existing-id").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:     "repository error",
			clientID: "error-id",
			mockSetup: func(mr *MockRepository) {
				mr.On("Delete", "error-id").Return(assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mockSetup(mockRepo)

			handler := NewClientHandler(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, "/clients/"+tt.clientID, nil)
			req.SetPathValue("id", tt.clientID)
			rec := httptest.NewRecorder()

			handler.Delete(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
