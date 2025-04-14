package handlers

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/dto"
	"avito-backend-trainee-assignment-spring-2025/internal/api/handlers/mocks"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/apperrors"
	"avito-backend-trainee-assignment-spring-2025/internal/domain/models"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/repoerrors"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_addProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	pvzID := uuid.New()
	productID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success add product",
			requestBody: map[string]interface{}{
				"type":  "электроника",
				"pvzId": pvzID.String(),
			},
			setupMocks: func() {
				mockProductService.EXPECT().
					AddProduct(gomock.Any(), "электроника", pvzID).
					Return(&models.Product{
						ID:          productID,
						DateTime:    now,
						Type:        "электроника",
						ReceptionID: receptionID,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":          productID.String(),
				"dateTime":    now.Format(time.RFC3339Nano),
				"type":        "электроника",
				"receptionId": receptionID.String(),
			},
		},
		{
			name: "No active reception",
			requestBody: map[string]interface{}{
				"type":  "электроника",
				"pvzId": pvzID.String(),
			},
			setupMocks: func() {
				mockProductService.EXPECT().
					AddProduct(gomock.Any(), "электроника", pvzID).
					Return(nil, apperrors.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "No active reception for this pickup point.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(requestBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set(string(userRoleKey), "employee")

			handler.addProduct(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, tt.expectedBody["id"], responseBody["id"])
				assert.Equal(t, tt.expectedBody["type"], responseBody["type"])
				assert.Equal(t, tt.expectedBody["receptionId"], responseBody["receptionId"])
				assert.NotEmpty(t, responseBody["dateTime"])
			} else {
				assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
			}
		})
	}
}

func TestHandler_closeReception(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		pvzIDParam     string
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Success close reception",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockReceptionService.EXPECT().
					CloseReception(gomock.Any(), pvzID).
					Return(&models.Reception{
						ID:       receptionID,
						DateTime: now,
						PVZID:    pvzID,
						Status:   models.ReceptionStatusClosed,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":       receptionID.String(),
				"dateTime": now.Format(time.RFC3339Nano),
				"pvzId":    pvzID.String(),
				"status":   "close",
			},
		},
		{
			name:       "Reception already closed",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockReceptionService.EXPECT().
					CloseReception(gomock.Any(), pvzID).
					Return(nil, apperrors.ErrReceptionAlreadyClosed)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "This reception is already closed.",
			},
		},
		{
			name:       "No active reception",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockReceptionService.EXPECT().
					CloseReception(gomock.Any(), pvzID).
					Return(nil, apperrors.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "No active reception for this pickup point.",
			},
		},
		{
			name:           "Invalid PVZ ID",
			pvzIDParam:     "invalid-uuid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "Invalid PVZ ID format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req, _ := http.NewRequest(http.MethodPost, "/pvz/"+tt.pvzIDParam+"/close_last_reception", nil)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set(string(userRoleKey), "employee")
			c.Params = gin.Params{{Key: "pvzId", Value: tt.pvzIDParam}}

			handler.closeReception(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, tt.expectedBody["id"], responseBody["id"])
				assert.Equal(t, tt.expectedBody["pvzId"], responseBody["pvzId"])
				assert.Equal(t, tt.expectedBody["status"], responseBody["status"])
				assert.NotEmpty(t, responseBody["dateTime"])
			} else {
				assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
			}
		})
	}
}

func TestHandler_createPVZ(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	pvzID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success create PVZ in Moscow",
			requestBody: map[string]interface{}{
				"city": "Москва",
			},
			setupMocks: func() {
				mockPVZService.EXPECT().
					CreatePVZ(gomock.Any(), "Москва").
					Return(&models.PVZ{
						ID:               pvzID,
						RegistrationDate: now,
						City:             "Москва",
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":               pvzID.String(),
				"registrationDate": now.Format(time.RFC3339Nano),
				"city":             "Москва",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(requestBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set(string(userRoleKey), "moderator")

			handler.createPVZ(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, tt.expectedBody["id"], responseBody["id"])
				assert.Equal(t, tt.expectedBody["city"], responseBody["city"])
				assert.NotEmpty(t, responseBody["registrationDate"])
			} else {
				assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
			}
		})
	}
}

