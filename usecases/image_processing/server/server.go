package main

import (
	"fmt"
	imageProcessing "github.com/JacobMoxham/PartIIProjectImplementation/image_recognition"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"log"
	"net/http"
)

func imageProcessingHandler(w http.ResponseWriter, r *http.Request) {
	top5Labels := imageProcessing.GetTop5LabelsFromImageReader(r.Body)
	returnString := ""
	for _, l := range top5Labels {
		returnString += fmt.Sprintf("label: %s, probability: %.2f%%\n", l.Label, l.Probability*100)
	}

	_, err := w.Write([]byte(returnString))
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}
}

func main() {
	// Create actual function to run
	finalHandler := http.HandlerFunc(imageProcessingHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, finalHandler)

	// Chain together "other" middlewares
	handler := middleware.PolicyAwareHandler(computationPolicy)

	// Register the composite handler at '/' on port 3000
	http.Handle("/", handler)
	log.Println("Listening...")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
