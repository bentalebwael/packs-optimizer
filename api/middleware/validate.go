package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pack-calculator/internal/order_calculations"
	"github.com/pack-calculator/internal/pack_configurations"
	"github.com/pack-calculator/pkg/errors"
)

// ValidateOrder validates the order calculation input
func ValidateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request order_calculations.CalculateAPIRequest

		// Decode JSON body
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, errors.NewValidationErrorWrap("Invalid JSON format", err))
			c.Abort()
			return
		}

		// Validate orderQuantity is positive
		if request.OrderQuantity <= 0 {
			c.JSON(http.StatusBadRequest, errors.NewValidationError("Order quantity must be a positive integer"))
			c.Abort()
			return
		}

		// Set orderCalc in context
		c.Set("payload", &request)

		// Continue to next handler if validation passes
		c.Next()
	}
}

// ValidatePacks validates the pack configuration input
func ValidatePacks() gin.HandlerFunc {
	return func(c *gin.Context) {
		var request pack_configurations.PackCfgAPIRequest

		// Decode JSON body
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, errors.NewValidationErrorWrap("Invalid JSON format", err))
			c.Abort()
			return
		}

		// Validate packSizes is not empty
		if len(request.PackSizes) == 0 {
			c.JSON(http.StatusBadRequest, errors.NewValidationError("Pack sizes cannot be empty"))
			c.Abort()
			return
		}

		// Validate all numbers are positive
		for _, size := range request.PackSizes {
			if size <= 0 {
				c.JSON(http.StatusBadRequest, errors.NewValidationError("Pack sizes must be positive integers"))
				c.Abort()
				return
			}
		}

		// Check for duplicates
		seen := make(map[int]bool)
		for _, size := range request.PackSizes {
			if seen[size] {
				c.JSON(http.StatusBadRequest, errors.NewValidationError("Pack sizes must not contain duplicates"))
				c.Abort()
				return
			}
			seen[size] = true
		}

		// Set packCfg in context
		c.Set("payload", &request)

		// Continue to next handler if validation passes
		c.Next()
	}
}
