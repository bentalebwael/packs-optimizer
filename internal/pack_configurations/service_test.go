package pack_configurations

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRepository is a mock implementation of Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, config *PackConfiguration) (*PackConfiguration, error) {
	args := m.Called(ctx, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PackConfiguration), args.Error(1)
}

func (m *MockRepository) GetByID(ctx context.Context, id uint) (*PackConfiguration, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PackConfiguration), args.Error(1)
}

func (m *MockRepository) GetBySignature(ctx context.Context, signature string) (*PackConfiguration, error) {
	args := m.Called(ctx, signature)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PackConfiguration), args.Error(1)
}

func (m *MockRepository) SetActive(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetActive(ctx context.Context) (*PackConfiguration, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PackConfiguration), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, config *PackConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context) ([]PackConfiguration, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PackConfiguration), args.Error(1)
}

func TestService_Create(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *PackConfiguration
		mock    func(*MockRepository)
		wantErr bool
	}{
		{
			name: "success - new configuration",
			config: &PackConfiguration{
				PackSizes: pq.Int64Array{250, 500, 1000, 2000, 5000},
			},
			mock: func(repo *MockRepository) {
				repo.On("GetBySignature", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*pack_configurations.PackConfiguration")).
					Return(&PackConfiguration{ID: 1}, nil)
				repo.On("SetActive", mock.Anything, uint(1)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "success - existing configuration",
			config: &PackConfiguration{
				PackSizes: pq.Int64Array{250, 500, 1000, 2000, 5000},
			},
			mock: func(repo *MockRepository) {
				repo.On("GetBySignature", mock.Anything, mock.AnythingOfType("string")).
					Return(&PackConfiguration{ID: 1}, nil)
				repo.On("SetActive", mock.Anything, uint(1)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - repository create error",
			config: &PackConfiguration{
				PackSizes: pq.Int64Array{250, 500, 1000, 2000, 5000},
			},
			mock: func(repo *MockRepository) {
				repo.On("GetBySignature", mock.Anything, mock.AnythingOfType("string")).
					Return(nil, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*pack_configurations.PackConfiguration")).
					Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "error - set active error",
			config: &PackConfiguration{
				PackSizes: pq.Int64Array{250, 500, 1000, 2000, 5000},
			},
			mock: func(repo *MockRepository) {
				repo.On("GetBySignature", mock.Anything, mock.AnythingOfType("string")).
					Return(&PackConfiguration{ID: 1}, nil)
				repo.On("SetActive", mock.Anything, uint(1)).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mock(mockRepo)

			s := NewService(logger, mockRepo)
			err := s.Create(context.Background(), tt.config)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetActive(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		mock    func(*MockRepository)
		want    *PackConfiguration
		wantErr bool
	}{
		{
			name: "success",
			mock: func(repo *MockRepository) {
				repo.On("GetActive", mock.Anything).
					Return(&PackConfiguration{
						ID:        1,
						PackSizes: pq.Int64Array{250, 500, 1000},
						Active:    true,
					}, nil)
			},
			want: &PackConfiguration{
				ID:        1,
				PackSizes: pq.Int64Array{250, 500, 1000},
				Active:    true,
			},
			wantErr: false,
		},
		{
			name: "error",
			mock: func(repo *MockRepository) {
				repo.On("GetActive", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.mock(mockRepo)

			s := NewService(logger, mockRepo)
			got, err := s.GetActive(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
