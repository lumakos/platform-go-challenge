package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"platform-go-challenge/src/models"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var (
	ErrUserNotFound  = errors.New("User not found or no favorites exist")
	ErrAssetNotFound = errors.New("Asset not found")
)

// Rate limiting constants
const (
	rateLimit  = 20          // Max requests per minute
	rateWindow = time.Minute // Time window for rate limiting
)

var rateLimitStore sync.Map // Key: userID (string), Value: []time.Time

// Decodes the JSON payload from the request body into a provided destination object.
func DecodeRequestBody(r *http.Request, dest interface{}) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

// Sends an HTTP error response with a specified status code and error message in JSON format.
func RespondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
	})
}

// Sends an HTTP success response with a specified status code and a JSON payload.
func RespondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}, isDirect bool) {
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
func ValidateContentType(r *http.Request) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("Invalid content type")
	}
	return nil
}

// Extracts and validates pagination parameters from the request.
func ParsePagination(r *http.Request) (int, int) {
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
func PaginateAssets(assets []models.Asset, page, limit int) []models.Asset {
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
func GeneratePaginationMetadata(totalCount, page, limit int) map[string]interface{} {
	totalPages := (totalCount + limit - 1) / limit // Calculate total pages (rounded up)

	return map[string]interface{}{
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"current_page": page,
		"per_page":     limit,
	}
}

// Parses the userID and assetID from the URL path variables in the incoming HTTP request.
func ParseIDs(r *http.Request) (string, uint64, error) {
	userID := mux.Vars(r)["userID"]
	assetIDStr := mux.Vars(r)["assetID"]

	assetID, err := strconv.ParseUint(assetIDStr, 10, 32)
	if err != nil {
		return "", 0, fmt.Errorf("Invalid asset ID")
	}

	return userID, assetID, nil
}

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
