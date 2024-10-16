package main

import (
	"fmt"
	"net/http"
)

func main() {
	serveMux := http.NewServeMux()

	// serve current directory when accessing the root of the server
	serveMux.Handle("/", http.FileServer(http.Dir(".")))

	server := &http.Server{
		Addr:    ":8080",  // Bind to localhost:8080
		Handler: serveMux, // Use the ServeMux as the handler
	}

	// Start the server
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
