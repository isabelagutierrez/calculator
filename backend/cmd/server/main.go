package main

import (
	"log"
	"net/http"

	"github.com/isabelagutierrez/calculator/backend/internal/api"
)

func main() {
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", api.NewHandler()))
}
