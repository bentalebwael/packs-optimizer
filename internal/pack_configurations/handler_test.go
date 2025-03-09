package pack_configurations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	apperrors "github.com/pack-calculator/pkg/errors"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, config *PackConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockService) GetActive(ctx context.Context) (*PackConfiguration, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PackConfiguration), args.Error(1)
}

// ErrorResponse matches the JSON error response structure
type ErrorResponse struct {
	Type    string      `json:"Type"`
	Message string      `json:"Message"`
	Err     interface{} `json:"Err"`
}

func TestHandler_GetActivePackConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockSetup      func(*MockService)
		wantStatusCode int
		wantBody       func() interface{}
	}{
		{
			name: "success case",
			mockSetup: func(m *MockService) {
				m.On("GetActive", mock.Anything).Return(&PackConfiguration{
					PackSizes: pq.Int64Array{250, 500, 1000},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: func() interface{} {
				return &PackCfgAPIResponse{
					PackSizes: []int{250, 500, 1000},
				}
			},
		},
		{
			name: "service error",
			mockSetup: func(m *MockService) {
				m.On("GetActive", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: func() interface{} {
				return &ErrorResponse{
					Type:    string(apperrors.ErrorTypeInternal),
					Message: "Failed to retrieve pack configuration",
					Err:     map[string]interface{}{},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request with context
			req := httptest.NewRequest(http.MethodGet, "/pack-configurations/active", nil)
			c.Request = req

			mockService := new(MockService)
			tt.mockSetup(mockService)

			logger := zap.NewNop()
			handler := NewHandler(logger, mockService)
			handler.GetActivePackConfiguration(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			var got interface{}
			switch tt.wantStatusCode {
			case http.StatusOK:
				got = &PackCfgAPIResponse{}
			default:
				got = &ErrorResponse{}
			}

			err := json.Unmarshal(w.Body.Bytes(), got)
			assert.NoError(t, err)

			want := tt.wantBody()
			assert.Equal(t, want, got)

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_CreatePackConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		mockSetup      func(*MockService)
		wantStatusCode int
		wantBody       func() interface{}
	}{
		{
			name: "success case",
			setupContext: func(c *gin.Context) {
				c.Set("payload", &PackCfgAPIRequest{
					PackSizes: []int{250, 500, 1000},
				})
			},
			mockSetup: func(m *MockService) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(cfg *PackConfiguration) bool {
					sizes := []int64(cfg.PackSizes)
					return len(sizes) == 3 && sizes[0] == 250 && sizes[1] == 500 && sizes[2] == 1000
				})).Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: func() interface{} {
				return &PackCfgAPIResponse{
					PackSizes: []int{250, 500, 1000},
				}
			},
		},
		{
			name: "missing payload",
			setupContext: func(c *gin.Context) {
				// Don't set payload
			},
			mockSetup: func(m *MockService) {
				// No mock needed
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: func() interface{} {
				return &ErrorResponse{
					Type:    string(apperrors.ErrorTypeInternal),
					Message: "Failed to retrieve payload from context",
					Err:     map[string]interface{}{},
				}
			},
		},
		{
			name: "service error",
			setupContext: func(c *gin.Context) {
				c.Set("payload", &PackCfgAPIRequest{
					PackSizes: []int{250, 500, 1000},
				})
			},
			mockSetup: func(m *MockService) {
				m.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: func() interface{} {
				return &ErrorResponse{
					Type:    string(apperrors.ErrorTypeInternal),
					Message: "Failed to create pack configuration",
					Err:     map[string]interface{}{},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request with context
			req := httptest.NewRequest(http.MethodPost, "/pack-configurations", nil)
			c.Request = req

			mockService := new(MockService)
			if tt.setupContext != nil {
				tt.setupContext(c)
			}
			tt.mockSetup(mockService)

			logger := zap.NewNop()
			handler := NewHandler(logger, mockService)
			handler.CreatePackConfiguration(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			var got interface{}
			switch tt.wantStatusCode {
			case http.StatusOK:
				got = &PackCfgAPIResponse{}
			default:
				got = &ErrorResponse{}
			}

			err := json.Unmarshal(w.Body.Bytes(), got)
			assert.NoError(t, err)

			want := tt.wantBody()
			assert.Equal(t, want, got)

			mockService.AssertExpectations(t)
		})
	}
}
