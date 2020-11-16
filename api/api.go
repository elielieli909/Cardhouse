package api

import (
	"log"
	"net/http"
)

// StartServer is called in main in order to instantiate the API listener
func StartServer() {
	router := NewRouter()

	log.Fatal(http.ListenAndServe(":8080", router))
}
