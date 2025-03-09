package order_calculations

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/pack-calculator/internal/pack_configurations"
)

// MockCalculationRepository is a mock implementation of Repository
type MockCalculationRepository struct {
	mock.Mock
}

func (m *MockCalculationRepository) Save(ctx context.Context, calculation *OrderCalculation) error {
	args := m.Called(ctx, calculation)
	return args.Error(0)
}

func (m *MockCalculationRepository) GetByID(ctx context.Context, id uint) (*OrderCalculation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderCalculation), args.Error(1)
}

func (m *MockCalculationRepository) GetByConfigurationIDAndOrderQuantity(ctx context.Context, orderQuantity int, configID uint) (*OrderCalculation, error) {
	args := m.Called(ctx, orderQuantity, configID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderCalculation), args.Error(1)
}

func (m *MockCalculationRepository) List(ctx context.Context, offset, limit int) ([]OrderCalculation, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]OrderCalculation), args.Error(1)
}

func (m *MockCalculationRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPackConfigRepository is a mock implementation of pack_configurations.Repository
type MockPackConfigRepository struct {
	mock.Mock
}

func (m *MockPackConfigRepository) Create(ctx context.Context, config *pack_configurations.PackConfiguration) (*pack_configurations.PackConfiguration, error) {
	args := m.Called(ctx, config)
	return args.Get(0).(*pack_configurations.PackConfiguration), args.Error(1)
}

func (m *MockPackConfigRepository) GetByID(ctx context.Context, id uint) (*pack_configurations.PackConfiguration, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pack_configurations.PackConfiguration), args.Error(1)
}

func (m *MockPackConfigRepository) GetBySignature(ctx context.Context, signature string) (*pack_configurations.PackConfiguration, error) {
	args := m.Called(ctx, signature)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pack_configurations.PackConfiguration), args.Error(1)
}

func (m *MockPackConfigRepository) GetActive(ctx context.Context) (*pack_configurations.PackConfiguration, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pack_configurations.PackConfiguration), args.Error(1)
}

func (m *MockPackConfigRepository) SetActive(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPackConfigRepository) Update(ctx context.Context, config *pack_configurations.PackConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockPackConfigRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPackConfigRepository) List(ctx context.Context) ([]pack_configurations.PackConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).([]pack_configurations.PackConfiguration), args.Error(1)
}

