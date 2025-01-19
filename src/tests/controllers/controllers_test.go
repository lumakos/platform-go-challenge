package controllers_test

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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var FavoritesStore sync.Map

func TestGetFavorites(t *testing.T) {
	userID := "user123"
	// Adding a sample asset to the favorites store
	asset := models.Asset{
		ID:          uuid.New().String(),
		Type:        models.Chart,
		Title:       "Chart Title",
		Description: "Chart Description",
		Data:        []byte(`{"x_axis": "Time", "y_axis": "Value", "data": [1, 2, 3]}`),
	}
	FavoritesStore.Store(userID, []models.Asset{asset})

	// Creating request to get favorites
	req, err := http.NewRequest("GET", "/v1/favorites/"+userID+"?page=1&limit=10", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(controllers.GetFavorites)
	handler.ServeHTTP(rr, req)

	// Assert the response status code is 200 OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Assert the response contains the asset ID
	var assets []models.Asset
	err = json.Unmarshal(rr.Body.Bytes(), &assets)
	assert.NoError(t, err)
	assert.Len(t, assets, 1)
	assert.Equal(t, asset.ID, assets[0].ID)
}

func TestAddFavorite(t *testing.T) {
	router := routes.RegisterFavoriteRoutes()
	asset := models.Asset{
		Type:        models.Chart,
		Title:       "Test Chart",
		Description: "Test Description",
	}
	body, _ := json.Marshal(asset)
	req, _ := http.NewRequest("POST", "/v1/favorites/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}
}

// func TestRemoveFavorite(t *testing.T) {
// 	router := routes.SetupRoutes()
// 	asset := models.Asset{
// 		ID:          "testAssetID",
// 		Type:        models.Chart,
// 		Title:       "Test Chart",
// 		Description: "Test Description",
// 	}
// 	controllers.FavoritesStore.Store("testUser", []models.Asset{asset})
// 	req, _ := http.NewRequest("DELETE", "/v1/favorites/testUser/testAssetID", nil)
// 	rr := httptest.NewRecorder()
// 	router.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
// 	}
// }

// func TestEditFavorite(t *testing.T) {
// 	router := routes.SetupRoutes()
// 	asset := models.Asset{
// 		ID:          "testAssetID",
// 		Type:        models.Chart,
// 		Title:       "Test Chart",
// 		Description: "Old Description",
// 	}
// 	controllers.FavoritesStore.Store("testUser", []models.Asset{asset})
// 	updatedAsset := models.Asset{
// 		Description: "Updated Description",
// 	}
// 	body, _ := json.Marshal(updatedAsset)
// 	req, _ := http.NewRequest("PUT", "/v1/favorites/testUser/testAssetID", bytes.NewBuffer(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	rr := httptest.NewRecorder()
// 	router.ServeHTTP(rr, req)

// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
// 	}
// }
