package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) addProduct(c *gin.Context) {
	var req dto.ProductCreateRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in addProduct")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	pvzID, err := uuid.Parse(req.PVZID)
	if err != nil {
		log.Debug().Err(err).Str("pvz_id", req.PVZID).Msg("Invalid PVZ ID format")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid PVZ ID format"})
		return
	}

	product, err := h.productService.AddProduct(c.Request.Context(), req.Type, pvzID)
	if err != nil {
		log.Error().Err(err).
			Str("type", req.Type).
			Str("pvz_id", pvzID.String()).
			Msg("Product addition failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := dto.ProductResponseDTO{
		ID:          product.ID.String(),
		DateTime:    product.DateTime,
		Type:        product.Type,
		ReceptionID: product.ReceptionID.String(),
	}

	log.Info().
		Str("product_id", product.ID.String()).
		Str("type", product.Type).
		Str("reception_id", product.ReceptionID.String()).
		Msg("Product added successfully")

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) deleteLastProduct(c *gin.Context) {
	pvzID, err := uuid.Parse(c.Param("pvzId"))
	if err != nil {
		log.Debug().Err(err).Str("pvz_id", c.Param("pvzId")).Msg("Invalid PVZ ID format")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid PVZ ID format"})
		return
	}

	err = h.productService.DeleteLastProduct(c.Request.Context(), pvzID)
	if err != nil {
		log.Error().Err(err).Str("pvz_id", pvzID.String()).Msg("Product deletion failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	log.Info().
		Str("pvz_id", pvzID.String()).
		Msg("Last product deleted successfully")

	c.JSON(http.StatusOK, gin.H{"message": "Последний товар успешно удален"})
}
