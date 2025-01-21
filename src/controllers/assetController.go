package controllers

import (
	"encoding/json"
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

// Rate limit check function
func CheckRateLimit(userID string) bool {
	now := time.Now()
	windowStart := now.Add(-rateWindow)

	// Load or create the rate limit timestamp list for the user
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

// Gets User's Favorites
func GetUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	// Rate limiting check
	if !CheckRateLimit(userID) {
		http.Error(w, "Rate limit exceeded, try again later", http.StatusTooManyRequests)
		return
	}

	// Prepare pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	value, ok := UserStore.Load(userID)
	if !ok {
		response := map[string]interface{}{
			"data": []models.Asset{},
			"meta": map[string]interface{}{
				"total_count":  0,
				"total_pages":  0,
				"current_page": page,
				"per_page":     limit,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	assets := value.([]models.Asset)
	totalCount := len(assets)                      // Total number of items in the store
	totalPages := (totalCount + limit - 1) / limit // Calculate total pages (rounded up)
	start := (page - 1) * limit
	end := start + limit
	if start > totalCount {
		start = totalCount
	}
	if end > totalCount {
		end = totalCount
	}

	paginatedAssets := assets[start:end]

	// Prepare response with pagination metadata
	response := map[string]interface{}{
		"data": paginatedAssets,
		"meta": map[string]interface{}{
			"total_count":  totalCount,
			"total_pages":  totalPages,
			"current_page": page,
			"per_page":     limit,
		},
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	var description string
	switch asset.Type {
	case models.Chart:
		var chartData models.ChartData
		if err := json.Unmarshal(asset.Data, &chartData); err != nil {
			http.Error(w, "Invalid chart data", http.StatusBadRequest)
			return
		}
		description = models.RetrieveDescription(chartData)
	case models.Insight:
		var insightData models.InsightData
		if err := json.Unmarshal(asset.Data, &insightData); err != nil {
			http.Error(w, "Invalid insight data", http.StatusBadRequest)
			return
		}
		description = models.RetrieveDescription(insightData)
	case models.Audience:
		var audienceData models.AudienceData
		if err := json.Unmarshal(asset.Data, &audienceData); err != nil {
			http.Error(w, "Invalid audience data", http.StatusBadRequest)
			return
		}
		description = models.RetrieveDescription(audienceData)
	default:
		http.Error(w, "Unsupported asset type", http.StatusBadRequest)
		return
	}

	value, _ := UserStore.LoadOrStore(userID, []models.Asset{})
	assets := value.([]models.Asset)

	asset.ID = uint(len(assets) + 1)
	asset.Description = description

	assets = append(assets, asset)
	UserStore.Store(userID, assets)

	// Create response
	response := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"id":          asset.ID,
				"type":        asset.Type,
				"Description": asset.Description,
				"data": map[string]interface{}{
					"text": description,
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Removes Favorite from User's list
func RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]
	assetIDStr := mux.Vars(r)["assetID"]

	// Convert assetID from string to uint
	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid asset ID", http.StatusBadRequest)
		return
	}

	value, ok := UserStore.Load(userID)
	if !ok {
		http.Error(w, "User not found or no favorites exist", http.StatusNotFound)
		return
	}

	assets := value.([]models.Asset)

	// Search for and remove the asset
	found := false
	var removedAsset models.Asset
	for i, asset := range assets {
		if asset.ID == uint(assetID) {
			removedAsset = asset
			assets = append(assets[:i], assets[i+1:]...)
			UserStore.Store(userID, assets)
			found = true
			break
		}
	}

	if found {
		// Prepare response
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id": removedAsset.ID,
					"message": map[string]interface{}{
						"text": "Asset removed from favorites",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Asset not found", http.StatusNotFound)
	}
}

// EditDescription updates the description of a user's favorite asset
func EditDescription(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	userID := mux.Vars(r)["userID"]
	assetIDStr := mux.Vars(r)["assetID"]

	// Convert assetID to uint
	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid asset ID", http.StatusBadRequest)
		return
	}

	// Rate limiting check
	if !CheckRateLimit(userID) {
		http.Error(w, "Rate limit exceeded, try again later", http.StatusTooManyRequests)
		return
	}

	// Decode the request body to get the updated description
	var updatedAsset models.Asset
	if err := json.NewDecoder(r.Body).Decode(&updatedAsset); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	value, ok := UserStore.Load(userID)
	if !ok {
		http.Error(w, "User not found or no favorites exist", http.StatusNotFound)
		return
	}

	assets := value.([]models.Asset)

	// Locate and update the asset description
	found := false
	var updatedAssetDetails models.Asset
	for i, asset := range assets {
		if asset.ID == uint(assetID) {
			assets[i].Description = updatedAsset.Description
			updatedAssetDetails = assets[i]
			UserStore.Store(userID, assets)
			found = true
			break
		}
	}

	if found {
		// Prepare response
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          updatedAssetDetails.ID,
					"type":        updatedAssetDetails.Type,
					"Description": updatedAssetDetails.Description,
					"message": map[string]interface{}{
						"text": "Asset description updated",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Asset not found", http.StatusNotFound)
	}
}
