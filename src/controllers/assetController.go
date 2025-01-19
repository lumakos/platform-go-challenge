package controllers

import (
	"encoding/json"
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
	rateLimit  = 5           // Max requests per minute
	rateWindow = time.Minute // Time window for rate limiting
)

var rateLimitStore sync.Map // Key: userID (string), Value: []time.Time

//-----------------------------------------//
//              Rate Limiting              //
//-----------------------------------------//

// Rate limit check function
func CheckRateLimit(userID string) bool {
	now := time.Now()
	windowStart := now.Add(-rateWindow)

	value, _ := rateLimitStore.LoadOrStore(userID, []time.Time{})
	timestamps := value.([]time.Time)

	// Filter out timestamps outside the rate window
	var validTimestamps []time.Time
	for _, ts := range timestamps {
		if ts.After(windowStart) {
			validTimestamps = append(validTimestamps, ts)
		}
	}

	if len(validTimestamps) >= rateLimit {
		return false
	}

	// Store the current request timestamp
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

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	resultChan := make(chan []models.Asset)

	go func() {
		value, ok := UserStore.Load(userID)
		if !ok {
			resultChan <- []models.Asset{}
			return
		}

		assets := value.([]models.Asset)
		start := (page - 1) * limit
		end := start + limit
		if start > len(assets) {
			start = len(assets)
		}
		if end > len(assets) {
			end = len(assets)
		}

		resultChan <- assets[start:end]
	}()

	assets := <-resultChan
	json.NewEncoder(w).Encode(assets)
	w.WriteHeader(http.StatusOK)
}

func AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userID"]

	var asset models.Asset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := validations.ValidateAsset(&asset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Determine the description dynamically based on asset type
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

	done := make(chan bool)
	go func() {
		value, _ := UserStore.LoadOrStore(userID, []models.Asset{})
		assets := value.([]models.Asset)

		asset.ID = uint(len(assets) + 1)
		asset.Description = description

		assets = append(assets, asset)
		UserStore.Store(userID, assets)
		done <- true
	}()
	<-done

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Asset added to favorites")
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

	// Retrieve user's favorite assets
	value, ok := UserStore.Load(userID)
	if !ok {
		http.Error(w, "User not found or no favorites exist", http.StatusNotFound)
		return
	}

	assets, ok := value.([]models.Asset)
	if !ok {
		http.Error(w, "Invalid data format in store", http.StatusInternalServerError)
		return
	}

	// Search for the asset and remove it
	found := false
	for i, asset := range assets {
		if asset.ID == uint(assetID) {
			// Remove the asset from the slice
			assets = append(assets[:i], assets[i+1:]...)
			UserStore.Store(userID, assets)
			found = true
			break
		}
	}

	// Respond based on whether the asset was found and removed
	if found {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Asset removed from favorites")
	} else {
		http.Error(w, "Asset not found", http.StatusNotFound)
	}
}

// Edits Favorite's description
func EditDescription(w http.ResponseWriter, r *http.Request) {
	// Validate Content-Type
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	// Extract userID and assetID from URL parameters
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

	// Load the user's favorite assets
	value, ok := UserStore.Load(userID)
	if !ok {
		http.Error(w, "User not found or no favorites exist", http.StatusNotFound)
		return
	}

	assets, ok := value.([]models.Asset)
	if !ok {
		http.Error(w, "Invalid data format in store", http.StatusInternalServerError)
		return
	}

	// Locate the specific asset and update its description
	found := false
	for i, asset := range assets {
		if asset.ID == uint(assetID) {
			assets[i].Description = updatedAsset.Description
			UserStore.Store(userID, assets)
			found = true
			break
		}
	}

	// Respond based on whether the asset was found and updated
	if found {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Asset description updated")
	} else {
		http.Error(w, "Asset not found", http.StatusNotFound)
	}
}
