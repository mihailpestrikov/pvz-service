package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/auth"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"avito-backend-trainee-assignment-spring-2025/internal/services"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

//go:generate oapi-codegen -config ../../../oapi-codegen.yaml ../../../swagger.yaml

var userFriendlyErrors = map[error]string{
	apperrors.ErrInvalidCredentials:        "Invalid email or password. Please check your credentials.",
	apperrors.ErrEmailRequired:             "Email is required for registration.",
	apperrors.ErrInvalidEmail:              "Invalid email format specified.",
	apperrors.ErrPasswordRequired:          "Password is required for registration.",
	apperrors.ErrInvalidPassword:           "Password must contain at least 6 characters.",
	apperrors.ErrInvalidRole:               "Invalid role specified. Available roles: employee, moderator.",
	apperrors.ErrCityRequired:              "City is required to create a pickup point.",
	apperrors.ErrInvalidCity:               "Pickup points can only be created in the following cities: Moscow, Saint Petersburg, Kazan.",
	apperrors.ErrInvalidPVZID:              "Invalid pickup point ID specified.",
	apperrors.ErrReceptionAlreadyClosed:    "This reception is already closed.",
	apperrors.ErrReceptionCannotBeModified: "Closed reception cannot be modified.",
	apperrors.ErrActiveReceptionExists:     "Cannot create a new reception while the previous one is not closed.",
	apperrors.ErrNoActiveReception:         "No active reception for this pickup point.",
	apperrors.ErrInvalidReceptionID:        "Invalid reception ID specified.",
	apperrors.ErrProductTypeRequired:       "Product type is required.",
	apperrors.ErrInvalidProductType:        "Invalid product type specified. Available types: electronics, clothes, shoes.",
	apperrors.ErrInvalidProductID:          "Invalid product ID specified.",
	apperrors.ErrNoProductsToDelete:        "No products to delete in the current reception.",
	repoerrors.ErrPVZNotFound:              "Pickup point not found.",
	repoerrors.ErrReceptionNotFound:        "Reception not found.",
	repoerrors.ErrProductNotFound:          "Product not found.",
	repoerrors.ErrUserNotFound:             "User not found.",
	repoerrors.ErrUserAlreadyExists:        "User with this email already exists.",
	repoerrors.ErrPVZAlreadyExists:         "Pickup point with this ID already exists.",
}

var errorStatusCodes = map[error]int{
	repoerrors.ErrPVZNotFound:              http.StatusNotFound,
	repoerrors.ErrProductNotFound:          http.StatusNotFound,
	repoerrors.ErrReceptionNotFound:        http.StatusNotFound,
	repoerrors.ErrUserNotFound:             http.StatusNotFound,
	apperrors.ErrInvalidEmail:              http.StatusBadRequest,
	apperrors.ErrInvalidPassword:           http.StatusBadRequest,
	apperrors.ErrInvalidRole:               http.StatusBadRequest,
	apperrors.ErrInvalidCity:               http.StatusBadRequest,
	apperrors.ErrCityRequired:              http.StatusBadRequest,
	apperrors.ErrInvalidProductType:        http.StatusBadRequest,
	apperrors.ErrProductTypeRequired:       http.StatusBadRequest,
	apperrors.ErrActiveReceptionExists:     http.StatusBadRequest,
	apperrors.ErrNoActiveReception:         http.StatusBadRequest,
	apperrors.ErrReceptionAlreadyClosed:    http.StatusBadRequest,
	apperrors.ErrReceptionCannotBeModified: http.StatusBadRequest,
	apperrors.ErrNoProductsToDelete:        http.StatusBadRequest,
	apperrors.ErrInvalidCredentials:        http.StatusUnauthorized,
	repoerrors.ErrUserAlreadyExists:        http.StatusConflict,
	repoerrors.ErrPVZAlreadyExists:         http.StatusConflict,
}

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	userRoleKey contextKey = "user_role"
)

type Handler struct {
	userService      *services.UserService
	pvzService       *services.PVZService
	receptionService *services.ReceptionService
	productService   *services.ProductService
	config           *config.Config
}

func NewHandler(
	userService *services.UserService,
	pvzService *services.PVZService,
	receptionService *services.ReceptionService,
	productService *services.ProductService,
	config *config.Config,
) *Handler {
	return &Handler{
		userService:      userService,
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
		config:           config,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	gin.SetMode(h.config.Server.GinMode)

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(h.metricsMiddleware())

	router.POST("/dummyLogin", h.dummyLogin)
	router.POST("/register", h.register)
	router.POST("/login", h.login)

	authorized := router.Group("/")
	authorized.Use(h.authMiddleware())

	moderatorRoutes := authorized.Group("/")
	moderatorRoutes.Use(h.roleMiddleware("moderator"))
	{
		moderatorRoutes.POST("/pvz", h.createPVZ)
	}

	authorized.GET("/pvz", h.getPVZList)

	employeeRoutes := authorized.Group("/")
	employeeRoutes.Use(h.roleMiddleware("employee"))
	{
		employeeRoutes.POST("/receptions", h.createReception)
		employeeRoutes.POST("/pvz/:pvzId/close_last_reception", h.closeReception)
		employeeRoutes.POST("/products", h.addProduct)
		employeeRoutes.POST("/pvz/:pvzId/delete_last_product", h.deleteLastProduct)
	}

	return router
}

func getUserFriendlyError(err error) string {
	if friendlyMsg, exists := userFriendlyErrors[err]; exists {
		return friendlyMsg
	}

	var unwrappedErr error
	if errors.As(err, &unwrappedErr) {
		if friendlyMsg, exists := userFriendlyErrors[unwrappedErr]; exists {
			return friendlyMsg
		}
	}

	return err.Error()
}

func getErrorResponse(err error) (int, string) {
	statusCode := http.StatusInternalServerError

	if code, exists := errorStatusCodes[err]; exists {
		statusCode = code
	}

	var unwrappedErr error
	if errors.As(err, &unwrappedErr) {
		if code, exists := errorStatusCodes[unwrappedErr]; exists {
			statusCode = code
		}
	}

	message := getUserFriendlyError(err)

	return statusCode, message
}

func (h *Handler) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		metrics.RequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), string(rune(status))).Inc()
		metrics.RequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authorization header is required"})
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization format"})
			return
		}

		claims, err := auth.ValidateToken(bearerToken[1], h.config.JWT.Secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}

		c.Set(string(userIDKey), claims.UserID)
		c.Set(string(userRoleKey), claims.Role)

		c.Next()
	}
}

func (h *Handler) roleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(string(userRoleKey))
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "User role not found"})
			return
		}

		if role.(string) != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}
