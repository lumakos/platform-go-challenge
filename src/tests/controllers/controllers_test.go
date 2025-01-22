package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"platform-go-challenge/src/controllers"
	"platform-go-challenge/src/models"
	"platform-go-challenge/src/routes"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserFavorites(t *testing.T) {
	controllers.UserStore = sync.Map{}

	// Mock data
	userID := "1"
	assets := []models.Asset{
		{ID: 1, Type: models.Chart, Description: "Chart 1"},
		{ID: 2, Type: models.Insight, Description: "Insight 2"},
		{ID: 3, Type: models.Chart, Description: "Chart 3"},
		{ID: 4, Type: models.Insight, Description: "Insight 4"},
	}
	controllers.UserStore.Store(userID, assets)

	// Define test cases
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedData   int // Number of assets expected in the response
		expectedMeta   map[string]float64
	}{
		{
			name:           "Valid pagination - page 1, limit 2",
			url:            "/api/v1/users/1/favorites?page=1&limit=2",
			expectedStatus: http.StatusOK,
			expectedData:   2,
			expectedMeta: map[string]float64{
				"total_count":  4.0,
				"total_pages":  2.0,
				"current_page": 1.0,
				"per_page":     2.0,
			},
		},
		{
			name:           "Page beyond range",
			url:            "/api/v1/users/1/favorites?page=3&limit=2",
			expectedStatus: http.StatusOK,
			expectedData:   0,
			expectedMeta: map[string]float64{
				"total_count":  4.0,
				"total_pages":  2.0,
				"current_page": 3.0,
				"per_page":     2.0,
			},
		},
		{
			name:           "User not found",
			url:            "/api/v1/users/999/favorites?page=1&limit=2",
			expectedStatus: http.StatusNotFound,
			expectedData:   0,
			expectedMeta: map[string]float64{
				"total_count":  0.0,
				"total_pages":  0.0,
				"current_page": 1.0,
				"per_page":     2.0,
			},
		},
	}

	router := routes.RegisterFavoriteRoutes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			// Verify the response status
			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response map[string]interface{}
			err := json.NewDecoder(rr.Body).Decode(&response)
			assert.NoError(t, err)

			// Check the number of assets
			data, _ := response["data"].([]interface{})
			assert.Len(t, data, tt.expectedData)

			// Check pagination metadata
			meta, ok := response["meta"].(map[string]interface{})
			assert.True(t, ok)
			for key, value := range tt.expectedMeta {
				assert.Equal(t, value, meta[key].(float64))
			}
		})
	}
}

func TestAddFavorite(t *testing.T) {
	// Reset UserStore
	controllers.UserStore = sync.Map{}

	// Define user ID
	userID := "1"

	// Helper to add an asset and return the response
	addFavorite := func(asset models.Asset) *httptest.ResponseRecorder {
		body, _ := json.Marshal(asset)
		req, _ := http.NewRequest("POST", "/api/v1/users/"+userID+"/favorites", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		router := routes.RegisterFavoriteRoutes()
		router.ServeHTTP(rr, req)
		return rr
	}

	// Case 1: Add a valid chart asset
	chartData := models.ChartData{
		Title: "Chart Title",
		XAxis: "Time",
		YAxis: "Value",
		Data:  []float64{1.1, 2.2, 3.3},
	}
	asset := models.Asset{
		Type: models.Chart,
		Data: encodeToJSON(chartData),
	}
	rr := addFavorite(asset)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify the response contains ID = 1
	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should unmarshal without errors")

	// Extract 'data' and verify the 'id'
	data, dataExists := response["data"].(map[string]interface{})
	assert.True(t, dataExists, "Response should contain 'data' field as a map")
	id, idExists := data["id"].(float64)
	assert.True(t, idExists, "Data field should contain 'id'")
	assert.Equal(t, float64(1), id)

	// Case 2: Add another valid chart asset (ID should be 2)
	chartData.Title = "Second Chart"
	asset.Data = encodeToJSON(chartData)
	rr = addFavorite(asset)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify the response contains ID = 2
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should unmarshal without errors")
	data, dataExists = response["data"].(map[string]interface{})
	assert.True(t, dataExists, "Response should contain 'data' field as a map")
	id, idExists = data["id"].(float64)
	assert.True(t, idExists, "Data field should contain 'id'")
	assert.Equal(t, float64(2), id)
}

func TestRemoveFavorite(t *testing.T) {
	// Initialize an empty sync.Map for UserStore
	controllers.UserStore = sync.Map{}

	// Mock data
	userID := "1"
	assets := []models.Asset{
		{ID: 1, Type: models.Chart, Description: "Chart 1"},
		{ID: 2, Type: models.Insight, Description: "Insight 2"},
	}
	controllers.UserStore.Store(userID, assets)

	// Create a DELETE request to remove the asset with ID 1
	req, _ := http.NewRequest("DELETE", "/api/v1/users/"+userID+"/favorites/1", nil)
	rr := httptest.NewRecorder()
	router := routes.RegisterFavoriteRoutes()
	router.ServeHTTP(rr, req)

	// Assert the status code returned from the server
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify the asset removal in UserStore
	value, ok := controllers.UserStore.Load(userID)
	if !ok {
		t.Fatalf("User store not found")
	}

	// Verify that the asset list only contains one item (the one with ID 2)
	assets, ok = value.([]models.Asset)
	assert.True(t, ok, "Expected assets to be of type []models.Asset")
	assert.Len(t, assets, 1) // Only one asset should remain
	assert.Equal(t, "Insight 2", assets[0].Description)
}

func TestEditDescription(t *testing.T) {
	// Initialize an empty sync.Map for UserStore
	controllers.UserStore = sync.Map{}

	// Mock data
	userID := "1"
	assets := []models.Asset{
		{ID: 1, Type: models.Audience, Description: "Old Description"},
	}
	controllers.UserStore.Store(userID, assets)

	// Create the updated asset with a new description
	updatedAsset := models.Asset{Description: "New Description"}
	body, _ := json.Marshal(updatedAsset) // Convert the updated asset to JSON

	req, _ := http.NewRequest("PUT", "/api/v1/users/"+userID+"/favorites/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json") // Set the content type header

	rr := httptest.NewRecorder()
	router := routes.RegisterFavoriteRoutes() // Register the routes
	router.ServeHTTP(rr, req)                 // Serve the HTTP request

	// Assert that the response status code is OK (200)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Retrieve the user's favorite assets after the update
	value, _ := controllers.UserStore.Load(userID)
	assets, _ = value.([]models.Asset)
	// Assert that the asset's description was updated correctly
	assert.Equal(t, "New Description", assets[0].Description)

	// Test case for empty description
	updatedAsset.Description = ""
	body, _ = json.Marshal(updatedAsset) // Update the description to an empty string
	req, _ = http.NewRequest("PUT", "/api/v1/users/"+userID+"/favorites/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert that the response status code is OK (200) when the description is empty
	assert.Equal(t, http.StatusOK, rr.Code)
}

func encodeToJSON(data interface{}) json.RawMessage {
	bytes, _ := json.Marshal(data)
	return bytes
}
