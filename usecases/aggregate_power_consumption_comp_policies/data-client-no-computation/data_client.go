package main

import (
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	// This handler cannot compute any results

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()

	// Register the composite handler at '/' on port 3001
	http.Handle("/", middleware.PrivacyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
