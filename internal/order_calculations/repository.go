package order_calculations

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Repository defines the interface for order calculation persistence operations
type Repository interface {
	Save(ctx context.Context, calc *OrderCalculation) error
	GetByID(ctx context.Context, id uint) (*OrderCalculation, error)
	GetByConfigurationIDAndOrderQuantity(ctx context.Context, OrderQuantity int, configID uint) (*OrderCalculation, error)
	List(ctx context.Context, offset, limit int) ([]OrderCalculation, error)
	Delete(ctx context.Context, id uint) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository creates a new GORM-based repository
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Save(ctx context.Context, calc *OrderCalculation) error {
	return r.db.WithContext(ctx).Create(calc).Error
}

func (r *gormRepository) GetByID(ctx context.Context, id uint) (*OrderCalculation, error) {
	var calc OrderCalculation
	err := r.db.WithContext(ctx).Preload("Configuration").First(&calc, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &calc, nil
}

func (r *gormRepository) GetByConfigurationIDAndOrderQuantity(ctx context.Context, orderQuantity int, configID uint) (*OrderCalculation, error) {
	var calc OrderCalculation
	err := r.db.WithContext(ctx).
		Where("order_quantity = ? AND configuration_id = ?", orderQuantity, configID).
		Preload("Configuration").
		First(&calc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &calc, nil
}

func (r *gormRepository) List(ctx context.Context, offset, limit int) ([]OrderCalculation, error) {
	var calcs []OrderCalculation
	err := r.db.WithContext(ctx).
		Preload("Configuration").
		Offset(offset).
		Limit(limit).
		Order("timestamp DESC").
		Find(&calcs).Error
	if err != nil {
		return nil, err
	}
	return calcs, nil
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&OrderCalculation{}, id).Error
}
