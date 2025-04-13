package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/dto"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
)

func (h *Handler) dummyLogin(c *gin.Context) {
	var req dto.DummyLoginRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in dummyLogin")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	token, err := h.userService.DummyLogin(req.Role)
	if err != nil {
		log.Error().Err(err).Str("role", req.Role).Msg("Dummy login failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponseDTO{Token: token})
}

func (h *Handler) register(c *gin.Context) {
	var req dto.RegisterRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in register")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("User registration failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	response := dto.UserResponseDTO{
		ID:    user.ID.String(),
		Email: user.Email,
		Role:  user.Role,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *Handler) login(c *gin.Context) {
	var req dto.LoginRequestDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Debug().Err(err).Msg("Invalid request format in login")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request format"})
		return
	}

	token, err := h.userService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Login failed")

		statusCode, message := getErrorResponse(err)
		c.JSON(statusCode, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponseDTO{Token: token})
}
