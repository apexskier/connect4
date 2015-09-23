package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Running")

	log.Fatal(http.ListenAndServe(":7000", http.FileServer(http.Dir("client"))))
}