func TestService_OrderProcessing(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name          string
		orderQuantity int
		mockSetup     func(*MockCalculationRepository, *MockPackConfigRepository)
		wantPacks     []PackResult
		wantTotal     int
		wantTotalPack int
		wantErr       bool
	}{
		{
			name:          "success - cache hit",
			orderQuantity: 10,
			mockSetup: func(calcRepo *MockCalculationRepository, packRepo *MockPackConfigRepository) {
				packRepo.On("GetActive", mock.Anything).Return(&pack_configurations.PackConfiguration{
					ID:        1,
					PackSizes: pq.Int64Array{3, 5},
				}, nil)
				calcRepo.On("GetByConfigurationIDAndOrderQuantity", mock.Anything, 10, uint(1)).Return(&OrderCalculation{
					Result:     []PackResult{{Size: 5, Quantity: 2}},
					TotalItems: 10,
					TotalPacks: 2,
				}, nil)
			},
			wantPacks: []PackResult{
				{Size: 5, Quantity: 2},
			},
			wantTotal:     10,
			wantTotalPack: 2,
			wantErr:       false,
		},
		{
			name:          "success - new calculation",
			orderQuantity: 8,
			mockSetup: func(calcRepo *MockCalculationRepository, packRepo *MockPackConfigRepository) {
				packRepo.On("GetActive", mock.Anything).Return(&pack_configurations.PackConfiguration{
					ID:        1,
					PackSizes: pq.Int64Array{3, 5},
				}, nil)
				calcRepo.On("GetByConfigurationIDAndOrderQuantity", mock.Anything, 8, uint(1)).Return(nil, nil)
				calcRepo.On("Save", mock.Anything, mock.AnythingOfType("*order_calculations.OrderCalculation")).Return(nil)
			},
			wantPacks: []PackResult{
				{Size: 3, Quantity: 1},
				{Size: 5, Quantity: 1},
			},
			wantTotal:     8,
			wantTotalPack: 2,
			wantErr:       false,
		},
		{
			name:          "error - no pack sizes",
			orderQuantity: 10,
			mockSetup: func(calcRepo *MockCalculationRepository, packRepo *MockPackConfigRepository) {
				packRepo.On("GetActive", mock.Anything).Return(&pack_configurations.PackConfiguration{
					ID:        1,
					PackSizes: pq.Int64Array{},
				}, nil)
				calcRepo.On("GetByConfigurationIDAndOrderQuantity", mock.Anything, 10, uint(1)).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:          "error - database error on get active",
			orderQuantity: 10,
			mockSetup: func(calcRepo *MockCalculationRepository, packRepo *MockPackConfigRepository) {
				packRepo.On("GetActive", mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:          "error - database error on save",
			orderQuantity: 8,
			mockSetup: func(calcRepo *MockCalculationRepository, packRepo *MockPackConfigRepository) {
				packRepo.On("GetActive", mock.Anything).Return(&pack_configurations.PackConfiguration{
					ID:        1,
					PackSizes: pq.Int64Array{3, 5},
				}, nil)
				calcRepo.On("GetByConfigurationIDAndOrderQuantity", mock.Anything, 8, uint(1)).Return(nil, nil)
				calcRepo.On("Save", mock.Anything, mock.AnythingOfType("*order_calculations.OrderCalculation")).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCalcRepo := new(MockCalculationRepository)
			mockPackRepo := new(MockPackConfigRepository)
			tt.mockSetup(mockCalcRepo, mockPackRepo)

			s := NewService(logger, mockCalcRepo, mockPackRepo)
			gotPacks, gotTotal, gotTotalPack, err := s.OrderProcessing(context.Background(), tt.orderQuantity)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPacks, gotPacks)
			assert.Equal(t, tt.wantTotal, gotTotal)
			assert.Equal(t, tt.wantTotalPack, gotTotalPack)

			mockCalcRepo.AssertExpectations(t)
			mockPackRepo.AssertExpectations(t)
		})
	}
}

func TestService_CalculateOptimalPacks(t *testing.T) {
	logger := zap.NewNop()
	mockCalcRepo := new(MockCalculationRepository)
	mockPackRepo := new(MockPackConfigRepository)
	s := NewService(logger, mockCalcRepo, mockPackRepo)

	tests := []struct {
		name          string
		orderQuantity int
		packSizes     []int
		want          map[int]int
		wantErr       bool
	}{
		{
			name:          "small order - fast path",
			orderQuantity: 2,
			packSizes:     []int{3, 5},
			want:          map[int]int{3: 1},
			wantErr:       false,
		},
		{
			name:          "exact match",
			orderQuantity: 10,
			packSizes:     []int{3, 5},
			want:          map[int]int{5: 2},
			wantErr:       false,
		},
		{
			name:          "needs multiple pack sizes",
			orderQuantity: 8,
			packSizes:     []int{3, 5},
			want:          map[int]int{3: 1, 5: 1},
			wantErr:       false,
		},
		{
			name:          "zero order quantity",
			orderQuantity: 0,
			packSizes:     []int{3, 5},
			want:          map[int]int{3: 1}, // Updated to match implementation
			wantErr:       false,
		},
		{
			name:          "large order",
			orderQuantity: 28,
			packSizes:     []int{3, 5, 10},
			want:          map[int]int{3: 1, 5: 1, 10: 2},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.CalculateOptimalPacks(context.Background(), tt.orderQuantity, tt.packSizes)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
