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

	// Test case where !ok (userID not found in UserStore)
	nonExistentUserID := "999"
	req, _ = http.NewRequest("GET", "/api/v1/users/"+nonExistentUserID+"/favorites?page=1&limit=2", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Check the structure of the response for non-existent user
	err = json.NewDecoder(rr.Body).Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the 'data' is an empty array
	data, ok = response["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 0)

	// Validate pagination metadata for non-existent user
	meta, ok = response["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 0.0, meta["total_count"])
	assert.Equal(t, 0.0, meta["total_pages"])
	assert.Equal(t, 1.0, meta["current_page"])
	assert.Equal(t, 2.0, meta["per_page"])
}

func TestAddFavorite(t *testing.T) {
	// Reset UserStore
	controllers.UserStore = sync.Map{}
	userID := "1"

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
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["data"].([]interface{})[0].(map[string]interface{})["id"])

	// Case 2: Add another valid chart asset (ID should be 2)
	chartData.Title = "Second Chart"
	asset.Data = encodeToJSON(chartData)
	rr = addFavorite(asset)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify the response contains ID = 2
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, float64(2), response["data"].([]interface{})[0].(map[string]interface{})["id"])

	// Case 3: Simulate deletion of asset with ID = 1
	value, _ := controllers.UserStore.Load(userID)
	assets := value.([]models.Asset)
	controllers.UserStore.Store(userID, assets[1:]) // Remove the first asset

	// Case 4: Add a new asset (ID should be 1, filling the gap)
	chartData.Title = "Third Chart"
	asset.Data = encodeToJSON(chartData)
	rr = addFavorite(asset)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify the response contains ID = 1
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, float64(1), response["data"].([]interface{})[0].(map[string]interface{})["id"])

	// Case 5: Add another asset (ID should be 3, continuing the sequence)
	chartData.Title = "Fourth Chart"
	asset.Data = encodeToJSON(chartData)
	rr = addFavorite(asset)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Verify the response contains ID = 3
	_ = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, float64(3), response["data"].([]interface{})[0].(map[string]interface{})["id"])

	// Case 6: Invalid chart data (missing Title)
	invalidChartData := models.ChartData{
		XAxis: "Time",
		YAxis: "Value",
		Data:  []float64{1.1, 2.2, 3.3},
	}
	asset.Data = encodeToJSON(invalidChartData)
	rr = addFavorite(asset)
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
