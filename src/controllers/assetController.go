package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"platform-go-challenge/src/models"
	"platform-go-challenge/src/validations"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var UserStore sync.Map // Key: userID (string), Value: []Asset

// Rate limiting constants
const (
	rateLimit  = 20          // Max requests per minute
	rateWindow = time.Minute // Time window for rate limiting
)

var rateLimitStore sync.Map // Key: userID (string), Value: []time.Time

// Gets User's Favorites
func GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	// Check for rate limiting
	if !CheckRateLimit(userID) {
		respondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded, try again later")
		return
	}

	page, limit := parsePagination(r)

	assets, err := getUserFavorites(userID)
	if err != nil {
		respondWithSuccess(w, http.StatusNotFound, map[string]interface{}{
			"data": []map[string]interface{}{}, // Empty array if no assets
			"meta": generatePaginationMetadata(0, 0, page, limit),
		}, true)
		return
	}

	// Paginate the assets
	totalCount := len(assets)
	paginatedAssets := paginateAssets(assets, page, limit)

	// Send the response with the correct structure
	respondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"data": paginatedAssets,
		"meta": generatePaginationMetadata(totalCount, len(paginatedAssets), page, limit),
	}, true)
}

// Adds an Asset to User's favorite list
func AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	// Rate limiting check
	if !CheckRateLimit(userID) {
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
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	value, _ := UserStore.LoadOrStore(userID, []models.Asset{})
	assets := value.([]models.Asset)

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

	// Assign the next ID and set the description
	asset.ID = nextID
	asset.Description = description

	// Add the new asset to the list
	assets = append(assets, asset)
	UserStore.Store(userID, assets)

	respondWithSuccess(w, http.StatusCreated, map[string]interface{}{
		"id":          asset.ID,
		"type":        asset.Type,
		"Description": asset.Description,
		"data": map[string]interface{}{
			"text": description,
		},
	}, false)
}

// Retrieves the user's favorite assets from the store.
func getUserFavorites(userID string) ([]models.Asset, error) {
	value, ok := UserStore.Load(userID)
	if !ok {
		return nil, fmt.Errorf("User not found or no favorites exist")
	}
	return value.([]models.Asset), nil
}

// Updates the description of a user's favorite asset
func EditDescription(w http.ResponseWriter, r *http.Request) {
	if err := validateContentType(r); err != nil {
		respondWithError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}

	userID, assetID, err := parseIDs(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if !CheckRateLimit(userID) {
		respondWithError(w, http.StatusTooManyRequests, "Rate limit exceeded, try again later")
		return
	}

	var updatedAsset models.Asset
	if err := decodeRequestBody(r, &updatedAsset); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}

	if err := updateAssetDescription(userID, assetID, updatedAsset.Description.(string)); err != nil {
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, ErrAssetNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	respondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"message": "Asset description updated",
	}, false)
}

// Removes Favorite from User's list
func RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, assetID, err := parseIDs(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	value, ok := UserStore.Load(userID)
	if !ok {
		respondWithError(w, http.StatusNotFound, "User not found or no favorites exist")
		return
	}

	assets := value.([]models.Asset)
	removedAsset, updatedAssets, err := removeAssetByID(assets, uint(assetID))
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	// Store the updated assets list back into UserStore
	UserStore.Store(userID, updatedAssets)

	respondWithSuccess(w, http.StatusOK, map[string]interface{}{
		"id": removedAsset.ID,
		"message": map[string]interface{}{
			"text": "Asset removed from favorites",
		},
	}, false)
}

func removeAssetByID(assets []models.Asset, assetID uint) (models.Asset, []models.Asset, error) {
	// Iterate over the assets slice and find the asset to remove
	for i, asset := range assets {
		if asset.ID == assetID {
			removedAsset := asset
			// Remove the asset from the slice
			updatedAssets := append(assets[:i], assets[i+1:]...)
			// Return the removed asset and the updated assets list
			return removedAsset, updatedAssets, nil
		}
	}
	// If the asset is not found, return an error
	return models.Asset{}, assets, fmt.Errorf("Asset not found")
}

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

func updateAssetDescription(userID string, assetID uint64, newDescription string) error {
	value, ok := UserStore.Load(userID)
	if !ok {
		return ErrUserNotFound
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
		return ErrAssetNotFound
	}

	return nil
}

// Helper Functions

// Rate limit check function
func CheckRateLimit(userID string) bool {
	now := time.Now()
	windowStart := now.Add(-rateWindow)

	value, _ := rateLimitStore.LoadOrStore(userID, []time.Time{})
	timestamps := value.([]time.Time)

	// Remove timestamps outside the window
	var validTimestamps []time.Time
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	if len(validTimestamps) >= rateLimit {
		return false
	}

	// Store the new timestamp
	validTimestamps = append(validTimestamps, now)
	rateLimitStore.Store(userID, validTimestamps)

	return true
}

// Parses the userID and assetID from the URL path variables in the incoming HTTP request.
func parseIDs(r *http.Request) (string, uint64, error) {
	userID := mux.Vars(r)["userID"]
	assetIDStr := mux.Vars(r)["assetID"]

	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("Invalid asset ID")
	}

	return userID, assetID, nil
}

var (
	ErrUserNotFound  = errors.New("User not found or no favorites exist")
	ErrAssetNotFound = errors.New("Asset not found")
)

// Decodes the JSON payload from the request body into a provided destination object.
func decodeRequestBody(r *http.Request, dest interface{}) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

// Sends an HTTP error response with a specified status code and error message in JSON format.
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

// Sends an HTTP success response with a specified status code and a JSON payload.
func respondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}, isDirect bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Check if the response should be wrapped in "data" or not
	if isDirect {
		// Directly encode data (for GetUserFavorites and other special cases)
		json.NewEncoder(w).Encode(data)
	} else {
		// Wrap data in a "data" key (for all other cases)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": data,
		})
	}
}

// Validates that the Content-Type header of the incoming HTTP request is set to application/json.
func validateContentType(r *http.Request) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("Invalid content type")
	}
	return nil
}

// Extracts and validates pagination parameters from the request.
func parsePagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	return page, limit
}

// Handles slicing the assets for pagination based on the page and limit.
func paginateAssets(assets []models.Asset, page, limit int) []models.Asset {
	totalCount := len(assets)
	start := (page - 1) * limit
	end := start + limit

	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	return assets[start:end]
}

// Generates pagination metadata to include in the response.
func generatePaginationMetadata(totalCount, itemCount, page, limit int) map[string]interface{} {
	totalPages := (totalCount + limit - 1) / limit // Calculate total pages (rounded up)

	return map[string]interface{}{
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": page,
		"per_page":     limit,
	}
}
