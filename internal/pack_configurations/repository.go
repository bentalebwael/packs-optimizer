package pack_configurations

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository defines the interface for pack configuration persistence operations
type Repository interface {
	Create(ctx context.Context, config *PackConfiguration) (*PackConfiguration, error)
	GetByID(ctx context.Context, id uint) (*PackConfiguration, error)
	GetBySignature(ctx context.Context, signature string) (*PackConfiguration, error)
	GetActive(ctx context.Context) (*PackConfiguration, error)
	SetActive(ctx context.Context, id uint) error
	Update(ctx context.Context, config *PackConfiguration) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]PackConfiguration, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM-based repository
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, config *PackConfiguration) (*PackConfiguration, error) {
	err := r.db.WithContext(ctx).Create(config).Error
	return config, err
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (*PackConfiguration, error) {
	var config PackConfiguration
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *gormRepository) GetBySignature(ctx context.Context, signature string) (*PackConfiguration, error) {
	var config PackConfiguration
	err := r.db.WithContext(ctx).Where("signature = ?", signature).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *gormRepository) GetActive(ctx context.Context) (*PackConfiguration, error) {
	var config PackConfiguration
	err := r.db.WithContext(ctx).Where("active = ?", true).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

func (r *gormRepository) SetActive(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate current active configuration if exists
		if err := tx.Model(&PackConfiguration{}).Where("active = ?", true).Update("active", false).Error; err != nil {
			return err
		}

		if err := tx.Model(&PackConfiguration{}).Where("id = ?", id).Update("active", true).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *gormRepository) Update(ctx context.Context, config *PackConfiguration) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&PackConfiguration{}, id).Error
}

func (r *gormRepository) List(ctx context.Context) ([]PackConfiguration, error) {
	var configs []PackConfiguration
	err := r.db.WithContext(ctx).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}
