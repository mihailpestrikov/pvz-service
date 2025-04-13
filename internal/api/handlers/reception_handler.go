package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) createReception(c *gin.Context) {
	var req dto.PostReceptionsJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in createReception")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	pvzID, err := uuid.Parse(req.PvzId.String())
	if err != nil {
		log.Debug().Err(err).Str("pvz_id", req.PvzId.String()).Msg("Invalid PVZ ID format")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid PVZ ID format"})
		return
	}

	reception, err := h.receptionService.CreateReception(c.Request.Context(), pvzID)
	if err != nil {
		log.Error().Err(err).Str("pvz_id", pvzID.String()).Msg("Reception creation failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := dto.Reception{
		Id:       &reception.ID,
		DateTime: reception.DateTime,
		PvzId:    reception.PVZID,
		Status:   dto.ReceptionStatus(reception.Status),
	}

	log.Info().
		Str("reception_id", reception.ID.String()).
		Str("pvz_id", reception.PVZID.String()).
		Msg("Reception created successfully")

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) closeReception(c *gin.Context) {
	pvzIdParam := c.Param("pvzId")
	pvzID, err := uuid.Parse(pvzIdParam)
	if err != nil {
		log.Debug().Err(err).Str("pvz_id", pvzIdParam).Msg("Invalid PVZ ID format")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid PVZ ID format"})
		return
	}

	reception, err := h.receptionService.CloseReception(c.Request.Context(), pvzID)
	if err != nil {
		log.Error().Err(err).Str("pvz_id", pvzID.String()).Msg("Reception closing failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := dto.Reception{
		Id:       &reception.ID,
		DateTime: reception.DateTime,
		PvzId:    reception.PVZID,
		Status:   dto.ReceptionStatus(reception.Status),
	}

	log.Info().
		Str("reception_id", reception.ID.String()).
		Str("pvz_id", reception.PVZID.String()).
		Msg("Reception closed successfully")

	c.JSON(http.StatusOK, response)
}
