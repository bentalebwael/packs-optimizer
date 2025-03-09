package order_calculations

import (
	"time"

	packcfg "github.com/pack-calculator/internal/pack_configurations"
)

// OrderCalculation represents a calculation result entity in the database
type OrderCalculation struct {
	ID              uint                      `gorm:"column:id;primarykey;autoIncrement;column:id" json:"id"`
	OrderQuantity   int                       `gorm:"column:order_quantity;not null" json:"orderQuantity"`
	Result          []PackResult              `gorm:"column:result;serializer:json;not null" json:"result"`
	TotalItems      int                       `gorm:"column:total_items;not null" json:"totalItems"`
	TotalPacks      int                       `gorm:"column:total_packs;not null" json:"totalPacks"`
	ConfigurationID uint                      `gorm:"column:configuration_id;not null" json:"configurationId"`
	Configuration   packcfg.PackConfiguration `gorm:"foreignKey:ConfigurationID" json:"-"`
	Timestamp       time.Time                 `gorm:"column:timestamp;not null;default:CURRENT_TIMESTAMP" json:"timestamp"`
}

// PackResult represents a single pack in the result
type PackResult struct {
	Size     int `json:"size"`
	Quantity int `json:"quantity"`
}

// CalculateAPIRequest represents an API request to calculate pack_configurations for an order
type CalculateAPIRequest struct {
	OrderQuantity int `json:"orderQuantity"`
}

// CalculateAPIResponse represents an API response for a calculation request
type CalculateAPIResponse struct {
	OrderQuantity int          `json:"orderQuantity"`
	TotalItems    int          `json:"totalItems"`
	TotalPacks    int          `json:"totalPacks"`
	Packs         []PackResult `json:"pack_configurations"`
	Success       bool         `json:"success"`
	ErrorMessage  string       `json:"errorMessage,omitempty"`
}
