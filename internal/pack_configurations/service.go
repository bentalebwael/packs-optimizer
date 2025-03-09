package pack_configurations

import (
	"context"

	"go.uber.org/zap"

	"github.com/pack-calculator/pkg/postgres"
	"github.com/pack-calculator/pkg/utils"
)

type Service interface {
	Create(ctx context.Context, config *PackConfiguration) error
	GetActive(ctx context.Context) (*PackConfiguration, error)
}

type service struct {
	logger *zap.Logger
	repo   Repository
}

func NewService(logger *zap.Logger, repo Repository) Service {
	return &service{
		logger: logger,
		repo:   repo,
	}
}

func (s *service) Create(ctx context.Context, config *PackConfiguration) error {
	// Calculate hash signature
	packSizes := postgres.Int64ArrayToIntSlice(config.PackSizes)
	config.Signature = utils.CalculateArrayHash(packSizes)

	packConfiguration, err := s.repo.GetBySignature(ctx, config.Signature)
	if err != nil {
		return err
	}
	if packConfiguration == nil {
		packConfiguration, err = s.repo.Create(ctx, config)
		if err != nil {
			return err
		}
	}
	return s.repo.SetActive(ctx, packConfiguration.ID)
}

func (s *service) GetActive(ctx context.Context) (*PackConfiguration, error) {
	return s.repo.GetActive(ctx)
}
