// package routes_test

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"platform-go-challenge/src/controllers"
// 	"platform-go-challenge/src/routes"
// 	"testing"

// 	"github.com/go-playground/assert/v2"
// )

// func TestSetupRoutes(t *testing.T) {
// 	r := routes.RegisterFavoriteRoutes()

// 	// Define test cases
// 	tests := []struct {
// 		method       string
// 		url          string
// 		expectedCode int
// 		handler      http.HandlerFunc
// 	}{
// 		{
// 			method:       http.MethodGet,
// 			url:          "/v1/favorites/1", // Replace userID with an example
// 			expectedCode: http.StatusOK,     // assuming GetFavorites responds with 200
// 			handler:      controllers.GetUserFavorites,
// 		},
// 		// {
// 		// 	method:       http.MethodPost,
// 		// 	url:          "/v1/favorites/1",
// 		// 	expectedCode: http.StatusCreated, // assuming AddFavorite responds with 201
// 		// 	handler:      controllers.AddFavorite,
// 		// },
// 		// {
// 		// 	method:       http.MethodDelete,
// 		// 	url:          "/v1/favorites/1/assetID", // Replace assetID with an example
// 		// 	expectedCode: http.StatusOK,                  // assuming RemoveFavorite responds with 200
// 		// 	handler:      controllers.RemoveFavorite,
// 		// },
// 		// {
// 		// 	method:       http.MethodPut,
// 		// 	url:          "/v1/favorites/1/assetID", // Replace assetID with an example
// 		// 	expectedCode: http.StatusOK,                  // assuming EditFavorite responds with 200
// 		// 	handler:      controllers.EditFavorite,
// 		// },
// 	}

// 	// Run tests
// 	for _, tt := range tests {
// 		t.Run(tt.method+" "+tt.url, func(t *testing.T) {
// 			req, err := http.NewRequest(tt.method, tt.url, nil)
// 			if err != nil {
// 				t.Fatalf("Error creating request: %v", err)
// 			}

// 			// Create a response recorder to capture the response
// 			rr := httptest.NewRecorder()

// 			// Serve the request using the routes handler
// 			r.ServeHTTP(rr, req)

// 			// Check if the response status code is as expected
// 			assert.Equal(t, tt.expectedCode, rr.Code)
// 		})
// 	}
// }
