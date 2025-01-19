package routes

import (
	"fmt"
	"platform-go-challenge/src/controllers"

	"github.com/gorilla/mux"
)

func RegisterFavoriteRoutes() *mux.Router {
	// var muxBase = "/api/v1/favorites"

	// r := mux.NewRouter().StrictSlash(true)
	// r.HandleFunc(fmt.Sprintf("%s/{userID}", muxBase), controllers.GetUserFavorites).Methods("GET")
	// r.HandleFunc(fmt.Sprintf("%s/{userID}", muxBase), controllers.AddFavorite).Methods("POST")
	// r.HandleFunc(fmt.Sprintf("%s/{userID}/{assetID}", muxBase), controllers.RemoveFavorite).Methods("DELETE")
	// r.HandleFunc(fmt.Sprintf("%s/{userID}/{assetID}", muxBase), controllers.EditDescription).Methods("PUT")

	var muxBase = "/api/v1/users"

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc(fmt.Sprintf("%s/{userID}/favorites", muxBase), controllers.GetUserFavorites).Methods("GET")
	r.HandleFunc(fmt.Sprintf("%s/{userID}/favorites", muxBase), controllers.AddFavorite).Methods("POST")
	r.HandleFunc(fmt.Sprintf("%s/{userID}/favorites/{assetID}", muxBase), controllers.RemoveFavorite).Methods("DELETE")
	r.HandleFunc(fmt.Sprintf("%s/{userID}/favorites/{assetID}", muxBase), controllers.EditDescription).Methods("PUT")

	return r
}
