package order_calculations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	apperrors "github.com/pack-calculator/pkg/errors"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) OrderProcessing(ctx context.Context, orderQuantity int) ([]PackResult, int, int, error) {
	args := m.Called(ctx, orderQuantity)
	if args.Get(0) == nil {
		return nil, 0, 0, args.Error(3)
	}
	return args.Get(0).([]PackResult), args.Int(1), args.Int(2), args.Error(3)
}

func (m *MockService) CalculateOptimalPacks(ctx context.Context, orderQuantity int, packSizes []int) (map[int]int, error) {
	args := m.Called(ctx, orderQuantity, packSizes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[int]int), args.Error(1)
}

// ErrorResponse matches the JSON error response structure
type ErrorResponse struct {
	Type    string      `json:"Type"`
	Message string      `json:"Message"`
	Err     interface{} `json:"Err"`
}

func TestHandler_CalculatePacksForOrder(t *testing.T) {
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
				c.Set("payload", &CalculateAPIRequest{OrderQuantity: 10})
			},
			mockSetup: func(m *MockService) {
				m.On("OrderProcessing", mock.Anything, 10).Return([]PackResult{
					{Size: 5, Quantity: 2},
				}, 10, 2, nil)
			},
			wantStatusCode: http.StatusOK,
			wantBody: func() interface{} {
				return &CalculateAPIResponse{
					OrderQuantity: 10,
					TotalItems:    10,
					TotalPacks:    2,
					Packs: []PackResult{
						{Size: 5, Quantity: 2},
					},
					Success: true,
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
				c.Set("payload", &CalculateAPIRequest{OrderQuantity: 10})
			},
			mockSetup: func(m *MockService) {
				m.On("OrderProcessing", mock.Anything, 10).Return(nil, 0, 0, errors.New("service error"))
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBody: func() interface{} {
				return &ErrorResponse{
					Type:    string(apperrors.ErrorTypeInternal),
					Message: "Failed to process order request",
					Err:     map[string]interface{}{},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			mockService := new(MockService)
			tt.mockSetup(mockService)

			tt.setupContext(c)

			logger := zap.NewNop()
			handler := NewHandler(logger, mockService)
			handler.CalculatePacksForOrder(c)

			assert.Equal(t, tt.wantStatusCode, w.Code)

			var got interface{}
			switch tt.wantStatusCode {
			case http.StatusOK:
				got = &CalculateAPIResponse{}
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
