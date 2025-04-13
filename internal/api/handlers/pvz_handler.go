package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/dto"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) createPVZ(c *gin.Context) {
	var req dto.PostPvzJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in createPVZ")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	pvz, err := h.pvzService.CreatePVZ(c.Request.Context(), string(req.City))
	if err != nil {
		log.Error().Err(err).Str("city", string(req.City)).Msg("PVZ creation failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := dto.PVZ{
		Id:               &pvz.ID,
		RegistrationDate: &pvz.RegistrationDate,
		City:             dto.PVZCity(pvz.City),
	}

	log.Info().
		Str("pvz_id", pvz.ID.String()).
		Str("city", pvz.City).
		Msg("PVZ created successfully")

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) getPVZList(c *gin.Context) {
	var filterDTO dto.GetPvzParams

	if err := c.ShouldBindQuery(&filterDTO); err != nil {
		log.Debug().Err(err).Msg("Invalid query parameters in getPVZList")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid query parameters"})
		return
	}

	filter := models.PVZFilter{
		StartDate: filterDTO.StartDate,
		EndDate:   filterDTO.EndDate,
		Page:      1,
		Limit:     10,
	}

	if filterDTO.Page != nil && *filterDTO.Page > 0 {
		filter.Page = *filterDTO.Page
	}

	if filterDTO.Limit != nil && *filterDTO.Limit > 0 && *filterDTO.Limit <= 30 {
		filter.Limit = *filterDTO.Limit
	}

	pvzList, total, err := h.pvzService.GetAllPVZWithReceptions(c.Request.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get PVZ list")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := mapPVZListToDTO(pvzList, total, filter.Page, filter.Limit)

	log.Info().
		Int("total_count", total).
		Int("returned_count", len(pvzList)).
		Int("page", filter.Page).
		Int("limit", filter.Limit).
		Msg("Retrieved PVZ list successfully")

	c.JSON(http.StatusOK, response)
}

func mapPVZListToDTO(pvzList []models.PVZWithReceptions, total, page, limit int) dto.PVZListResponseDTO {
	items := make([]dto.PVZWithReceptionsResponseDTO, len(pvzList))

	for i, pvz := range pvzList {
		receptions := make([]dto.ReceptionWithProductsDTO, 0)

		for _, reception := range pvz.Receptions {
			products := make([]dto.Product, len(reception.Products))

			for j, product := range reception.Products {
				products[j] = dto.Product{
					DateTime:    &product.DateTime,
					Id:          &product.ID,
					Type:        dto.ProductType(product.Type),
					ReceptionId: product.ReceptionID,
				}
			}

			receptions = append(receptions, dto.ReceptionWithProductsDTO{
				Reception: dto.Reception{
					Id:       &reception.ID,
					DateTime: reception.DateTime,
					PvzId:    reception.PVZID,
					Status:   dto.ReceptionStatus(reception.Status),
				},
				Products: products,
			})
		}

		items[i] = dto.PVZWithReceptionsResponseDTO{
			PVZ: dto.PVZ{
				Id:               &pvz.PVZ.ID,
				RegistrationDate: &pvz.PVZ.RegistrationDate,
				City:             dto.PVZCity(pvz.PVZ.City),
			},
			Receptions: receptions,
		}
	}

	return dto.PVZListResponseDTO{
		Items:      items,
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}
}
