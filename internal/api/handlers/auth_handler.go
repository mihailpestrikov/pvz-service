package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/dto"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) dummyLogin(c *gin.Context) {
	var req dto.PostDummyLoginJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in dummyLogin")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	role := string(req.Role)

	token, err := h.userService.DummyLogin(role)
	if err != nil {
		log.Error().Err(err).Str("role", role).Msg("Dummy login failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *Handler) register(c *gin.Context) {
	var req dto.PostRegisterJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in register")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	role := string(req.Role)

	user, err := h.userService.Register(c.Request.Context(), string(req.Email), req.Password, role)
	if err != nil {
		log.Error().Err(err).Str("email", string(req.Email)).Msg("User registration failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := dto.User{
		Id:    &user.ID,
		Email: types.Email(user.Email),
		Role:  dto.UserRole(user.Role),
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) login(c *gin.Context) {
	var req dto.PostLoginJSONRequestBody

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in login")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), string(req.Email), req.Password)
	if err != nil {
		log.Error().Err(err).Str("email", string(req.Email)).Msg("Login failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, token)
}
