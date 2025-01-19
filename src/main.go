package main

import (
	"fmt"
	"log"
	"net/http"
	"platform-go-challenge/src/routes"
)

func main() {
	r := routes.RegisterFavoriteRoutes()

	fmt.Println("Server started at :8088")
	log.Fatal(http.ListenAndServe(":8088", r))
}
