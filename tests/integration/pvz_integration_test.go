package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type PVZRequest struct {
	City string `json:"city"`
}

type PVZResponse struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type ReceptionRequest struct {
	PvzId string `json:"pvzId"`
}

type ReceptionResponse struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzId    string    `json:"pvzId"`
	Status   string    `json:"status"`
}

type ProductRequest struct {
	Type  string `json:"type"`
	PvzId string `json:"pvzId"`
}

type ProductResponse struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionId string    `json:"receptionId"`
}

const (
	DefaultBaseURL = "http://api-test:8080"
	ModeratorRole  = "moderator"
	EmployeeRole   = "employee"
	CityMoscow     = "Москва"
	ProductTypes   = 3
)

func getBaseURL() string {
	if url := os.Getenv("TEST_API_URL"); url != "" {
		return url
	}
	return DefaultBaseURL
}

func makeRequest(t *testing.T, method, url string, body interface{}, token string) ([]byte, int) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err, "Failed to create request")

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Request failed")
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	return respBody, resp.StatusCode
}

func getToken(t *testing.T, role string) string {
	baseURL := getBaseURL()
	reqBody := DummyLoginRequest{Role: role}
	respBody, statusCode := makeRequest(t, "POST", baseURL+"/dummyLogin", reqBody, "")
	require.Equal(t, http.StatusOK, statusCode, "Failed to get token")

	var token string
	err := json.Unmarshal(respBody, &token)
	require.NoError(t, err, "Failed to unmarshal token")
	require.NotEmpty(t, token, "Token is empty")

	return token
}

func TestPVZWorkflow(t *testing.T) {
	baseURL := getBaseURL()
	fmt.Printf("Using API URL: %s\n", baseURL)

	moderatorToken := getToken(t, ModeratorRole)
	require.NotEmpty(t, moderatorToken, "Moderator token is empty")

	pvzReq := PVZRequest{City: CityMoscow}
	pvzRespBody, statusCode := makeRequest(t, "POST", baseURL+"/pvz", pvzReq, moderatorToken)
	require.Equal(t, http.StatusCreated, statusCode, "Failed to create PVZ")

	var pvzResp PVZResponse
	err := json.Unmarshal(pvzRespBody, &pvzResp)
	require.NoError(t, err, "Failed to unmarshal PVZ response")
	require.NotEmpty(t, pvzResp.ID, "PVZ ID is empty")
	require.Equal(t, CityMoscow, pvzResp.City, "PVZ city is incorrect")

	pvzId := pvzResp.ID
	fmt.Printf("Created PVZ with ID: %s\n", pvzId)

	employeeToken := getToken(t, EmployeeRole)
	require.NotEmpty(t, employeeToken, "Employee token is empty")

	receptionReq := ReceptionRequest{PvzId: pvzId}
	receptionRespBody, statusCode := makeRequest(t, "POST", baseURL+"/receptions", receptionReq, employeeToken)
	require.Equal(t, http.StatusCreated, statusCode, "Failed to create reception")

	var receptionResp ReceptionResponse
	err = json.Unmarshal(receptionRespBody, &receptionResp)
	require.NoError(t, err, "Failed to unmarshal reception response")
	require.NotEmpty(t, receptionResp.ID, "Reception ID is empty")
	require.Equal(t, pvzId, receptionResp.PvzId, "Reception PVZ ID is incorrect")
	require.Equal(t, "in_progress", receptionResp.Status, "Reception status is not in_progress")

	fmt.Printf("Created reception with ID: %s\n", receptionResp.ID)

	productTypes := []string{"электроника", "одежда", "обувь"}
	for i := 0; i < 50; i++ {

		productType := productTypes[i%len(productTypes)]

		productReq := ProductRequest{
			Type:  productType,
			PvzId: pvzId,
		}

		productRespBody, statusCode := makeRequest(t, "POST", baseURL+"/products", productReq, employeeToken)
		require.Equal(t, http.StatusCreated, statusCode, fmt.Sprintf("Failed to add product #%d", i+1))

		var productResp ProductResponse
		err = json.Unmarshal(productRespBody, &productResp)
		require.NoError(t, err, "Failed to unmarshal product response")
		require.NotEmpty(t, productResp.ID, "Product ID is empty")
		require.Equal(t, productType, productResp.Type, "Product type is incorrect")

		if (i+1)%10 == 0 {
			fmt.Printf("Added %d products\n", i+1)
		}
	}

	fmt.Println("Added all 50 products")

	_, statusCode = makeRequest(t, "POST", fmt.Sprintf("%s/pvz/%s/close_last_reception", baseURL, pvzId), nil, employeeToken)
	require.Equal(t, http.StatusOK, statusCode, "Failed to close reception")

	fmt.Println("Reception closed successfully")
	fmt.Println("Integration test completed successfully")
}