func TestHandler_createReception(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	pvzID := uuid.New()
	receptionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success create reception",
			requestBody: map[string]interface{}{
				"pvzId": pvzID.String(),
			},
			setupMocks: func() {
				mockReceptionService.EXPECT().
					CreateReception(gomock.Any(), pvzID).
					Return(&models.Reception{
						ID:       receptionID,
						DateTime: now,
						PVZID:    pvzID,
						Status:   models.ReceptionStatusInProgress,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":       receptionID.String(),
				"dateTime": now.Format(time.RFC3339Nano),
				"pvzId":    pvzID.String(),
				"status":   "in_progress",
			},
		},
		{
			name: "Active reception exists",
			requestBody: map[string]interface{}{
				"pvzId": pvzID.String(),
			},
			setupMocks: func() {
				mockReceptionService.EXPECT().
					CreateReception(gomock.Any(), pvzID).
					Return(nil, apperrors.ErrActiveReceptionExists)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "Cannot create a new reception while the previous one is not closed.",
			},
		},
		{
			name: "PVZ not found",
			requestBody: map[string]interface{}{
				"pvzId": pvzID.String(),
			},
			setupMocks: func() {
				mockReceptionService.EXPECT().
					CreateReception(gomock.Any(), pvzID).
					Return(nil, repoerrors.ErrPVZNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"message": "Pickup point not found.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(requestBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set(string(userRoleKey), "employee")

			handler.createReception(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, tt.expectedBody["id"], responseBody["id"])
				assert.Equal(t, tt.expectedBody["pvzId"], responseBody["pvzId"])
				assert.Equal(t, tt.expectedBody["status"], responseBody["status"])
				assert.NotEmpty(t, responseBody["dateTime"])
			} else {
				assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
			}
		})
	}
}

func TestHandler_deleteLastProduct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	pvzID := uuid.New()

	tests := []struct {
		name           string
		pvzIDParam     string
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Success delete last product",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), pvzID).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Last product deleted successfully",
			},
		},
		{
			name:       "No products to delete",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), pvzID).
					Return(apperrors.ErrNoProductsToDelete)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "No products to delete in the current reception.",
			},
		},
		{
			name:       "Reception is closed",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), pvzID).
					Return(apperrors.ErrReceptionCannotBeModified)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "Closed reception cannot be modified.",
			},
		},
		{
			name:       "No active reception",
			pvzIDParam: pvzID.String(),
			setupMocks: func() {
				mockProductService.EXPECT().
					DeleteLastProduct(gomock.Any(), pvzID).
					Return(apperrors.ErrNoActiveReception)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "No active reception for this pickup point.",
			},
		},
		{
			name:           "Invalid PVZ ID",
			pvzIDParam:     "invalid-uuid",
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"message": "Invalid PVZ ID format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req, _ := http.NewRequest(http.MethodPost, "/pvz/"+tt.pvzIDParam+"/delete_last_product", nil)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set(string(userRoleKey), "employee")
			c.Params = gin.Params{{Key: "pvzId", Value: tt.pvzIDParam}}

			handler.deleteLastProduct(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
		})
	}
}

