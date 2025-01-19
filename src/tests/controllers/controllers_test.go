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
	// Prepare mock data
	userID := "1"
	assets := []models.Asset{
		{ID: 1, Type: models.Chart, Description: "Chart 1"},
		{ID: 2, Type: models.Insight, Description: "Insight 2"},
		{ID: 3, Type: models.Chart, Description: "Chart 3"},
		{ID: 4, Type: models.Insight, Description: "Insight 4"},
		{ID: 5, Type: models.Chart, Description: "Chart 5"},
		{ID: 6, Type: models.Insight, Description: "Insight 6"},
	}
	controllers.UserStore.Store(userID, assets)

	// Test valid pagination - page 1, limit 2
	req, err := http.NewRequest("GET", "/api/v1/users/"+userID+"/favorites?page=1&limit=2", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := routes.RegisterFavoriteRoutes()
	router.ServeHTTP(rr, req)

	// Check the response code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the structure of the response
	var response map[string]interface{}
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the 'data' is an array of assets
	data, ok := response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)

	// Validate pagination metadata
	meta, ok := response["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 6.0, meta["total_count"])
	assert.Equal(t, 3.0, meta["total_pages"])
	assert.Equal(t, 1.0, meta["current_page"])
	assert.Equal(t, 2.0, meta["per_page"])

	// Test valid pagination - page 2, limit 2
	req, _ = http.NewRequest("GET", "/api/v1/users/"+userID+"/favorites?page=2&limit=2", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the structure of the response
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	data, ok = response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)

	// Validate pagination metadata for page 2
	meta, ok = response["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 6.0, meta["total_count"])
	assert.Equal(t, 3.0, meta["total_pages"])
	assert.Equal(t, 2.0, meta["current_page"])
	assert.Equal(t, 2.0, meta["per_page"])

	// Test page number greater than total pages
	req, _ = http.NewRequest("GET", "/api/v1/users/"+userID+"/favorites?page=4&limit=2", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the structure of the response
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the 'data' is an empty array
	data, ok = response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 0)

	// Validate pagination metadata for an out-of-range page
	meta, ok = response["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 6.0, meta["total_count"])
	assert.Equal(t, 3.0, meta["total_pages"])
	assert.Equal(t, 4.0, meta["current_page"])
	assert.Equal(t, 2.0, meta["per_page"])
}

func TestAddFavorite(t *testing.T) {
	// New empty map
	controllers.UserStore = sync.Map{}

	// Valid chart data
	userID := "1"
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

	body, _ := json.Marshal(asset)
	req, _ := http.NewRequest("POST", "/api/v1/users/"+userID+"/favorites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router := routes.RegisterFavoriteRoutes()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	// Invalid chart data (missing Title)
	invalidChartData := models.ChartData{
		XAxis: "Time",
		YAxis: "Value",
		Data:  []float64{1.1, 2.2, 3.3},
	}
	asset.Data = encodeToJSON(invalidChartData)
	body, _ = json.Marshal(asset)
	req, _ = http.NewRequest("POST", "/api/v1/users/"+userID+"/favorites", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
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

	assert.Equal(t, http.StatusOK, rr.Code)

	value, ok := controllers.UserStore.Load(userID)
	if !ok {
		t.Fatalf("User store not found")
	}

	assets, ok = value.([]models.Asset)
	// Assert that only one asset remains in the user's favorites
	assert.Len(t, assets, 1)
	// Assert that the remaining asset has the correct description
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
