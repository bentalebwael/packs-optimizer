package pack_configurations

import (
	"github.com/lib/pq"
)

// PackConfiguration represents a pack configuration entity in the database
type PackConfiguration struct {
	ID        uint          `gorm:"column:id;primarykey;autoIncrement" json:"id"`
	PackSizes pq.Int64Array `gorm:"column:pack_sizes;type:int[];not null" json:"packSizes"`
	Signature string        `gorm:"column:signature;uniqueIndex" json:"signature"`
	Active    bool          `gorm:"column:active;default:false" json:"active"`
}

// PackCfgAPIRequest represents an API request to update pack sizes
type PackCfgAPIRequest struct {
	PackSizes []int `json:"packSizes"`
}

// PackCfgAPIResponse represents an API response for getting pack sizes
type PackCfgAPIResponse struct {
	PackSizes []int `json:"packSizes"`
}
