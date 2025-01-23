package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"platform-go-challenge/src/helpers"
	"platform-go-challenge/src/models"
	"platform-go-challenge/src/validations"
	"sync"

	"github.com/gorilla/mux"
)

var UserStore sync.Map // Key: userID (string), Value: []Asset

// Gets User's Favorites
func GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	// Check for rate limiting
	if !helpers.CheckRateLimit(userID) {
		helpers.RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded, try again later")
		return
	}

	page, limit := helpers.ParsePagination(r)

	assets, err := getUserFavorites(userID)
	if err != nil {
		helpers.RespondWithSuccess(w, http.StatusNotFound, map[string]interface{}{
			"data": []map[string]interface{}{}, // Empty array if no assets
			"meta": helpers.GeneratePaginationMetadata(0, page, limit),
		}, true)
		return
	}

	// Paginate the assets
	totalCount := len(assets)
	paginatedAssets := helpers.PaginateAssets(assets, page, limit)

	// Send the response with the correct structure
	helpers.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"data": paginatedAssets,
		"meta": helpers.GeneratePaginationMetadata(totalCount, page, limit),
	}, true)
}

// Adds an Asset to User's favorite list
func AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	// Rate limiting check
	if !helpers.CheckRateLimit(userID) {
		http.Error(w, "Rate limit exceeded, try again later", http.StatusTooManyRequests)
		return
	}

	var asset models.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := validations.ValidateAsset(&asset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	description, err := getDescriptionByType(asset.Type, asset.Data)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	value, _ := UserStore.LoadOrStore(userID, []models.Asset{})
	assets := value.([]models.Asset)

	// Calculate the next id
	nextID := calculateNextId(assets)

	// Assign the next ID and set the description
	asset.ID = uint(nextID)
	asset.Description = description

	// Add the new asset to the list
	assets = append(assets, asset)
	UserStore.Store(userID, assets)

	helpers.RespondWithSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":          asset.ID,
		"type":        asset.Type,
		"Description": asset.Description,
		"data": map[string]interface{}{
			"text": description,
		},
	}, false)
}

// Updates the description of a user's favorite asset
func EditDescription(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ValidateContentType(r); err != nil {
		helpers.RespondWithError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}

	userID, assetID, err := helpers.ParseIDs(r)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !helpers.CheckRateLimit(userID) {
		helpers.RespondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded, try again later")
		return
	}

	var updatedAsset models.Asset
	if err := helpers.DecodeRequestBody(r, &updatedAsset); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := updateAssetDescription(userID, assetID, updatedAsset.Description.(string)); err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) || errors.Is(err, helpers.ErrAssetNotFound) {
			helpers.RespondWithError(w, http.StatusNotFound, err.Error())
		} else {
			helpers.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	helpers.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Asset description updated",
	}, false)
}

// Removes Favorite from User's list
func RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, assetID, err := helpers.ParseIDs(r)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	value, ok := UserStore.Load(userID)
	if !ok {
		helpers.RespondWithError(w, http.StatusNotFound, "User not found or no favorites exist")
		return
	}

	assets := value.([]models.Asset)
	removedAsset, updatedAssets, err := removeAssetByID(assets, uint(assetID))
	if err != nil {
		helpers.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	// Store the updated assets list back into UserStore
	UserStore.Store(userID, updatedAssets)

	helpers.RespondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"id": removedAsset.ID,
		"message": map[string]interface{}{
			"text": "Asset removed from favorites",
		},
	}, false)
}

// Calculates the next Asset Id
func calculateNextId(assets []models.Asset) int {
	// Create a set of existing IDs
	existingIDs := make(map[uint]struct{})

	var maxID uint
	for _, a := range assets {
		existingIDs[a.ID] = struct{}{}
		if a.ID > maxID {
			maxID = a.ID
		}
	}

	// Find the smallest missing ID in the range 1 to maxID
	var nextID uint
	for i := uint(1); i <= maxID; i++ {
		if _, exists := existingIDs[i]; !exists {
			nextID = i
			break
		}
	}

	// If no missing ID is found, set the next ID to maxID + 1
	if nextID == 0 {
		nextID = maxID + 1
	}

	return int(nextID)
}

// Retrieves the user's favorite assets from the store.
func getUserFavorites(userID string) ([]models.Asset, error) {
	value, ok := UserStore.Load(userID)
	if !ok {
		return nil, fmt.Errorf("User not found or no favorites exist")
	}
	return value.([]models.Asset), nil
}

// Deletes asset by id
func removeAssetByID(assets []models.Asset, assetID uint) (models.Asset, []models.Asset, error) {
	for i, asset := range assets {
		if asset.ID == assetID {
			removedAsset := asset
			// Remove the asset from the slice
			updatedAssets := append(assets[:i], assets[i+1:]...)
			return removedAsset, updatedAssets, nil
		}
	}

	return models.Asset{}, assets, fmt.Errorf("Asset not found")
}

// Retrieve asset description by type
func getDescriptionByType(assetType models.AssetType, data json.RawMessage) (string, error) {
	switch models.AssetType(assetType) {
	case models.Chart:
		var chartData models.ChartData
		if err := json.Unmarshal(data, &chartData); err != nil {
			return "", fmt.Errorf("Invalid chart data")
		}
		return models.RetrieveDescription(chartData), nil
	case models.Insight:
		var insightData models.InsightData
		if err := json.Unmarshal(data, &insightData); err != nil {
			return "", fmt.Errorf("Invalid insight data")
		}
		return models.RetrieveDescription(insightData), nil
	case models.Audience:
		var audienceData models.AudienceData
		if err := json.Unmarshal(data, &audienceData); err != nil {
			return "", fmt.Errorf("Invalid audience data")
		}
		return models.RetrieveDescription(audienceData), nil
	default:
		return "", fmt.Errorf("Unsupported asset type")
	}
}

// Updates asset's description
func updateAssetDescription(userID string, assetID uint64, newDescription string) error {
	value, ok := UserStore.Load(userID)
	if !ok {
		return helpers.ErrUserNotFound
	}

	assets := value.([]models.Asset)
	found := false
	for i, asset := range assets {
		if asset.ID == uint(assetID) {
			assets[i].Description = newDescription
			UserStore.Store(userID, assets)
			found = true
			break
		}
	}

	if !found {
		return helpers.ErrAssetNotFound
	}

	return nil
}