func TestHandler_dummyLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedToken  bool
	}{
		{
			name: "Success login as employee",
			requestBody: map[string]interface{}{
				"role": "employee",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					DummyLogin("employee").
					Return("dummy-token-employee", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  true,
		},
		{
			name: "Success login as moderator",
			requestBody: map[string]interface{}{
				"role": "moderator",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					DummyLogin("moderator").
					Return("dummy-token-moderator", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(requestBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req

			handler.dummyLogin(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedToken {
				assert.NotEmpty(t, resp.Body.String())
			} else {
				var responseBody map[string]interface{}
				json.Unmarshal(resp.Body.Bytes(), &responseBody)
				assert.Contains(t, responseBody, "message")
			}
		})
	}
}

func TestHandler_getPVZList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()
	now := time.Now()

	pvzWithReceptions := []models.PVZWithReceptions{
		{
			PVZ: &models.PVZ{
				ID:               pvzID,
				RegistrationDate: now,
				City:             "Москва",
			},
			Receptions: []*models.Reception{
				{
					ID:       receptionID,
					DateTime: now,
					PVZID:    pvzID,
					Status:   models.ReceptionStatusClosed,
					Products: []models.Product{
						{
							ID:          productID,
							DateTime:    now,
							Type:        "электроника",
							ReceptionID: receptionID,
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMocks     func()
		expectedStatus int
		expectedItems  int
	}{
		{
			name:        "Success get PVZ list",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockPVZService.EXPECT().
					GetAllPVZWithReceptions(gomock.Any(), gomock.Any()).
					Return(pvzWithReceptions, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedItems:  1,
		},
		{
			name:        "Empty PVZ list",
			queryParams: "?page=1&limit=10",
			setupMocks: func() {
				mockPVZService.EXPECT().
					GetAllPVZWithReceptions(gomock.Any(), gomock.Any()).
					Return([]models.PVZWithReceptions{}, 0, nil)
			},
			expectedStatus: http.StatusOK,
			expectedItems:  0,
		},
		{
			name:        "Filtered by date",
			queryParams: "?startDate=2023-01-01T00:00:00Z&endDate=2023-12-31T23:59:59Z",
			setupMocks: func() {
				mockPVZService.EXPECT().
					GetAllPVZWithReceptions(gomock.Any(), gomock.Any()).
					Return(pvzWithReceptions, 1, nil)
			},
			expectedStatus: http.StatusOK,
			expectedItems:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req, _ := http.NewRequest(http.MethodGet, "/pvz"+tt.queryParams, nil)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req
			c.Set(string(userRoleKey), "employee")

			handler.getPVZList(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody dto.PVZListResponseDTO
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			assert.Equal(t, tt.expectedItems, len(responseBody.Items))
		})
	}
}

func TestHandler_login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedToken  bool
	}{
		{
			name: "Success login",
			requestBody: map[string]interface{}{
				"email":    "user@example.com",
				"password": "password123",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Login(gomock.Any(), "user@example.com", "password123").
					Return("test-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedToken:  true,
		},
		{
			name: "Invalid credentials",
			requestBody: map[string]interface{}{
				"email":    "user@example.com",
				"password": "wrong-password",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Login(gomock.Any(), "user@example.com", "wrong-password").
					Return("", apperrors.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedToken:  false,
		},
		{
			name: "User not found",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Login(gomock.Any(), "nonexistent@example.com", "password123").
					Return("", repoerrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedToken:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(requestBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req

			handler.login(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			if tt.expectedToken {
				assert.NotEmpty(t, resp.Body.String())
			} else {
				var responseBody map[string]interface{}
				json.Unmarshal(resp.Body.Bytes(), &responseBody)
				assert.Contains(t, responseBody, "message")
			}
		})
	}
}

func TestHandler_register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func()
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success register employee",
			requestBody: map[string]interface{}{
				"email":    "employee@example.com",
				"password": "password123",
				"role":     "employee",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Register(gomock.Any(), "employee@example.com", "password123", "employee").
					Return(&models.User{
						ID:    userID,
						Email: "employee@example.com",
						Role:  models.RoleEmployee,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":    userID.String(),
				"email": "employee@example.com",
				"role":  "employee",
			},
		},
		{
			name: "Success register moderator",
			requestBody: map[string]interface{}{
				"email":    "moderator@example.com",
				"password": "password123",
				"role":     "moderator",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Register(gomock.Any(), "moderator@example.com", "password123", "moderator").
					Return(&models.User{
						ID:    userID,
						Email: "moderator@example.com",
						Role:  models.RoleModerator,
					}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"id":    userID.String(),
				"email": "moderator@example.com",
				"role":  "moderator",
			},
		},
		{
			name: "User already exists",
			requestBody: map[string]interface{}{
				"email":    "existing@example.com",
				"password": "password123",
				"role":     "employee",
			},
			setupMocks: func() {
				mockUserService.EXPECT().
					Register(gomock.Any(), "existing@example.com", "password123", "employee").
					Return(nil, repoerrors.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"message": "User with this email already exists.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			requestBodyBytes, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(requestBodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(resp)
			c.Request = req

			handler.register(c)

			assert.Equal(t, tt.expectedStatus, resp.Code)

			var responseBody map[string]interface{}
			json.Unmarshal(resp.Body.Bytes(), &responseBody)

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, tt.expectedBody["id"], responseBody["id"])
				assert.Equal(t, tt.expectedBody["email"], responseBody["email"])
				assert.Equal(t, tt.expectedBody["role"], responseBody["role"])
			} else {
				assert.Equal(t, tt.expectedBody["message"], responseBody["message"])
			}
		})
	}
}

func Test_mapPVZListToDTO(t *testing.T) {
	now := time.Now()
	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	pvzWithReceptions := []models.PVZWithReceptions{
		{
			PVZ: &models.PVZ{
				ID:               pvzID,
				RegistrationDate: now,
				City:             "Москва",
			},
			Receptions: []*models.Reception{
				{
					ID:       receptionID,
					DateTime: now,
					PVZID:    pvzID,
					Status:   models.ReceptionStatusClosed,
					Products: []models.Product{
						{
							ID:          productID,
							DateTime:    now,
							Type:        "электроника",
							ReceptionID: receptionID,
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		pvzList  []models.PVZWithReceptions
		total    int
		page     int
		limit    int
		expected dto.PVZListResponseDTO
	}{
		{
			name:    "Map PVZ list with receptions and products",
			pvzList: pvzWithReceptions,
			total:   1,
			page:    1,
			limit:   10,
			expected: dto.PVZListResponseDTO{
				Items: []dto.PVZWithReceptionsResponseDTO{
					{
						PVZ: dto.PVZ{
							Id:               &pvzID,
							RegistrationDate: &now,
							City:             dto.PVZCity("Москва"),
						},
						Receptions: []dto.ReceptionWithProductsDTO{
							{
								Reception: dto.Reception{
									Id:       &receptionID,
									DateTime: now,
									PvzId:    pvzID,
									Status:   dto.ReceptionStatus(models.ReceptionStatusClosed),
								},
								Products: []dto.Product{
									{
										Id:          &productID,
										DateTime:    &now,
										Type:        dto.ProductType("электроника"),
										ReceptionId: receptionID,
									},
								},
							},
						},
					},
				},
				TotalCount: 1,
				Page:       1,
				Limit:      10,
			},
		},
		{
			name:    "Empty PVZ list",
			pvzList: []models.PVZWithReceptions{},
			total:   0,
			page:    1,
			limit:   10,
			expected: dto.PVZListResponseDTO{
				Items:      []dto.PVZWithReceptionsResponseDTO{},
				TotalCount: 0,
				Page:       1,
				Limit:      10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapPVZListToDTO(tt.pvzList, tt.total, tt.page, tt.limit)

			assert.Equal(t, tt.expected.TotalCount, result.TotalCount)
			assert.Equal(t, tt.expected.Page, result.Page)
			assert.Equal(t, tt.expected.Limit, result.Limit)
			assert.Equal(t, len(tt.expected.Items), len(result.Items))

			if len(tt.expected.Items) > 0 {
				assert.Equal(t, tt.expected.Items[0].PVZ.Id.String(), result.Items[0].PVZ.Id.String())
				assert.Equal(t, tt.expected.Items[0].PVZ.City, result.Items[0].PVZ.City)

				if len(tt.expected.Items[0].Receptions) > 0 {
					assert.Equal(t, tt.expected.Items[0].Receptions[0].Reception.Id.String(),
						result.Items[0].Receptions[0].Reception.Id.String())
					assert.Equal(t, tt.expected.Items[0].Receptions[0].Reception.Status,
						result.Items[0].Receptions[0].Reception.Status)

					if len(tt.expected.Items[0].Receptions[0].Products) > 0 {
						assert.Equal(t, tt.expected.Items[0].Receptions[0].Products[0].Id.String(),
							result.Items[0].Receptions[0].Products[0].Id.String())
						assert.Equal(t, tt.expected.Items[0].Receptions[0].Products[0].Type,
							result.Items[0].Receptions[0].Products[0].Type)
					}
				}
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid token",
			authHeader:     "Bearer valid-token",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header is required",
		},
		{
			name:           "Invalid auth format",
			authHeader:     "InvalidFormat token123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization format",
		},
		{
			name:           "Bearer without token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			c.Request = req

			middleware := handler.authMiddleware()

			middleware(c)
			c.Next()

			if tt.expectedStatus == http.StatusOK {
			} else {
				assert.True(t, c.IsAborted())
				assert.Equal(t, tt.expectedStatus, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedError, response["message"])
			}
		})
	}
}

func TestRoleMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	tests := []struct {
		name           string
		role           string
		requiredRole   string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Correct role employee",
			role:           "employee",
			requiredRole:   "employee",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Correct role moderator",
			role:           "moderator",
			requiredRole:   "moderator",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name:           "Employee trying to access moderator route",
			role:           "employee",
			requiredRole:   "moderator",
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
		{
			name:           "Moderator trying to access employee route",
			role:           "moderator",
			requiredRole:   "employee",
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
		{
			name:           "Missing role",
			role:           "",
			requiredRole:   "employee",
			expectedStatus: http.StatusForbidden,
			expectedError:  "User role not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			c.Request = req

			if tt.role != "" {
				c.Set(string(userRoleKey), tt.role)
			}

			middleware := handler.roleMiddleware(tt.requiredRole)

			middleware(c)

			if tt.expectedStatus == http.StatusOK {
				assert.False(t, c.IsAborted())
			} else {
				assert.True(t, c.IsAborted())
				assert.Equal(t, tt.expectedStatus, w.Code)

				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedError, response["message"])
			}
		})
	}
}

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserService := mocks.NewMockUserServiceInterface(ctrl)
	mockPVZService := mocks.NewMockPVZServiceInterface(ctrl)
	mockReceptionService := mocks.NewMockReceptionServiceInterface(ctrl)
	mockProductService := mocks.NewMockProductServiceInterface(ctrl)

	testConfig := &config.Config{}

	handler := NewHandler(
		mockUserService,
		mockPVZService,
		mockReceptionService,
		mockProductService,
		testConfig,
	)

	tests := []struct {
		name          string
		setupRequest  func() (*httptest.ResponseRecorder, *gin.Context)
		handlerStatus int
	}{
		{
			name: "Success response",
			setupRequest: func() (*httptest.ResponseRecorder, *gin.Context) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				c.Request = req
				return w, c
			},
			handlerStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := tt.setupRequest()

			middleware := handler.metricsMiddleware()

			middleware(c)

			c.Status(tt.handlerStatus)

			assert.Equal(t, tt.handlerStatus, w.Code)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "Repository error - PVZ not found",
			err:          repoerrors.ErrPVZNotFound,
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Pickup point not found.",
		},
		{
			name:         "App error - Invalid city",
			err:          apperrors.ErrInvalidCity,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Pickup points can only be created in the following cities: Moscow, Saint Petersburg, Kazan.",
		},
		{
			name:         "App error - Active reception exists",
			err:          apperrors.ErrActiveReceptionExists,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Cannot create a new reception while the previous one is not closed.",
		},
		{
			name:         "Unknown error",
			err:          errors.New("unknown error"),
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "unknown error",
		},
		{
			name:         "Wrapped error",
			err:          fmt.Errorf("database error: %w", repoerrors.ErrUserNotFound),
			expectedCode: http.StatusNotFound,
			expectedMsg:  "User not found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg := getErrorResponse(tt.err)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}
