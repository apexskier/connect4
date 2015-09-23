package main

import (
	"fmt"
	"github.com/apexskier/connect4/server"
	"log"
	"net/http"
)

func main() {
	router := server.NewRouter()
	router.Handle("/", http.FileServer(http.Dir("client")))
	router.Handle("/main.js", http.FileServer(http.Dir("client")))
	router.Handle("/cookies.js", http.FileServer(http.Dir("client")))

	fmt.Println("Running")

	log.Fatal(http.ListenAndServe(":8000", router))
}
