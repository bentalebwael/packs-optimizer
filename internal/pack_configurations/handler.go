package pack_configurations

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/pack-calculator/pkg/errors"
	"github.com/pack-calculator/pkg/postgres"
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

// GetActivePackConfiguration returns the current pack configuration
func (h *Handler) GetActivePackConfiguration(c *gin.Context) {
	packCfg, err := h.service.GetActive(c.Request.Context())
	if err != nil {
		errMsg := "Failed to retrieve pack configuration"
		h.logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.NewInternalErrorWrap(errMsg, err))
		return
	}

	response := PackCfgAPIResponse{
		PackSizes: postgres.Int64ArrayToIntSlice(packCfg.PackSizes),
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreatePackConfiguration(c *gin.Context) {
	payload, exists := c.Get("payload")
	if !exists {
		errMsg := "Failed to retrieve payload from context"
		h.logger.Error(errMsg)
		c.JSON(http.StatusInternalServerError, errors.NewInternalError(errMsg))
		return
	}

	packCfg := payload.(*PackCfgAPIRequest)
	newPackConfiguration := &PackConfiguration{
		PackSizes: postgres.IntSliceToPqArray(packCfg.PackSizes),
	}

	err := h.service.Create(c.Request.Context(), newPackConfiguration)
	if err != nil {
		errMsg := "Failed to create pack configuration"
		h.logger.Error(errMsg, zap.Error(err))
		c.JSON(http.StatusInternalServerError, errors.NewInternalErrorWrap(errMsg, err))
		return
	}

	response := PackCfgAPIResponse{
		PackSizes: packCfg.PackSizes,
	}
	c.JSON(http.StatusOK, response)
}
