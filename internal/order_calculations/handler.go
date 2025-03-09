package order_calculations

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/pack-calculator/pkg/errors"
)

type Handler struct {
	logger  *zap.Logger
	service Service
}

func NewHandler(logger *zap.Logger, service Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

// CalculatePacksForOrder calculates the optimal pack_configurations for an order
func (h *Handler) CalculatePacksForOrder(c *gin.Context) {
	payload, exists := c.Get("payload")
	if !exists {
		errMsg := "Failed to retrieve payload from context"
		h.logger.Error(errMsg)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError(errMsg))
		return
	}
	request := payload.(*CalculateAPIRequest)

	// Calculate optimal pack_configurations
	packs, totalItems, totalPacks, err := h.service.OrderProcessing(c, request.OrderQuantity)
	if err != nil {
		errMsg := "Failed to process order request"
		h.logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.NewInternalErrorWrap(errMsg, err))
		return
	}

	response := CalculateAPIResponse{
		OrderQuantity: request.OrderQuantity,
		TotalItems:    totalItems,
		TotalPacks:    totalPacks,
		Packs:         packs,
		Success:       true,
	}
	c.JSON(http.StatusOK, response)
}
